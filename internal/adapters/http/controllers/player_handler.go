package controllers

import "net/http"

func (c *Controller) me(w http.ResponseWriter, r *http.Request) {
	telegramID, err := telegramID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "valid X-Telegram-User-ID required")
		return
	}
	player, err := c.players.Get(r.Context(), telegramID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load player")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"telegramId": player.TelegramID, "balance": player.BottleBalance, "lastPassiveClaimAt": player.LastPassiveClaimAt})
}
func (c *Controller) claimActive(w http.ResponseWriter, r *http.Request) {
	telegramID, err := telegramID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "valid X-Telegram-User-ID required")
		return
	}
	player, retryAfter, err := c.players.ClaimActive(r.Context(), telegramID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "cooldown service unavailable")
		return
	}
	if retryAfter > 0 {
		writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": "cooldown active", "retryAfterSeconds": int(retryAfter.Seconds())})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bottlesAwarded": 1, "balance": player.BottleBalance})
}
func (c *Controller) claimPassive(w http.ResponseWriter, r *http.Request) {
	telegramID, err := telegramID(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "valid X-Telegram-User-ID required")
		return
	}
	player, bottles, err := c.players.ClaimPassive(r.Context(), telegramID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not claim passive bottles")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bottlesAwarded": bottles, "balance": player.BottleBalance})
}
