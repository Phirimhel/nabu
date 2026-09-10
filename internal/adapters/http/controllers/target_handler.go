package controllers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (c *Controller) createTarget(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name"`
		PhotoURL    string `json:"photoUrl"`
		Description string `json:"description"`
	}
	if err := decodeJSON(w, r, &input); err != nil || strings.TrimSpace(input.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	target, err := c.targets.Create(r.Context(), input.Name, input.PhotoURL, input.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create target")
		return
	}
	writeJSON(w, http.StatusCreated, target)
}
func (c *Controller) leaderboard(w http.ResponseWriter, r *http.Request) {
	targets, err := c.targets.Leaderboard(r.Context(), queryLimit(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load leaderboard")
		return
	}
	writeJSON(w, http.StatusOK, targets)
}
func (c *Controller) searchTargets(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeError(w, http.StatusBadRequest, "q is required")
		return
	}
	targets, err := c.targets.Search(r.Context(), query, queryLimit(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not search targets")
		return
	}
	writeJSON(w, http.StatusOK, targets)
}
func (c *Controller) nabutilit(w http.ResponseWriter, r *http.Request) {
	telegramID, err := telegramID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "valid X-Telegram-User-ID required")
		return
	}
	var input struct {
		Comment string `json:"comment"`
	}
	if err := decodeJSON(w, r, &input); err != nil || strings.TrimSpace(input.Comment) == "" {
		writeError(w, http.StatusBadRequest, "comment is required")
		return
	}
	target, err := c.targets.Nabutilit(r.Context(), telegramID, chi.URLParam(r, "id"), input.Comment)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, target)
}
