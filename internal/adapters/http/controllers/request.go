package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func decodeJSON(w http.ResponseWriter, r *http.Request, value any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}
func telegramID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.Header.Get("X-Telegram-User-ID"), 10, 64)
}
func queryLimit(r *http.Request) int {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		return 50
	}
	if limit > 100 {
		return 100
	}
	return limit
}
