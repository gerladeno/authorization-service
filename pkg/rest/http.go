package rest

import (
	"compress/flate"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gerladeno/authorization-service/pkg/metrics"
	"github.com/gerladeno/authorization-service/pkg/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
)

const gitURL = "https://github.com/gerladeno/authorization-service"

type TokenProvider interface {
	StartAuthentication(ctx context.Context, user *models.User) error
	SignIn(ctx context.Context, user *models.User, code string) (string, error)
	ParseToken(accessToken string) (string, error)
	GetToken(uuid string) (string, error)
}

type ProfileStore interface {
	GetUser(ctx context.Context, phone string) (*models.User, error)
	UpsertUser(ctx context.Context, user *models.User) error
}

func NewRouter(log *logrus.Logger, provider TokenProvider, store ProfileStore, host, version string) chi.Router {
	handler := newHandler(log, provider, store)
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(cors.AllowAll().Handler)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.NewCompressor(flate.DefaultCompression).Handler)
	r.NotFound(notFoundHandler)
	r.Get("/ping", pingHandler)
	r.Get("/version", versionHandler(version))
	r.Group(func(r chi.Router) {
		r.Use(metrics.NewPromMiddleware(host))
		r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: log, NoColor: true}))
		r.Use(middleware.Timeout(30 * time.Second))
		r.Use(middleware.Throttle(100))
		r.Route("/public", func(r chi.Router) {
			r.Route("/v1", func(r chi.Router) {
				r.Get("/authenticate", handler.authenticate)
				r.Get("/signIn", handler.signIn)
				r.Get("/verify", handler.verify)
				r.Group(func(r chi.Router) {
					r.Use(handler.auth)
					// protected endpoints
				})
			})
		})
		r.Route("/private", func(r chi.Router) {
			r.Use(handler.customAuth)
			r.Route("/v1", func(r chi.Router) {
				r.Get("/token/{uuid}", handler.getToken)
			})
		})
	})
	return r
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "404 page not found. Check docs: "+gitURL, http.StatusNotFound)
}

func pingHandler(w http.ResponseWriter, _ *http.Request) {
	writeResponse(w, "pong")
}

func versionHandler(version string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeResponse(w, version)
	}
}

func writeResponse(w http.ResponseWriter, data interface{}) {
	response := JSONResponse{Data: data}
	w.Header().Set("Content-type", "application/json")
	_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
}

func writeErrResponse(w http.ResponseWriter, message string, status int) {
	response := JSONResponse{Data: []int{}, Error: &message, Code: &status}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
}

type JSONResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Meta  *Meta       `json:"meta,omitempty"`
	Error *string     `json:"error,omitempty"`
	Code  *int        `json:"code,omitempty"`
}

type Meta struct {
	Count int `json:"count"`
}
