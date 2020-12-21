package models

import (
	"log"
	"time"

	"github.com/mirisbowring/primboard/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Session stores the user data, the token, the expiration of the session and
// the usergroups of the current user
type Session struct {
	User         models.User
	Token        string
	Expire       time.Time
	Usergroups   []primitive.ObjectID
	NodeTokenMap map[primitive.ObjectID]string
}

// InitUserGroups preselects usergroups for the user (for performance reasons)
func (s *Session) InitUserGroups(db *mongo.Database, user string) {
	groups, err := models.GetUserGroups(db, user)
	if err != nil {
		log.Println("Could not select usergroups for " + user)
		return
	}
	// map IDs to session
	for _, group := range groups {
		s.Usergroups = append(s.Usergroups, group.ID)
	}
}

// IsValid returns whether the current session is valid or not
// Username and Token must exist and should not be expired
func (s *Session) IsValid() bool {
	if (s.User == models.User{} || s.User.Username == "" || s.Token == "") {
		return false
	}
	return s.Expire.Sub(time.Now()).Seconds() > 0
}

// RenewToken creates a new token and resets the expiry interval
// func (s *Session) RenewToken() string {
// 	b := make([]byte, 30)
// 	rand.Read(b)
// 	s.Token = fmt.Sprintf("%x", b)
// 	s.Expire = time.Now().Add(1 * time.Hour)
// 	return s.Token
// }

// RemoveEmptySessions removes nil objects from session slice
func RemoveEmptySessions(sessions []*Session) []*Session {
	if sessions == nil {
		return sessions
	}
	for i := 0; i < len(sessions); {
		if sessions[i] != nil {
			i++
			continue
		}

		if i < len(sessions)-1 {
			copy(sessions[i:], sessions[i+1:])
		}

		sessions[len(sessions)-1] = nil
		sessions = sessions[:len(sessions)-1]
	}
	return sessions
}
