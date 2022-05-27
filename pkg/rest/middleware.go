package rest

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gerladeno/authorization-service/pkg/common"
)

type idType string

const idKey idType = `userID`

func (h *handler) auth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeErrResponse(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 {
			writeErrResponse(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if headerParts[0] != "Bearer" {
			writeErrResponse(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		id, err := h.provider.ParseToken(headerParts[1])
		switch {
		case err == nil:
		case errors.Is(err, common.ErrInvalidAccessToken):
			writeErrResponse(w, "Unauthorized", http.StatusUnauthorized)
			return
		default:
			h.log.Warnf("err parsing token: %v", err)
			writeErrResponse(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), idKey, id))
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (h *handler) customAuth(next http.Handler) http.Handler {
	var fn http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		secret := r.Header.Get("key")
		if secret != "secret_key" {
			writeErrResponse(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		writeErrResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
		//next.ServeHTTP(w, r)
	}
	return fn
}
