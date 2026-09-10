package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
)

func (c *Controller) paymentWebhook(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	mac := hmac.New(sha256.New, c.secret)
	mac.Write(raw)
	if !hmac.Equal([]byte(hex.EncodeToString(mac.Sum(nil))), []byte(r.Header.Get("X-Webhook-Signature"))) {
		writeError(w, http.StatusUnauthorized, "invalid signature")
		return
	}
	var input struct {
		Provider   string `json:"provider"`
		ExternalID string `json:"externalId"`
		TelegramID int64  `json:"telegramId"`
		Bottles    int64  `json:"bottles"`
	}
	if json.Unmarshal(raw, &input) != nil || input.Provider == "" || input.ExternalID == "" || input.TelegramID <= 0 || input.Bottles <= 0 {
		writeError(w, http.StatusBadRequest, "invalid payment event")
		return
	}
	credited, player, err := c.payments.Process(r.Context(), input.Provider, input.ExternalID, input.TelegramID, input.Bottles, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not process payment")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"credited": credited, "balance": player.BottleBalance})
}
