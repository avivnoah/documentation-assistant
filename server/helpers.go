package server

import (
	"context"
	"helpers"
	"logging"

	ingestion "github.com/avivnoah/documentation-assistant/pkg/ingestion"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

// Create batches
type batchJob struct {
	batchNum  int
	documents []schema.Document
}

func Ingest(ctx context.Context, logger logging.Logger, store *vectorstores.VectorStore, urlToLearn string) error {
	return ingestion.Ingest(ctx, logger, store, urlToLearn)
}
func runLLM(ctx context.Context, logger logging.Logger, store *vectorstores.VectorStore, numDocs int, query string, conversationMemory *memory.ConversationBuffer) (map[string]any, error) {
	modelName := "gemini"
	llm, err := helpers.InitializeLLM(modelName, "", "")
	if err != nil {
		logger.Error(ctx, "Failed to initialize OpenAI LLM", map[string]any{"error": err.Error()})
		return nil, err
	}

	stuffDocumentsChain := chains.LoadStuffQA(llm)
	condenseQuestionGeneratorChain := chains.LoadCondenseQuestionGenerator(llm)
	qaChain := chains.NewConversationalRetrievalQA(
		stuffDocumentsChain,
		condenseQuestionGeneratorChain,
		vectorstores.ToRetriever(*store, numDocs),
		conversationMemory,
	)
	qaChain.ReturnSourceDocuments = true
	// qaChain.RephraseQuestion = true - I don't know about this one yet.

	logger.Info(ctx, "Running LLM query", map[string]any{"query": query})

	// The chain will call LoadMemoryVariables internally which returns a STRING
	inputValues := map[string]any{
		"question": query,
	}
	// Use the chains.Call wrapper so memory.LoadMemoryVariables is executed
	// (it will also call SaveContext after the chain completes).
	result, err := chains.Call(ctx, qaChain, inputValues)
	if err != nil {
		logger.Error(ctx, "Failed to run QA chain", map[string]any{"error": err.Error()})
		return nil, err
	}

	formatted_result := map[string]any{
		"query":            result["query"],
		"result":           result["text"],
		"source_documents": result["source_documents"],
	}

	logger.Info(ctx, "LLM query completed successfully", map[string]any{"result": formatted_result["result"]})
	return formatted_result, nil
}

func runLLM2_BKP(ctx context.Context, logger logging.Logger, store *vectorstores.VectorStore, numDocs int, query string, chatHistory schema.ChatMessageHistory, conversationMemory *memory.ConversationBuffer) (map[string]any, error) {
	modelName := "gemini"
	llm, err := helpers.InitializeLLM(modelName, "", "")
	if err != nil {
		logger.Error(ctx, "Failed to initialize OpenAI LLM", map[string]any{"error": err.Error()})
		return nil, err
	}

	// llmChain := chains.NewLLMChain(
	// 	llm,
	// 	prompts.NewPromptTemplate(
	// 		localPrompt.REPHRASE_PROMPT,
	// 		[]string{"chat_history", "question"},
	// 	),
	// )
	// stuffDocumentsChain := chains.NewStuffDocuments(llmChain)

	stuffDocumentsChain := chains.LoadStuffQA(llm)
	condenseQuestionGeneratorChain := chains.LoadCondenseQuestionGenerator(llm)
	// historyAwareRetriever := chains.NewConversationalRetrievalQAFromLLM(
	// 	llm,
	// 	vectorstores.ToRetriever(*store, numDocs),
	// 	memory.NewConversationBuffer(memory.WithChatHistory(chatHistory)),
	// )
	// conversationMemory := memory.NewConversationBuffer(
	// 	memory.WithChatHistory(chatHistory),
	// 	memory.WithReturnMessages(true),
	// 	memory.WithInputKey("question"), // Set input key
	// 	memory.WithOutputKey("text"),    // Set output key
	// 	memory.WithHumanPrefix("Human"),
	// 	memory.WithAIPrefix("AI"),
	// )

	qaChain := chains.NewConversationalRetrievalQA(
		stuffDocumentsChain,
		condenseQuestionGeneratorChain,
		// historyAwareRetriever.Retriever,
		vectorstores.ToRetriever(*store, numDocs),
		//memory.NewConversationBuffer(memory.WithReturnMessages(true)),
		conversationMemory,
	)
	qaChain.ReturnSourceDocuments = true
	qaChain.RephraseQuestion = true
	// qaChain.ReturnGeneratedQuestion = true
	// qaChain := chains.NewConversationalRetrievalQA(stuffDocumentsChain, condenseQuestionGeneratorChain, historyAwareRetriever.Retriever, memory.NewConversationBuffer(memory.WithChatHistory(chatHistory)))
	// panic(qaChain.Memory.GetMemoryKey(ctx))
	logger.Info(ctx, "Running LLM query", map[string]any{"query": query})
	chatMessages, err := chatHistory.Messages(ctx)
	if err != nil {
		logger.Error(ctx, "Failed to get chat history", map[string]any{"error": err.Error()})
		return nil, err
	}

	logger.Info(ctx, "Chat history", map[string]any{"chat_history": chatMessages})
	values := map[string]any{
		"question": query,
		// "history":  chatMessages,
		// "input":    formattedChatMessages,
	}
	mmkey := qaChain.Memory.GetMemoryKey(ctx)
	logger.Info(ctx, qaChain.InputKey, map[string]any{"input_key_in_values": values[qaChain.InputKey]})
	logger.Info(ctx, mmkey, map[string]any{"memory_key_in_values": values[mmkey]})
	result, err := qaChain.Call(ctx, values)
	if err != nil {
		logger.Error(ctx, "Failed to run QA chain", map[string]any{"error": err.Error()})
		return nil, err
	}

	err = conversationMemory.SaveContext(ctx, values, result)
	if err != nil {
		logger.Error(ctx, "Failed to save context to memory", map[string]any{"error": err.Error()})
	}

	formatted_result := map[string]any{
		"query":            result["query"],
		"result":           result["text"],
		"source_documents": result["source_documents"],
	}

	logger.Info(ctx, "LLM query completed successfully", map[string]any{"result": formatted_result["result"]})
	return formatted_result, nil
}

// func Ingest(ctx context.Context, logger logging.Logger, store *vectorstores.VectorStore, urlToLearn string) error {
// 	// Configure global certificates at application startup
// 	// if err := certconfig.ConfigureGlobalCerts(); err != nil {
// 	// 	logger.Error(ctx, "Failed to configure certificates: %v", map[string]any{"error": err.Error()})
// 	// 	return err
// 	// }
// 	logger.Info(ctx, "Starting to crawl documentation", map[string]any{"url": urlToLearn})

// 	// Initialize the OpenAIEmbedder with custom configuration
// 	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-3-small"))
// 	if err != nil {
// 		logger.Error(ctx, "Failed to initialize OpenAI LLM", map[string]any{"error": err.Error()})
// 		return err
// 	}
// 	e, err := embeddings.NewEmbedder(llm, embeddings.WithBatchSize(50), embeddings.WithStripNewLines(true))
// 	if err != nil {
// 		logger.Error(ctx, "Failed to create embedder", map[string]any{"error": err.Error()})
// 		return err
// 	}
// 	// Create a new Pinecone vector store.
// 	pineconeHost := os.Getenv("PINECONE_HOST")
// 	if pineconeHost == "" {
// 		logger.Error(ctx, "PINECONE_HOST environment variable not set. Set it to your index endpoint from the Pinecone console, e.g. https://<index>-xxxx.svc.us-west1-gcp.pinecone.io", map[string]any{"error": err.Error()})
// 		return err
// 	}
// 	*store, err = pinecone.New(
// 		pinecone.WithHost(pineconeHost),
// 		pinecone.WithEmbedder(e),
// 		pinecone.WithNameSpace("lc-docs-ns"),
// 	)
// 	if err != nil {
// 		logger.Error(ctx, "Failed to create Pinecone vector store", map[string]any{"error": err.Error()})
// 		return err
// 	}
// 	tavilyCrawl := tavilycrawl.New(tavilycrawl.Options{APIKey: os.Getenv("TAVILY_API_KEY")})

// 	// tavilyExtract := tavilyextract.New(tavilyextract.Options{APIKey: os.Getenv("TAVILY_API_KEY")})
// 	// tavilyMap := tavilymap.New(tavilymap.Options{APIKey: os.Getenv("TAVILY_API_KEY")})
// 	// fmt.Println("Tavily tools initialized", tavilyExtract.Name(), tavilyMap.Name(), tavilyCrawl.Name())

// 	crawlResp, err := tavilyCrawl.CallRaw(ctx, urlToLearn, tavilycrawl.CrawlParams{
// 		MaxDepth:     1,
// 		Limit:        100,
// 		MaxBreadth:   15,
// 		ExtractDepth: "advanced",
// 		// Instructions: "content on ai agents",
// 	})
// 	if err != nil {
// 		logger.Error(ctx, "Tavily crawl failed", map[string]any{"error": err.Error()})
// 		return err
// 	}
// 	logger.Info(ctx, "Successfully crawled the documentation site", map[string]any{
// 		"base_url":      crawlResp.BaseURL,
// 		"pages_crawled": len(crawlResp.Results),
// 		"response_time": crawlResp.ResponseTime,
// 	})

// 	// Access results by index
// 	allDocs := make([]schema.Document, 0, len(crawlResp.Results))
// 	for _, result := range crawlResp.Results {
// 		allDocs = append(allDocs, schema.Document{
// 			PageContent: result.RawContent,
// 			Metadata: map[string]any{
// 				"source": result.URL,
// 			},
// 		})
// 	}

// 	// Split document into chunks
// 	chunkSize := 4000
// 	chunkOverlap := 200
// 	logger.Info(ctx, "Splitting documents into chunks:", map[string]any{"total documents": len(allDocs), "chunk_size": chunkSize, "chunk_overlap": chunkOverlap})
// 	splitter := textsplitter.NewRecursiveCharacter(
// 		textsplitter.WithChunkSize(chunkSize),
// 		textsplitter.WithChunkOverlap(chunkOverlap),
// 	)
// 	// Embed & Store
// 	// Unify the new chunks into documents
// 	// Build documents from all input documents' chunks, preserving metadata
// 	documents := make([]schema.Document, 0)
// 	totalChunks := 0
// 	for docIdx, doc := range allDocs {
// 		chunks, err := splitter.SplitText(doc.PageContent)
// 		if err != nil {
// 			logger.Error(ctx, "Failed to split text", map[string]any{"error": err.Error(), "doc_index": docIdx})
// 			return err
// 		}
// 		for chunkIdx, chunk := range chunks {
// 			// copy metadata and add provenance fields
// 			meta := map[string]any{}
// 			if doc.Metadata != nil {
// 				for k, v := range doc.Metadata {
// 					meta[k] = v
// 				}
// 			}
// 			meta["doc_index"] = docIdx
// 			meta["chunk_index"] = chunkIdx

// 			documents = append(documents, schema.Document{
// 				PageContent: chunk,
// 				Metadata:    meta,
// 			})
// 			totalChunks++
// 		}
// 	}

// 	logger.Info(ctx, "Successfully split all documents into chunks", map[string]any{
// 		"total_documents": len(allDocs),
// 		"total_chunks":    totalChunks,
// 	})

// 	// Store documents in batches
// 	batchSize := 50
// 	numWorkers := 5 // Adjust based on rate limits and performance
// 	totalBatches := (len(documents) + batchSize - 1) / batchSize
// 	logger.Info(ctx, "Processing documents in batches with worker pool", map[string]any{
// 		"batch_size":    batchSize,
// 		"total_batches": totalBatches,
// 		"workers":       numWorkers,
// 	})

// 	// Channel for jobs and results
// 	jobs := make(chan batchJob, totalBatches)
// 	results := make(chan struct {
// 		batchNum int
// 		ids      []string
// 		err      error
// 	}, totalBatches)

// 	// Create worker pool
// 	for w := 1; w <= numWorkers; w++ {
// 		go func(workerID int) {
// 			for job := range jobs {
// 				logger.Info(ctx, "Worker processing batch", map[string]any{
// 					"worker":        workerID,
// 					"batch":         job.batchNum,
// 					"total_batches": totalBatches,
// 					"batch_size":    len(job.documents),
// 				})

// 				ids, err := (*store).AddDocuments(ctx, job.documents)

// 				results <- struct {
// 					batchNum int
// 					ids      []string
// 					err      error
// 				}{
// 					batchNum: job.batchNum,
// 					ids:      ids,
// 					err:      err,
// 				}

// 				if err != nil {
// 					logger.Error(ctx, "Worker failed to store batch", map[string]any{
// 						"worker": workerID,
// 						"batch":  job.batchNum,
// 						"error":  err.Error(),
// 					})
// 				} else {
// 					logger.Info(ctx, "Worker successfully stored batch", map[string]any{
// 						"worker":     workerID,
// 						"batch":      job.batchNum,
// 						"batch_size": len(ids),
// 					})
// 				}
// 			}
// 		}(w)
// 	}

// 	// Send jobs to workers
// 	go func() {
// 		for i := 0; i < len(documents); i += batchSize {
// 			end := i + batchSize
// 			if end > len(documents) {
// 				end = len(documents)
// 			}

// 			batch := documents[i:end]
// 			batchNum := (i / batchSize) + 1

// 			jobs <- batchJob{
// 				batchNum:  batchNum,
// 				documents: batch,
// 			}
// 		}
// 		close(jobs)
// 	}()

// 	// Collect results
// 	allIDs := make([]string, 0, len(documents))
// 	var firstError error
// 	for i := 0; i < totalBatches; i++ {
// 		result := <-results
// 		if result.err != nil && firstError == nil {
// 			firstError = result.err
// 		}
// 		allIDs = append(allIDs, result.ids...)
// 	}
// 	close(results)

// 	if firstError != nil {
// 		logger.Error(ctx, "Failed to store all batches", map[string]any{"error": firstError.Error()})
// 		return firstError
// 	}

// 	logger.Info(ctx, "Successfully stored all documents concurrently", map[string]any{"total_count": len(allIDs)})
// 	logger.Info(ctx, "PIPELINE COMPLETED, INGESTION FINISHED SUCCESSFULLY.", map[string]any{})
// 	logger.Info(ctx, "urls crawled", map[string]any{"base_url": crawlResp.BaseURL})
// 	logger.Info(ctx, "total documents ingested", map[string]any{"count": len(allDocs)})
// 	logger.Info(ctx, "total chunks Created", map[string]any{"count": len(documents)})
// 	return nil
// }

// func runMain() {
// 	var store vectorstores.VectorStore
// 	var err error

// 	// retrieval_and_generation()
// 	helpers.LoadDotEnv(".env")
// 	logger := logging.New()
// 	ctx := context.Background()

// 	// Create a new Pinecone vector store.
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
// 		logger.Error(ctx, "PINECONE_HOST environment variable not set. Set it to your index endpoint from the Pinecone console, e.g. https://<index>-xxxx.svc.us-west1-gcp.pinecone.io", map[string]any{"error": err.Error()})
// 		return
// 	}
// 	store, err = pinecone.New(
// 		pinecone.WithHost(pineconeHost),
// 		pinecone.WithEmbedder(e),
// 		pinecone.WithNameSpace("lc-docs-ns"),
// 	)

// 	err = Ingest(context.Background(), logging.New(), &store, "https://python.langchain.com/")
// 	res, err := runLLM(context.Background(), logger, &store, 5, "What is a Langchain chain?")
// 	if err != nil {
// 		logger.Error(ctx, "Failed to run retrieval and generation", map[string]any{"error": err.Error()})
// 	}
// 	logger.Info(ctx, "Retrieval and generation result", map[string]any{"result": res["result"]})
// }
