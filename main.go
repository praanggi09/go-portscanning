package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type ScanResult struct {
	Port int  `json:"port"`
	Open bool `json:"open"`
	// Protocol string `json:"protocol"`
	Service string `json:"service"`
}

type ScanRequest struct {
	Host  string `json:"host"`
	Ports []int  `json:"ports"`
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/scan", scanPorts).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func scanPorts(w http.ResponseWriter, r *http.Request) {
	var request ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results := []ScanResult{}
	for _, port := range request.Ports {
		result := ScanResult{Port: port, Service: identifyService(port)}
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(request.Host, strconv.Itoa(port)), 1*time.Second)
		if err == nil {
			result.Open = true
			conn.Close()
		} else {
			result.Open = false
		}
		results = append(results, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func identifyService(port int) string {
	// Mapping of common ports to services
	services := map[int]string{
		21:    "FTP",
		22:    "SSH",
		23:    "Telnet",
		25:    "SMTP",
		53:    "DNS",
		80:    "HTTP",
		110:   "POP3",
		143:   "IMAP",
		443:   "HTTPS",
		3306:  "MySQL",
		5432:  "PostgreSQL",
		6379:  "Redis",
		27017: "MongoDB",
	}

	if service, exists := services[port]; exists {
		return service
	}
	return "Unknown"
}
