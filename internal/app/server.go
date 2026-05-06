package app

import (
	"net/http"
	"time"

	"proidentity-mail/internal/admin"
)

func NewHTTPServer(cfg Config) *http.Server {
	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           admin.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}
}
