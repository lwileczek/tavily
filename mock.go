package tavily

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type SearchHandler struct{}

func (h *SearchHandler) search(data *TavilyRequest) (TavilyResponse, error) {
	resp := TavilyResponse{
		Query: data.Query,
	}
	if data.IncludeAnswer {
		resp.Answer = "A valid answer"
	}

	for i := uint32(0); i < data.MaxResults; i++ {
		t := TavilyResult{
			Title:      "A nice title",
			URL:        "https://github.com/lwileczek/tavily",
			Content:    "Some modified page content",
			RawContent: nil,
			Score:      rand.Float64(),
		}

		if data.IncludeRawContent {
			raw := "<div>Some modified page content</div>"
			t.RawContent = &raw
		}
	}

	return resp, nil
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Only POST requests are allowed")
		return
	}

	defer r.Body.Close()

	var request TavilyRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid JSON request: %v", err)
		return
	}

	if request.ApiKey == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Invalid API key")
		return
	}

	startTime := time.Now()
	response, err := h.search(&request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error processing search: %v", err)
		return
	}

	response.ResponseTime = time.Since(startTime).Seconds()

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err = encoder.Encode(response)
	if err != nil {
		fmt.Printf("Error encoding response: %v", err)
		return
	}
}
