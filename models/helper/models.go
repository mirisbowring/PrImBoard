package helper

import (
	"github.com/mirisbowring/primboard/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GroupMediaHelper is used in some api helper functions
type GroupMediaHelper struct {
	Medias   []models.Media
	Groups   []models.UserGroup
	MediaIDs []primitive.ObjectID
	GroupIDs []primitive.ObjectID
}
