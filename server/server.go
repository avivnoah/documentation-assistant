package server

import (
	"context"
	"fmt"
	"logging"
	"net/http"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pinecone"
)

type Server struct {
	store  vectorstores.VectorStore
	logger logging.Logger
	port   string
}

type Config struct {
	PineconeHost string
	Port         string
}

// NewServer creates and initializes a new server instance
func NewServer(ctx context.Context, logger logging.Logger, config Config) (*Server, error) {
	store, err := initializeVectorStore(ctx, logger, config.PineconeHost)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize vector store: %w", err)
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	return &Server{
		store:  store,
		logger: logger,
		port:   config.Port,
	}, nil
}

// initializeVectorStore creates and configures the Pinecone vector store
func initializeVectorStore(ctx context.Context, logger logging.Logger, pineconeHost string) (vectorstores.VectorStore, error) {
	if pineconeHost == "" {
		return nil, fmt.Errorf("PINECONE_HOST environment variable not set")
	}

	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-3-small"))
	if err != nil {
		logger.Error(ctx, "Failed to initialize OpenAI LLM", map[string]any{"error": err.Error()})
		return nil, err
	}

	embedder, err := embeddings.NewEmbedder(llm,
		embeddings.WithBatchSize(50),
		embeddings.WithStripNewLines(true))
	if err != nil {
		logger.Error(ctx, "Failed to create embedder", map[string]any{"error": err.Error()})
		return nil, err
	}

	store, err := pinecone.New(
		pinecone.WithHost(pineconeHost),
		pinecone.WithEmbedder(embedder),
		pinecone.WithNameSpace("lc-docs-ns"),
	)
	if err != nil {
		logger.Error(ctx, "Failed to create Pinecone vector store", map[string]any{"error": err.Error()})
		return nil, err
	}

	logger.Info(ctx, "Vector store initialized successfully", map[string]any{"host": pineconeHost})
	return store, nil
}

// Start begins listening for HTTP requests
func (s *Server) Start(ctx context.Context) error {
	s.registerHandlers()
	s.logServerInfo(ctx)

	if err := http.ListenAndServe(":"+s.port, nil); err != nil {
		s.logger.Error(ctx, "Server failed", map[string]any{"error": err.Error()})
		return err
	}
	return nil
}

// registerHandlers sets up all HTTP endpoints
func (s *Server) registerHandlers() {
	http.HandleFunc("/run", s.handleQuery)
	http.HandleFunc("/ingest", s.handleIngest)
	http.HandleFunc("/health", s.handleHealth)
}

// logServerInfo prints server startup information
func (s *Server) logServerInfo(ctx context.Context) {
	s.logger.Info(ctx, "Starting HTTP server", map[string]any{"port": s.port})
	fmt.Printf("Server running on http://localhost:%s\n", s.port)
	fmt.Printf("Endpoints available:\n")
	fmt.Printf("  POST /run     - Query the documentation\n")
	fmt.Printf("  POST /ingest  - Ingest new documentation\n")
	fmt.Printf("  GET  /health  - Health check\n")
}

// func runServer() {
// 	var store vectorstores.VectorStore
// 	helpers.LoadDotEnv(".env")
// 	logger := logging.New()
// 	ctx := context.Background()

// 	// Initialize vector store once at startup
// 	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-3-small"))
// 	if err != nil {
// 		logger.Error(ctx, "Failed to initialize OpenAI LLM", map[string]any{"error": err.Error()})
// 		return
// 	}

// 	e, err := embeddings.NewEmbedder(llm, embeddings.WithBatchSize(50), embeddings.WithStripNewLines(true))
// 	if err != nil {
// 		logger.Error(ctx, "Failed to create embedder", map[string]any{"error": err.Error()})
// 		return
// 	}

// 	pineconeHost := os.Getenv("PINECONE_HOST")
// 	if pineconeHost == "" {
// 		logger.Error(ctx, "PINECONE_HOST environment variable not set", map[string]any{})
// 		return
// 	}

// 	store, err = pinecone.New(
// 		pinecone.WithHost(pineconeHost),
// 		pinecone.WithEmbedder(e),
// 		pinecone.WithNameSpace("lc-docs-ns"),
// 	)
// 	if err != nil {
// 		logger.Error(ctx, "Failed to create Pinecone vector store", map[string]any{"error": err.Error()})
// 		return
// 	}

// 	// HTTP handler for /run endpoint
// 	http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodPost {
// 			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 			return
// 		}

// 		var req QueryRequest
// 		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 			json.NewEncoder(w).Encode(QueryResponse{Error: "Invalid request body"})
// 			return
// 		}

// 		if req.NumDocs == 0 {
// 			req.NumDocs = 5
// 		}

// 		result, err := runLLM(context.Background(), logger, &store, req.NumDocs, req.Query)

// 		w.Header().Set("Content-Type", "application/json")
// 		if err != nil {
// 			json.NewEncoder(w).Encode(QueryResponse{Error: err.Error()})
// 			return
// 		}

// 		json.NewEncoder(w).Encode(QueryResponse{
// 			Result:          result["result"],
// 			Query:           req.Query,
// 			SourceDocuments: result["source_documents"],
// 		})
// 	})

// 	// New endpoint: /ingest - to ingest new documentation
// 	http.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodPost {
// 			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 			return
// 		}

// 		var req struct {
// 			URL string `json:"url"`
// 		}
// 		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 			w.Header().Set("Content-Type", "application/json")
// 			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
// 			return
// 		}

// 		if req.URL == "" {
// 			w.Header().Set("Content-Type", "application/json")
// 			json.NewEncoder(w).Encode(map[string]string{"error": "URL is required"})
// 			return
// 		}

// 		logger.Info(ctx, "Starting ingestion", map[string]any{"url": req.URL})

// 		// Run ingestion in background
// 		go func() {
// 			if err := ingest(context.Background(), logger, &store, req.URL); err != nil {
// 				logger.Error(ctx, "Ingestion failed", map[string]any{"error": err.Error(), "url": req.URL})
// 			} else {
// 				logger.Info(ctx, "Ingestion completed successfully", map[string]any{"url": req.URL})
// 			}
// 		}()

// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(map[string]string{
// 			"status":  "started",
// 			"message": "Ingestion process started in background",
// 			"url":     req.URL,
// 		})
// 	})

// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "8080"
// 	}

// 	logger.Info(ctx, "Starting HTTP server", map[string]any{"port": port})
// 	fmt.Printf("Server running on http://localhost:%s\n", port)

// 	if err := http.ListenAndServe(":"+port, nil); err != nil {
// 		logger.Error(ctx, "Server failed", map[string]any{"error": err.Error()})
// 	}
// }
