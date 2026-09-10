package controllers

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"nabutilivanie/internal/application"
)

type Controller struct {
	targets  *application.TargetService
	players  *application.PlayerService
	payments *application.PaymentService
	secret   []byte
}

func NewRouter(targets *application.TargetService, players *application.PlayerService, payments *application.PaymentService, webhookSecret string) http.Handler {
	c := &Controller{targets: targets, players: players, payments: payments, secret: []byte(webhookSecret)}
	r := chi.NewRouter()
	r.Use(recoverer)
	r.Get("/healthz", c.health)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(jsonContent)
		r.Post("/targets", c.createTarget)
		r.Get("/targets", c.leaderboard)
		r.Get("/targets/search", c.searchTargets)
		r.Post("/targets/{id}/nabutilit", c.nabutilit)
		r.Get("/me", c.me)
		r.Post("/me/claim/active", c.claimActive)
		r.Post("/me/claim/passive", c.claimPassive)
		r.Post("/payments/webhook", c.paymentWebhook)
	})

	static, _ := fs.Sub(webFiles, "web")
	r.Get("/", func(w http.ResponseWriter, req *http.Request) { http.ServeFileFS(w, req, static, "index.html") })
	r.Handle("/*", http.FileServer(http.FS(static)))
	return r
}
