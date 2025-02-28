package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/machadovilaca/prometheus-rag/pkg/rag"
)

// Server is the HTTP server for the RAG
type Server struct {
	host string
	port string

	rag *rag.Client
}

// New creates a new Server
func New(host string, port string) (*Server, error) {
	rag, err := rag.New()
	if err != nil {
		return nil, fmt.Errorf("failed to run RAG: %v", err)
	}

	return &Server{
		host: host,
		port: port,

		rag: rag,
	}, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	http.HandleFunc("/healthz", s.handleHealthz)
	http.HandleFunc("/query", s.handleQuery)

	log.Info().Msgf("starting HTTP server on %s:%s", s.host, s.port)
	return http.ListenAndServe(fmt.Sprintf("%s:%s", s.host, s.port), nil)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msgf("received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := s.rag.Query(request.Query)
	if err != nil {
		log.Error().Err(err).Msg("failed to process query")
		http.Error(w, fmt.Sprintf("Failed to process query: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{
		"response": response,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
