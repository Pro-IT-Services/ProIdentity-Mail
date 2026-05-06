package app

import (
	"net/http"
	"time"

	"proidentity-mail/internal/admin"
)

func NewHTTPServer(cfg Config, store admin.Store) *http.Server {
	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           admin.NewRouter(store),
		ReadHeaderTimeout: 5 * time.Second,
	}
}
