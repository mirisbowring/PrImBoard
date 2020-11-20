package models

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaQuery holds query options for the media api
type MediaQuery struct {
	After  primitive.ObjectID
	Before primitive.ObjectID
	Event  primitive.ObjectID
	Filter string
	From   primitive.ObjectID
	Until  primitive.ObjectID
	Size   int
	ASC    int16
}

// IsValid validates that the passed query combination is allowed for filtering
func (mq *MediaQuery) IsValid() error {
	if !mq.After.IsZero() && !mq.From.IsZero() {
		return errors.New("query params 'after' and 'from' are not allowed at once")
	} else if !mq.Before.IsZero() && !mq.Until.IsZero() {
		return errors.New("query params 'before' and 'until' are not allowed at once")
	}
	// if asc is anything but asc (1), set it to dsc (default)
	if mq.ASC != 1 {
		mq.ASC = -1
	}
	return nil
}
