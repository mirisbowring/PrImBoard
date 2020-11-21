package helper

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"os"

	th "github.com/bakape/thumbnailer/v2"
	"github.com/oliamb/cutter"
	log "github.com/sirupsen/logrus"
)

//Thumbnail creates a thumbnail from the passed file reader
func Thumbnail(file *os.File, thumbSize uint) (io.Reader, *th.Source) {
	rs := io.ReadSeeker(file)
	var src th.Source
	var thumb image.Image

	logfields := log.Fields{
		"image":     file.Name(),
		"thumbSize": thumbSize,
		"function":  "Thumbnail",
	}

	//create FFContext
	ctx, err := th.NewFFContext(rs)
	if err != nil {
		logfields["error"] = err.Error()
		log.WithFields(logfields).Error("could not create FFContext")
		return nil, nil
	}
	defer ctx.Close()

	// get original Dimensions
	src.Dims, err = ctx.Dims()
	if err != nil {
		logfields["error"] = err.Error()
		log.WithFields(logfields).Error("could not get dimensions of original image")
		return nil, nil
	}

	//calc dimension to fit smalles side to ThumbSize
	opts := th.Options{
		ThumbDims: calcRatio(src.Dims, thumbSize),
	}

	//thumbnail image
	src, thumb, err = th.Process(rs, opts)
	if err != nil {
		logfields["error"] = err.Error()
		log.WithFields(logfields).Error("could not process the thumbnail")
		return nil, nil
	}

	//crop image to centered square
	thumb, err = cutter.Crop(thumb, cutter.Config{
		Width:   1,
		Height:  1,
		Mode:    cutter.Centered,
		Options: cutter.Ratio, // Copy is useless here
	})
	if err != nil {
		logfields["error"] = err.Error()
		log.WithFields(logfields).Error("could not crop thumbnail")
		return nil, nil
	}

	//Encode Image with compression
	var opt = jpeg.Options{
		Quality: 80,
	}

	//write thumb into buffer
	buff := new(bytes.Buffer)
	err = jpeg.Encode(buff, thumb, &opt)
	if err != nil {
		logfields["error"] = err.Error()
		log.WithFields(logfields).Error("could not encode the new thumbnail")
		return nil, nil
	}

	log.WithFields(logfields).Info("created thumbnail")

	//return buffer as reader
	return bytes.NewReader(buff.Bytes()), &src
}

func calcRatio(dims th.Dims, thumbSize uint) th.Dims {
	if dims.Width == dims.Height {
		return th.Dims{
			Width:  thumbSize,
			Height: thumbSize,
		}
	} else if dims.Width > dims.Height {
		tmp := float64(thumbSize) / float64(dims.Height)
		tmp = float64(dims.Width) * tmp
		return th.Dims{
			Width:  uint(tmp),
			Height: thumbSize,
		}
	} else {
		tmp := float64(thumbSize) / float64(dims.Width)
		tmp = float64(dims.Height) * tmp
		return th.Dims{
			Width:  thumbSize,
			Height: uint(tmp),
		}
	}
}
