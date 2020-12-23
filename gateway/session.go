package gateway

import (
	"net/http"

	iModels "github.com/mirisbowring/primboard/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// NewSession initializes a new Session (if not exists) for the passed user
// otherwise it updates the token
func (g *AppGateway) NewSession(user string, db *mongo.Database, token string) *iModels.Session {
	s := g.GetSessionByUser(user)
	s.NodeTokenMap = make(map[primitive.ObjectID]string)
	s.Token = token
	if s.User == "" {
		s.User = user
		s.InitUserGroups(db, user)
		g.Sessions = append(g.Sessions, s)
	}
	return s
}

// CloseSession deletes the token/session pair from cache and lets the cookie expire
// func (g *AppGateway) CloseSession(w *http.ResponseWriter, r *http.Request) {
// 	token := g.ReadSessionCookie(w, r)
// 	g.Sessions = removeToken(g.Sessions, token)
// 	cookie := http.Cookie{
// 		Name:     g.Config.CookieTokenTitle,
// 		MaxAge:   -1,
// 		Path:     g.Config.CookiePath,
// 		Secure:   g.Config.CookieSecure,
// 		HttpOnly: g.Config.CookieHTTPOnly,
// 		Domain:   g.Config.Domain,
// 	}
// 	http.SetCookie(*w, &cookie)
// }

// GetSessionByUser Returns the session for the passed user if exist
func (g *AppGateway) GetSessionByUser(user string) *iModels.Session {
	// skip iteration if passed argument is invalid
	if user == "" {
		return new(iModels.Session)
	}
	return g.GetSessionByUsername(user)
}

// GetSessionByUsername returns the session for the passed username if exist
func (g *AppGateway) GetSessionByUsername(user string) *iModels.Session {
	// skip iteration if passed argument is invalid
	if user == "" {
		return new(iModels.Session)
	}
	// iterate over cached sessions
	for _, v := range g.Sessions {
		if v.User == user && v.User != "" {
			return v
		}
	}
	return new(iModels.Session)
}

// GetSession returns the session object for the passed token
func (g *AppGateway) GetSession(token string) *iModels.Session {
	for _, s := range g.Sessions {
		if s.Token == token {
			return s
		}
	}
	return nil
}

// SetSessionCookie renews the session attributes and adds the token to the cookie
// func (g *AppGateway) SetSessionCookie(w *http.ResponseWriter, r *http.Request, session *iModels.Session) {
// 	session.RenewToken()
// 	cookie := http.Cookie{
// 		Name:     g.Config.CookieTokenTitle,
// 		Value:    session.Token,
// 		Expires:  session.Expire,
// 		Path:     g.Config.CookiePath,
// 		Secure:   g.Config.CookieSecure,
// 		HttpOnly: g.Config.CookieHTTPOnly,
// 		Domain:   g.Config.CookieDomain,
// 		SameSite: http.SameSite(g.Config.CookieSameSite),
// 	}
// 	http.SetCookie(*w, &cookie)
// }

// ReadSessionCookie reads the stoken cookie from the request and returns the value
func (g *AppGateway) ReadSessionCookie(w *http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie(g.Config.CookieTokenTitle)
	if err != nil {
		// cookie not found or read
		return ""
	}
	return cookie.Value
}

// RemoveSessionByToken finds the session with the given token and removes it
// from the cached session slice
func (g *AppGateway) RemoveSessionByToken(token string) {
	g.Sessions = removeToken(g.Sessions, token)
}

// RemoveToken finds the entry with the passed token and deletes it
func removeToken(ss []*iModels.Session, token string) []*iModels.Session {
	for index, s := range ss {
		if s.Token == token {
			return remove(ss, index)
		}
	}
	return ss
}

// Remove deletes the item from the slice at the passed index
func remove(ss []*iModels.Session, index int) []*iModels.Session {
	ss[index] = ss[len(ss)-1]
	ss[len(ss)-1] = new(iModels.Session)
	return ss[:len(ss)-1]
}
