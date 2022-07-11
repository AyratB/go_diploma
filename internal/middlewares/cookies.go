package middlewares

import (
	"context"
	"errors"
	"github.com/AyratB/go_diploma/internal/utils"
	"net/http"
)

type CookieHandler struct {
	decoder *utils.Decoder
}

func NewCookieHandler(decoder *utils.Decoder) *CookieHandler {
	return &CookieHandler{decoder: decoder}
}

var allowURLWithoutAuthorization = map[string]bool{
	"/api/user/login":    true,
	"/api/user/register": true,
}

func (c *CookieHandler) CookieHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie(utils.CookieUserName)
		var currentUserLogin = ""

		if !allowURLWithoutAuthorization[r.RequestURI] && errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "need to register or login", http.StatusUnauthorized)
			return
		} else {
			if errors.Is(err, http.ErrNoCookie) {
				next.ServeHTTP(w, r)
			} else if err != nil {
				http.Error(w, "Cookie crumbled", http.StatusInternalServerError)
			} else {
				decodedUserLogin, err := c.decoder.Decode(cookie.Value) // get user login
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
				}

				if len(decodedUserLogin) != 0 {
					currentUserLogin = decodedUserLogin
				}

				ctx := context.WithValue(r.Context(), utils.KeyPrincipalID, currentUserLogin)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		}
	})
}
