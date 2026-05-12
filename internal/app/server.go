package app

import (
	"net/http"
	"time"

	"proidentity-mail/internal/admin"
)

func NewHTTPServer(cfg Config, store admin.Store, authConfig ...admin.AuthConfig) *http.Server {
	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           admin.NewRouter(store, authConfig...),
		ReadHeaderTimeout: 5 * time.Second,
	}
}
