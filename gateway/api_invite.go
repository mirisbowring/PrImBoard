package gateway

import (
	"net/http"

	_http "github.com/mirisbowring/primboard/helper/http"
	"github.com/mirisbowring/primboard/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateInvite handles the webrequest for creating a new invite
func (g *AppGateway) GenerateInvite(w http.ResponseWriter, r *http.Request) {
	// create model by passed username
	i := models.Invite{}
	// try to select model
	result, err := i.Init(g.DB, g.Config.InviteValidity)
	if err != nil {
		// another error occured
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// select new token
	i = models.Invite{ID: result.InsertedID.(primitive.ObjectID)}
	if err = i.FindID(g.DB); err != nil {
		// error occured during select
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// could select user from mongo
	_http.RespondWithJSON(w, http.StatusOK, i)
}
