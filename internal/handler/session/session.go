package session

import (
	iModels "github.com/mirisbowring/primboard/internal/models"
	"github.com/mirisbowring/primboard/models"
)

// GetSessionByUser Returns the session for the passed user if exist
func GetSessionByUser(sessions []*iModels.Session, user models.User) *iModels.Session {
	// skip iteration if passed argument is invalid
	if (user == models.User{} || user.Username == "") {
		return new(iModels.Session)
	}
	return GetSessionByUsername(sessions, user.Username)
}

// GetSessionByUsername returns the session for the passed username if exist
func GetSessionByUsername(sessions []*iModels.Session, username string) *iModels.Session {
	// skip iteration if passed argument is invalid
	if username == "" {
		return new(iModels.Session)
	}
	// iterate over cached sessions
	for _, v := range sessions {
		if (v.User != models.User{} && v.User.Username == username) {
			return v
		}
	}
	return new(iModels.Session)
}

// RemoveSession finds the entry with the passed token and deletes it
func RemoveSession(ss []*iModels.Session, token string) []*iModels.Session {
	for index, s := range ss {
		if s.Token == token {
			return remove(ss, index)
		}
	}
	return ss
}

// remove deletes the item from the slice at the passed index
func remove(ss []*iModels.Session, index int) []*iModels.Session {
	ss[index] = ss[len(ss)-1]
	ss[len(ss)-1] = new(iModels.Session)
	return ss[:len(ss)-1]
}
