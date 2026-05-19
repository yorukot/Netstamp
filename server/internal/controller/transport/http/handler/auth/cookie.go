package auth

import (
	"net/http"
	"time"

	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

const sessionCookiePath = "/"

func newSessionCookie(value string, expiresIn int, secure bool) http.Cookie {
	return http.Cookie{
		Name:     httpmiddleware.SessionCookieName,
		Value:    value,
		Path:     sessionCookiePath,
		Expires:  time.Now().UTC().Add(time.Duration(expiresIn) * time.Second),
		MaxAge:   expiresIn,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func expiredSessionCookie(secure bool) http.Cookie {
	return http.Cookie{
		Name:     httpmiddleware.SessionCookieName,
		Value:    "",
		Path:     sessionCookiePath,
		Expires:  time.Unix(0, 0).UTC(),
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}
