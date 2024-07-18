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
	Port     int    `json:"port"`
	Open     bool   `json:"open"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
}

type ScanRequest struct {
	Host string `json:"host"`
}

type HostStatus struct {
	Host      string `json:"host"`
	Reachable bool   `json:"reachable"`
}

type FullScanResult struct {
	Ping  HostStatus   `json:"ping"`
	Ports []ScanResult `json:"ports"`
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

	if !isValidHost(request.Host) {
		http.Error(w, "Invalid hostname or IP address", http.StatusBadRequest)
		return
	}

	// Perform ping check
	pingResult := pingHost(request.Host)

	// Scan specified ports
	ports := []int{21, 22, 23, 25, 53, 80, 110, 143, 443, 3306, 5432, 6379, 27017}
	results := []ScanResult{}
	for _, port := range ports {
		result := ScanResult{
			Port:     port,
			Protocol: "tcp",
			Service:  identifyService(port),
		}
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(request.Host, strconv.Itoa(port)), 1*time.Second)
		if err == nil {
			result.Open = true
			conn.Close()
		} else {
			result.Open = false
		}
		results = append(results, result)
	}

	fullResult := FullScanResult{
		Ping:  pingResult,
		Ports: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fullResult)
}

func isValidHost(host string) bool {
	// Check if the host is a valid IP address
	if net.ParseIP(host) != nil {
		return true
	}

	// Check if the host is a valid hostname
	if _, err := net.LookupHost(host); err == nil {
		return true
	}

	return false
}

func pingHost(host string) HostStatus {
	pingable := false

	conn, err := net.DialTimeout("udp", net.JoinHostPort(host, "12345"), 1*time.Second)
	if err != nil {
		return HostStatus{Host: host, Reachable: pingable}
	}
	defer conn.Close()

	pingable = true
	return HostStatus{Host: host, Reachable: pingable}
}

// func pingHost(host string) HostStatus {
// 	pinger, err := ping.NewPinger(host)
// 	if err != nil {
// 		return HostStatus{Host: host, Reachable: false}
// 	}
// 	pinger.Count = 3
// 	pinger.Timeout = 3 * time.Second
// 	err = pinger.Run()
// 	if err != nil {
// 		return HostStatus{Host: host, Reachable: false}
// 	}
// 	stats := pinger.Statistics()
// 	return HostStatus{Host: host, Reachable: stats.PacketsRecv > 0}
// }

// func pingHost(host string) HostStatus {
// 	pinger, err := ping.NewPinger(host)
// 	if err != nil {
// 		return HostStatus{Host: host, Reachable: false}
// 	}
// 	pinger.Count = 3
// 	pinger.Timeout = 3 * time.Second
// 	err = pinger.Run()
// 	if err != nil {
// 		return HostStatus{Host: host, Reachable: false}
// 	}
// 	stats := pinger.Statistics()
// 	return HostStatus{Host: host, Reachable: stats.PacketsRecv > 0}
// }

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
