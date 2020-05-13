package primboard

import (
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

//Sessions is a map of token/session strings for authenticated users
var Sessions = []*Session{}

// NewSession initializes a new Session (if not exists) for the passed user
// otherwise it updates the token
func (a *App) NewSession(user User) *Session {
	s := GetSessionByUser(user)
	s.RenewToken()
	if (s.User == User{} || s.User.Username == "") {
		s.User = user
		s.initUserGroups(a.DB, user.Username)
		Sessions = append(Sessions, s)
	}
	return s
}

// CloseSession deletes the token/session pair from cache and lets the cookie expire
func CloseSession(w *http.ResponseWriter, r *http.Request) {
	token := ReadSessionCookie(w, r)
	Sessions = RemoveToken(Sessions, token)
	cookie := http.Cookie{
		Name:     api.Config.CookieTokenTitle,
		MaxAge:   -1,
		Path:     api.Config.CookiePath,
		Secure:   api.Config.CookieSecure,
		HttpOnly: api.Config.CookieHTTPOnly,
		Domain:   api.Config.Domain,
	}
	http.SetCookie(*w, &cookie)
}

// GetSessionByUser Returns the session for the passed user if exist
func GetSessionByUser(user User) *Session {
	// skip iteration if passed argument is invalid
	if (user == User{} || user.Username == "") {
		return new(Session)
	}
	return GetSessionByUsername(user.Username)
}

// GetSessionByUsername returns the session for the passed username if exist
func GetSessionByUsername(user string) *Session {
	// skip iteration if passed argument is invalid
	if user == "" {
		return new(Session)
	}
	// iterate over cached sessions
	for _, v := range Sessions {
		if (v.User != User{} && v.User.Username == user) {
			return v
		}
	}
	return new(Session)
}

// GetSession returns the session object for the passed token
func GetSession(token string) *Session {
	for _, s := range Sessions {
		if s.Token == token {
			return s
		}
	}
	return nil
}

// SetSessionCookie renews the session attributes and adds the token to the cookie
func SetSessionCookie(w *http.ResponseWriter, r *http.Request, session *Session) {
	session.RenewToken()
	cookie := http.Cookie{
		Name:     api.Config.CookieTokenTitle,
		Value:    session.Token,
		Expires:  session.Expire,
		Path:     api.Config.CookiePath,
		Secure:   api.Config.CookieSecure,
		HttpOnly: api.Config.CookieHTTPOnly,
		Domain:   api.Config.Domain,
	}
	http.SetCookie(*w, &cookie)
}

// ReadSessionCookie reads the stoken cookie from the request and returns the value
func ReadSessionCookie(w *http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(api.Config.CookieTokenTitle)
	if err != nil {
		// cookie not found or read
		return ""
	}
	return cookie.Value
}

func (s *Session) initUserGroups(db *mongo.Database, user string) {
	// preselect usergroups for performance reasons
	groups, err := GetUserGroups(db, user)
	if err != nil {
		log.Println("Could not select usergroups for " + user)
		return
	}
	// map IDs to session
	for _, group := range groups {
		s.Usergroups = append(s.Usergroups, group.ID)
	}
}
