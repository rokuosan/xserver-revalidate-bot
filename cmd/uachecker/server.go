package main

import (
	"encoding/json"
	"net/http"
)

type Server struct {
	renderer *TemplateRenderer
	logger   *Logger
}

type TemplateData struct {
	UserAgent        string
	RemoteAddr       string
	Method           string
	RequestURI       string
	Timestamp        string
	HeadersFormatted string
}

func NewServer() *Server {
	return &Server{
		renderer: NewTemplateRenderer(),
		logger:   NewLogger(nil),
	}
}

func (s *Server) LoadTemplate(templatePath string) error {
	return s.renderer.LoadTemplate(templatePath)
}

func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
	record := createRecord(r)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := TemplateData{
		UserAgent:        record.UserAgent,
		RemoteAddr:       record.RemoteAddr,
		Method:           record.Method,
		RequestURI:       record.RequestURI,
		Timestamp:        record.Timestamp.Format("2006-01-02 15:04:05"),
		HeadersFormatted: formatHeaders(record.Headers),
	}

	err := s.renderer.Render(w, data)
	if err != nil {
		s.logger.LogError("Failed to render template", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.logRequest(record)
}

func (s *Server) HandleUA(w http.ResponseWriter, r *http.Request) {
	record := createRecord(r)

	w.Header().Set("Content-Type", "application/json")

	response := map[string]any{
		"user_agent": record.UserAgent,
		"timestamp":  record.Timestamp,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		s.logger.LogError("Failed to encode JSON response", err.Error())
		return
	}
	s.logRequest(record)
}

func (s *Server) HandleHeaders(w http.ResponseWriter, r *http.Request) {
	record := createRecord(r)

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(record)
	if err != nil {
		s.logger.LogError("Failed to encode headers JSON response", err.Error())
		return
	}
	s.logRequest(record)
}

func (s *Server) logRequest(record UserAgentRecord) {
	s.logger.LogRequest(record)
}
