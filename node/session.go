package node

import (
	iModels "github.com/mirisbowring/primboard/internal/models"
	"github.com/mirisbowring/primboard/models"
	log "github.com/sirupsen/logrus"
)

// AddSession appends a new user session to the node (verfies that session is
// valid) - Will override session if already exist
//
// 0 -> ok || 1 -> username is empty || 2 -> token ist empty
func (n *AppNode) AddSession(username string, token string) (int, string) {
	switch "" {
	case username:
		msg := "username cannot be empty to create a session"
		log.Error(msg)
		return 1, msg
	case token:
		msg := "token cannot be empty to create a session"
		log.Error(msg)
		return 2, msg
	default:
		// check if session for user exist already
		tmp := n.Sessions[:0]
		for _, session := range n.Sessions {
			// match found ?
			if session.User.Username == username {
				// refresh token
				session.Token = token
				return 0, ""
			}
			// collect if session is expired
			if session.IsValid() {
				tmp = append(tmp, session)
			}
		}
		// invalidate invalid sessions
		for i := len(tmp); i < len(n.Sessions); i++ {
			n.Sessions[i] = nil // or the zero value of T
		}
		// clear session slice
		n.Sessions = iModels.RemoveEmptySessions(n.Sessions)
		// append new session if did not exist before
		n.Sessions = append(n.Sessions, &iModels.Session{
			User:  models.User{Username: username},
			Token: token,
		})
		return 0, ""
	}
}

// ClearSessions removes the session from tmp store and removes the link
func ClearSessions(sessions []*iModels.Session) {

}
