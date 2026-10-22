package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type alertRequest struct{ AssetID, Title, Creator, Phone string }

func main() {
	pipeline := NewPipeline(NewSMSClient())
	http.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method must be POST", http.StatusMethodNotAllowed)
			return
		}
		var in alertRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		job, delivery, err := pipeline.IngestAndNotify(r.Context(), Asset{ID: in.AssetID, Title: in.Title}, in.Creator, in.Phone)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"job": job, "delivery": delivery})
	})
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
