package primboard

import (
	"crypto/rand"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Session stores the user data, the token, the expiration of the session and
// the usergroups of the current user
type Session struct {
	User       User
	Token      string
	Expire     time.Time
	Usergroups []primitive.ObjectID
}

// IsValid returns whether the current session is valid or not
func (s *Session) IsValid() bool {
	if (s.User == User{} || s.User.Username == "" || s.Token == "") {
		return false
	}
	return s.Expire.Sub(time.Now()).Seconds() > 0
}

// RenewToken creates a new token and resets the expiry interval
func (s *Session) RenewToken() string {
	b := make([]byte, 30)
	rand.Read(b)
	s.Token = fmt.Sprintf("%x", b)
	s.Expire = time.Now().Add(1 * time.Hour)
	return s.Token
}

// RemoveToken finds the entry with the passed token and deletes it
func RemoveToken(ss []*Session, token string) []*Session {
	for index, s := range ss {
		if s.Token == token {
			return Remove(ss, index)
		}
	}
	return ss
}

// Remove deletes the item from the slice at the passed index
func Remove(ss []*Session, index int) []*Session {
	ss[index] = ss[len(ss)-1]
	ss[len(ss)-1] = new(Session)
	return ss[:len(ss)-1]
}
