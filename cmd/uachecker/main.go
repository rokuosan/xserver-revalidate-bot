package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type UserAgentRecord struct {
	UserAgent  string            `json:"user_agent"`
	Headers    map[string]string `json:"headers"`
	RemoteAddr string            `json:"remote_addr"`
	Timestamp  time.Time         `json:"timestamp"`
	Method     string            `json:"method"`
	RequestURI string            `json:"request_uri"`
}

func main() {
	server := NewServer()

	err := server.LoadTemplate("cmd/uachecker/templates/index.html")
	if err != nil {
		log.Fatal("Error parsing template:", err)
	}

	http.HandleFunc("/", server.HandleRequest)
	http.HandleFunc("/api/ua", server.HandleUA)
	http.HandleFunc("/headers", server.HandleHeaders)

	fmt.Println("UA Checker Server starting on :8080")
	fmt.Println("Visit http://localhost:8080 to check your User-Agent")
	fmt.Println("API endpoint: http://localhost:8080/api/ua")
	fmt.Println("Headers endpoint: http://localhost:8080/headers")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createRecord(r *http.Request) UserAgentRecord {
	headers := make(map[string]string)
	for name, values := range r.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}

	return UserAgentRecord{
		UserAgent:  r.UserAgent(),
		Headers:    headers,
		RemoteAddr: r.RemoteAddr,
		Timestamp:  time.Now(),
		Method:     r.Method,
		RequestURI: r.RequestURI,
	}
}

func formatHeaders(headers map[string]string) string {
	result := ""
	for name, value := range headers {
		result += fmt.Sprintf("%s: %s\n", name, value)
	}
	return result
}
