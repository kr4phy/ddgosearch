package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limitParam := r.URL.Query().Get("limit")
	region := r.URL.Query().Get("region")
	safeSearchParam := r.URL.Query().Get("safeSearch")

	if query == "" {
		http.Error(w, "Missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 {
		limit = 10 // Default limit
	}

	safeSearch, err := strconv.Atoi(safeSearchParam)
	if err != nil {
		safeSearch = -1 // Default safe search level
	}

	results, err := ScrapeDuckDuckGo(SearchOptions{
		query:       query,
		limit:       limit,
		region:      region,
		safeSearch:  safeSearch,
	})
	if err != nil {
		http.Error(w, "Error scraping DuckDuckGo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/search", searchHandler)

	http.ListenAndServe(":8080", mux)
}
