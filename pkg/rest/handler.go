package rest

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"

	"github.com/google/uuid"

	"github.com/gerladeno/authorization-service/pkg/common"
	"github.com/gerladeno/authorization-service/pkg/models"
	"github.com/sirupsen/logrus"
)

type handler struct {
	log      *logrus.Entry
	provider TokenProvider
	store    ProfileStore
}

func newHandler(log *logrus.Logger, provider TokenProvider, store ProfileStore) *handler {
	h := handler{
		log:      log.WithField("module", "http_in"),
		provider: provider,
		store:    store,
	}
	return &h
}

func (h *handler) authenticate(w http.ResponseWriter, r *http.Request) {
	phone := r.URL.Query().Get("phone")
	if phone == "" {
		writeErrResponse(w, "Bad request", http.StatusBadRequest)
		return
	}
	user, err := h.store.GetUser(r.Context(), phone)
	switch {
	case err == nil:
	case errors.Is(err, common.ErrPhoneNotFound):
		user = &models.User{Phone: phone}
	default:
		h.log.Warnf("err finding user in authentication: %v", err)
		writeErrResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = h.provider.StartAuthentication(r.Context(), user)
	switch {
	case err == nil:
	case errors.Is(err, common.ErrInvalidPhoneNumber):
		writeErrResponse(w, "Bad request", http.StatusBadRequest)
		return
	default:
		h.log.Warnf("err initiating authentication: %v", err)
		writeErrResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeResponse(w, "Ok")
}

func (h *handler) signIn(w http.ResponseWriter, r *http.Request) {
	phone := r.URL.Query().Get("phone")
	if phone == "" {
		writeErrResponse(w, "Bad request", http.StatusBadRequest)
		return
	}
	user, err := h.store.GetUser(r.Context(), phone)
	switch {
	case err == nil:
	case errors.Is(err, common.ErrPhoneNotFound):
		id, e := uuid.NewUUID()
		if e != nil {
			h.log.Warnf("err generating user uuid: %v", err)
			writeErrResponse(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		user = &models.User{Phone: phone, UUID: id.String()}
	default:
		h.log.Warnf("err finding user in signIn: %v", err)
		writeErrResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	code := r.URL.Query().Get("code")
	token, err := h.provider.SignIn(r.Context(), user, code)
	switch {
	case err == nil:
	case errors.Is(err, common.ErrInvalidPhoneNumber):
		writeErrResponse(w, "Bad request", http.StatusBadRequest)
		return
	case errors.Is(err, common.ErrUnauthenticated):
		writeErrResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	default:
		h.log.Warnf("err signing in: %v", err)
		writeErrResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if err = h.store.UpsertUser(r.Context(), user); err != nil {
		h.log.Warnf("err signing in: %v", err)
		writeErrResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeResponse(w, map[string]string{"uuid": user.UUID, "token": token})
}

func (h *handler) verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeErrResponse(w, "Bad request", http.StatusBadRequest)
		return
	}
	id, err := h.provider.ParseToken(token)
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
	writeResponse(w, map[string]string{"userId": id})
}

func (h *handler) getToken(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "uuid")
	if !isValidUUID(id) {
		writeErrResponse(w, "Bad request", http.StatusBadRequest)
		return
	}
	token, err := h.provider.GetToken(id)
	if err != nil {
		h.log.Warnf("err parsing token: %v", err)
		writeErrResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeResponse(w, map[string]string{"uuid": id, "token": token})
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
