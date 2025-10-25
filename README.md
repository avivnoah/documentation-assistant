# Documentation Assistant

A small, local Retrieval-Augmented Generation (RAG) assistant: a Go backend that manages ingestion, vector storage and retrieval, plus a Python Streamlit UI for interactive queries.

This repository contains:

- `main.go` — program entrypoint; starts the HTTP server that exposes endpoints for querying and ingestion.
- `server/` — server package: HTTP handlers, server bootstrap and bridges to ingestion/query logic.
- `pkg/ingestion` (optional) — ingestion pipeline that crawls documentation, splits, embeds and stores vectors.
- `prompt.go` — prompt template used by the query chain.
- `app/core.py` — Streamlit frontend that calls the Go backend.
- `Makefile` — small convenience helper to build and run the services.

If you just want to run the system locally, the Makefile provides simple targets to build and run components.

## Features

- Query a vector-backed knowledge base via an HTTP endpoint (`/run`).
- Ingest new documentation by submitting a URL to the `/ingest` endpoint (background ingestion).
- Reset/clear vector database via the `/reset` endpoint (implementation depends on your vector provider).
- Streamlit UI for interactive querying and triggering ingestion/reset operations.

## Prerequisites

- Go 1.20+ (or the version declared in `go.mod`)
- Python 3.10+ and `streamlit` installed for the frontend
- A Pinecone account (or compatible vector store) and API endpoint if you use Pinecone
- Optional: Tavily API key for the crawler integration

## Environment

Copy (or create) an `.env` file in the project root containing the secrets/config your services need. Example variables used by the project:

```
PINECONE_HOST=https://<index>-xxxx.svc.us-west1-gcp.pinecone.io
TAVILY_API_KEY=your_tavily_api_key
OPENAI_API_KEY=sk-xxxx
PORT=8080
GO_LLM_URL=http://localhost:8080   # used by the Streamlit app if overriding default
```

Never commit `.env` to version control. Add any secrets to your machine's environment or a secret manager for production.

## Quickstart (local development)

Open a terminal and run the following from the project root:

1. Build the Go binary:

```bash
make build
```

2. Run just the Go server (reads `.env` if present):

```bash
make run-go
```

3. In a separate terminal, run the Streamlit UI:

```bash
make run-python
```

Or run both together (Go binary in background and Streamlit in foreground):

```bash
make run-both
```

Notes:
- `make run-both` will start the Go binary in the background within the current terminal session and then launch Streamlit. When Streamlit exits, the Make target attempts to stop the background Go process.

Alternative (recommended for long-running work): start services in separate terminal sessions or use `tmux` so logs and lifecycle are easier to manage.

## Makefile targets

- `make build` — runs `go mod tidy` and `go build` to create the `bin/documentation-assistant` binary.
- `make run-go` — builds (if necessary) and runs the Go binary (loads `.env` if present).
- `make run-python` — starts the Streamlit UI (with `PYTHONUNBUFFERED=1`).
- `make run-both` — starts the Go binary in background then runs Streamlit in foreground.
- `make clean` — removes the `bin/` directory.

## Project structure

```
documentation-assistant/
├── Makefile
├── README.md
├── go.mod
├── main.go
├── prompt.go
├── server/
│   ├── server.go
│   └── handlers.go
├── pkg/
│   └── ingestion/
├── app/
│   └── core.py
└── .vscode/
```

## Ingestion & Vector Store

The ingestion pipeline (crawling, splitting, embedding and storing) is intentionally implemented as plain Go code so it can run efficiently and be invoked asynchronously from the server. The ingestion implementation targets Pinecone via the `pinecone` client wrapper used in this project. Reset operations depend on the provider's API — the server exposes `/reset` to trigger a reset, but you should implement provider-specific deletion logic if you need a full programmatic teardown.

## Troubleshooting

- `go build` fails with module errors: run `go mod tidy` and ensure `go.mod` module path matches your imports.
- `ModuleNotFoundError: No module named 'streamlit'`: install Streamlit in your Python environment — e.g. `pip install streamlit`.
- `make: *** missing separator.`: Makefiles require literal TAB characters at the start of recipe lines. The included Makefile is already corrected, but if you edit it make sure to keep tabs.

## Contributing

---

Todos
- add a `README` badge or CI config (GitHub Actions) to run `go vet`/`go test`/`go build`,
- add a `requirements.txt` for the Python side,
- create a simple `.github/workflows/ci.yml` to run the build on pushes.
- move the non-public tools/helpers code into `internal/` packages and public reusable components into `pkg/`.
- Add unit tests for ingestion and query logic where possible.
