package primboard

import (
	"net/http"
)

//Sessions is a map of token/session strings for authenticated users
var Sessions = []*Session{}

//Authenticate is a middleware to pre-authenticate routes via the session token
func Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ReadSessionCookie(&w, r)
		s := GetSession(token)
		if s != nil && s.IsValid() {
			SetSessionCookie(&w, r, s)
			h.ServeHTTP(w, r)
		} else {
			RespondWithError(w, http.StatusUnauthorized, "Your session is invalid")
			return
		}
	})
}

// NewSession initializes a new Session (if not exists) for the passed user
// otherwise it updates the token
func NewSession(user User) *Session {
	s := GetSessionByUser(user)
	s.RenewToken()
	if (s.User == User{} || s.User.Username == "") {
		s.User = user
		Sessions = append(Sessions, s)
	}
	return s
}

// CloseSession deletes the token/session pair from cache and lets the cookie expire
func CloseSession(w *http.ResponseWriter, r *http.Request) {
	token := ReadSessionCookie(w, r)
	Sessions = RemoveToken(Sessions, token)
	cookie := http.Cookie{
		Name:   api.Config.CookieTokenTitle,
		MaxAge: -1,
	}
	http.SetCookie(*w, &cookie)
}

// GetSessionByUser Returns the session for the passed user if exist
func GetSessionByUser(user User) *Session {
	// skip iteration if passed argument is invalid
	if (user == User{} || user.Username == "") {
		return new(Session)
	}
	// iterate over cached sessions
	for _, v := range Sessions {
		if (v.User != User{} && v.User.Username == user.Username) {
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
