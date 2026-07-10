package auth

import (
	"net/http"
	"time"
)

const sessionCookiePath = "/"

func newSessionCookie(name, value string, expiresIn int, secure bool) http.Cookie {
	sameSite := http.SameSiteStrictMode
	if !secure {
		sameSite = http.SameSiteLaxMode
	}
	return http.Cookie{ //nolint:gosec // Secure is runtime-configured; app bootstrap enables it outside local development.
		Name:     name,
		Value:    value,
		Path:     sessionCookiePath,
		Expires:  time.Now().UTC().Add(time.Duration(expiresIn) * time.Second),
		MaxAge:   expiresIn,
		Secure:   secure,
		HttpOnly: true,
		SameSite: sameSite,
	}
}

func expiredSessionCookie(name string, secure bool) http.Cookie {
	sameSite := http.SameSiteStrictMode
	if !secure {
		sameSite = http.SameSiteLaxMode
	}
	return http.Cookie{ //nolint:gosec // Secure is runtime-configured; app bootstrap enables it outside local development.
		Name:     name,
		Value:    "",
		Path:     sessionCookiePath,
		Expires:  time.Unix(0, 0).UTC(),
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: true,
		SameSite: sameSite,
	}
}
