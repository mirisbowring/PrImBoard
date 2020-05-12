package primboard

import (
	"net/http"

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
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// select new token
	i = Invite{ID: result.InsertedID.(primitive.ObjectID)}
	if err = i.FindID(a.DB); err != nil {
		// error occured during select
		RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// could select user from mongo
	RespondWithJSON(w, http.StatusOK, i)
}
