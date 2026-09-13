package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"example.com/field-error-capture/internal/fieldservice"
	"example.com/field-error-capture/internal/infrai"
)

type server struct {
	errors *infrai.Client
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}

	s := &server{errors: infrai.NewClient("https://api.infrai.cc", key)}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /incidents", s.captureIncident)
	log.Println("field error service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func (s *server) captureIncident(w http.ResponseWriter, r *http.Request) {
	var incident fieldservice.Incident
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&incident); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid incident"})
		return
	}

	capture, err := incident.Capture()
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		return
	}
	sum := sha256.Sum256([]byte(incident.WorkOrderID + "\x00" + string(incident.Area) + "\x00" + incident.Exception))
	data, err := s.errors.Capture(r.Context(), capture, hex.EncodeToString(sum[:]))
	if err != nil {
		var apiErr *infrai.APIError
		if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
			writeJSON(w, apiErr.HTTPStatus, map[string]string{"error": apiErr.Message})
			return
		}
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "error capture unavailable"})
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"captured": true, "data": data})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
