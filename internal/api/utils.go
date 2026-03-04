package api

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteMessage(w http.ResponseWriter, status int, msgType, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{msgType: msg})
}

func DecodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var data T
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "requête invalide", http.StatusBadRequest)
		return data, err
	}
	return data, nil
}

func WriteInternalError(w http.ResponseWriter) {
	WriteMessage(w, http.StatusInternalServerError, "error", "erreur serveur")
}

func WriteInvalidCredentials(w http.ResponseWriter) {
	WriteMessage(w, http.StatusUnauthorized, "error", "identifiants invalides")
}

func WriteUnauthorized(w http.ResponseWriter) {
	WriteMessage(w, http.StatusUnauthorized, "error", "non autorisé")
}

func WriteForbidden(w http.ResponseWriter, msg string) {
	WriteMessage(w, http.StatusForbidden, "error", msg)
}

func WriteBadRequest(w http.ResponseWriter, msg string) {
	WriteMessage(w, http.StatusBadRequest, "error", msg)
}

func WriteNotFound(w http.ResponseWriter, msg string) {
	WriteMessage(w, http.StatusNotFound, "error", msg)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteMessage(w, status, "error", msg)
}
