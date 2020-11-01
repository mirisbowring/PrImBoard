package primboard

import (
	"net/http"

	_http "github.com/mirisbowring/PrImBoard/helper/http"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateInvite handles the webrequest for creating a new invite
func (a *App) GenerateInvite(w http.ResponseWriter, r *http.Request) {
	// create model by passed username
	i := Invite{}
	// try to select model
	result, err := i.Init(a.DB, a.Config.InviteValidity)
	if err != nil {
		// another error occured
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// select new token
	i = Invite{ID: result.InsertedID.(primitive.ObjectID)}
	if err = i.FindID(a.DB); err != nil {
		// error occured during select
		_http.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// could select user from mongo
	_http.RespondWithJSON(w, http.StatusOK, i)
}
