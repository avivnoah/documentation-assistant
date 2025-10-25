package server

import (
	"context"
	"encoding/json"
	"net/http"
)

type QueryRequest struct {
	Query   string `json:"query"`
	NumDocs int    `json:"num_docs"`
}

type QueryResponse struct {
	Result          interface{} `json:"result"`
	Query           string      `json:"query"`
	SourceDocuments interface{} `json:"source_documents,omitempty"`
	Error           string      `json:"error,omitempty"`
}

type IngestRequest struct {
	URL string `json:"url"`
}

type IngestResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
	Error   string `json:"error,omitempty"`
}

// handleQuery processes query requests to the LLM
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.NumDocs == 0 {
		req.NumDocs = 5
	}

	// Import and call your runLLM function
	result, err := runLLM(context.Background(), s.logger, &s.store, req.NumDocs, req.Query)

	if err != nil {
		respondWithError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, QueryResponse{
		Result:          result["result"],
		Query:           req.Query,
		SourceDocuments: result["source_documents"],
	})
}

// handleIngest processes documentation ingestion requests
func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		respondWithError(w, "URL is required", http.StatusBadRequest)
		return
	}

	s.logger.Info(context.Background(), "Starting ingestion", map[string]any{"url": req.URL})

	// Run ingestion in background
	go s.runIngestion(req.URL)

	respondWithJSON(w, IngestResponse{
		Status:  "started",
		Message: "Ingestion process started in background",
		URL:     req.URL,
	})
}

// runIngestion executes the ingestion process asynchronously
func (s *Server) runIngestion(url string) {
	ctx := context.Background()

	// Import and call your ingest function
	if err := Ingest(ctx, s.logger, &s.store, url); err != nil {
		s.logger.Error(ctx, "Ingestion failed", map[string]any{"error": err.Error(), "url": url})
	} else {
		s.logger.Info(ctx, "Ingestion completed successfully", map[string]any{"url": url})
	}
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, map[string]string{
		"status":  "healthy",
		"service": "documentation-assistant",
	})
}

// Helper functions for response handling
func respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
