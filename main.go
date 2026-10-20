package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
)

type processRequest struct {
	Content     string  `json:"content"`
	EstimateUSD float64 `json:"estimate_usd"`
}
type processResponse struct {
	Delivered bool   `json:"delivered"`
	Reply     string `json:"reply,omitempty"`
}

func main() {
	client, err := NewInfraiClient()
	if err != nil {
		panic(err)
	}
	if err := client.SetBudget(context.Background(), 25, "monthly"); err != nil {
		panic(err)
	}
	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in processRequest
		if json.NewDecoder(r.Body).Decode(&in) != nil || in.Content == "" {
			http.Error(w, "content is required", http.StatusBadRequest)
			return
		}
		usage, err := client.UsageTimeseries(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		if err := AllowProcessing(usage, 25, in.EstimateUSD); err != nil {
			http.Error(w, err.Error(), http.StatusPaymentRequired)
			return
		}
		reply, err := client.Chat(r.Context(), in.Content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(processResponse{Delivered: true, Reply: reply})
	})
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "8080"
	}
	if addr[0] != ':' {
		addr = ":" + addr
	}
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
