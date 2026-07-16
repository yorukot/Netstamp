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

func newExternalAuthFlowCookie(name, value string, expiresAt time.Time, secure bool) http.Cookie {
	return http.Cookie{ //nolint:gosec // Secure is runtime-configured; app bootstrap enables it outside local development.
		Name:     name,
		Value:    value,
		Path:     sessionCookiePath,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func expiredExternalAuthFlowCookie(name string, secure bool) http.Cookie {
	return http.Cookie{ //nolint:gosec // Secure is runtime-configured; app bootstrap enables it outside local development.
		Name:     name,
		Value:    "",
		Path:     sessionCookiePath,
		Expires:  time.Unix(0, 0).UTC(),
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}
