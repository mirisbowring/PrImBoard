package helper

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"os"

	th "github.com/bakape/thumbnailer"
	"github.com/oliamb/cutter"
)

//Thumbnail creates a thumbnail from the passed file reader
func Thumbnail(file *os.File, thumbSize uint) (io.Reader, *th.Source) {
	rs := io.ReadSeeker(file)
	var src th.Source
	var thumb image.Image
	//get original dimensions
	ctx, _ := th.NewFFContext(rs)
	src.Dims, _ = ctx.Dims()
	ctx.Close()
	//calc dimension to fit smalles side to ThumbSize
	opts := th.Options{
		ThumbDims: calcRatio(src.Dims, thumbSize),
	}
	//thumbnail image
	src, thumb, _ = th.Process(rs, opts)
	//crop image to centered square
	thumb, _ = cutter.Crop(thumb, cutter.Config{
		Width:   1,
		Height:  1,
		Mode:    cutter.Centered,
		Options: cutter.Ratio, // Copy is useless here
	})
	//Encode Image with compression
	var opt = jpeg.Options{
		Quality: 80,
	}
	//write thumb into buffer
	buff := new(bytes.Buffer)
	err := jpeg.Encode(buff, thumb, &opt)
	if err != nil {
		panic(err)
	}
	//return buffer as reader
	return bytes.NewReader(buff.Bytes()), &src
}

func calcRatio(dims th.Dims, thumbSize uint) th.Dims {
	if dims.Width == dims.Height {
		return dims
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
