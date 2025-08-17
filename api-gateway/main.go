package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"group13project/internal/database"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var db *database.DB

type Response struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type AllParsersResponse struct {
	Status string `json:"status"`
	Data   struct {
		BitShop []ParseResult     `json:"bitshop,omitempty"`
		JetMan  []ParseResult     `json:"jetman,omitempty"`
		XCore   []ParseResult     `json:"xcore,omitempty"`
		Ram     []ParseResult     `json:"ram,omitempty"`
		Errors  map[string]string `json:"errors,omitempty"`
	} `json:"data"`
	TimeTaken string `json:"time_taken"`
}

type ParseRequest struct {
	Query string `json:"query"`
}

type ParseResult struct {
	Name    string `json:"name"`
	ItemUrl string `json:"itemUrl"`
	Price   string `json:"price"`
	ImgUrl  string `json:"imgUrl"`
}

func init() {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	var err error
	db, err = database.NewDB(connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
}

func main() {
	http.HandleFunc("/api/v1/status", handleStatus)
	http.HandleFunc("/api/v1/search", handleSearch)
	http.HandleFunc("/api/v1/search/all", handleSearchAll)
	http.HandleFunc("/api/v1/history", handleHistory)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway starting on port %s", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query(`
        SELECT id, site_id, status, started_at, finished_at, error_message, items_parsed
        FROM parsing_history
        ORDER BY started_at DESC
        LIMIT 100
    `)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type History struct {
		ID          int     `json:"id"`
		SiteID      int     `json:"site_id"`
		Status      string  `json:"status"`
		StartedAt   string  `json:"started_at"`
		FinishedAt  *string `json:"finished_at"`
		ErrorMsg    *string `json:"error_message"`
		ItemsParsed int     `json:"items_parsed"`
	}

	var history []History
	for rows.Next() {
		var h History
		err := rows.Scan(&h.ID, &h.SiteID, &h.Status, &h.StartedAt, &h.FinishedAt, &h.ErrorMsg, &h.ItemsParsed)
		if err != nil {
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		history = append(history, h)
	}

	response := Response{
		Status: "success",
		Data:   history,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleSearchAll(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("Received request to /api/v1/search/all")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	log.Printf("Processing search query: %s", query)

	parseRequest := ParseRequest{Query: query}
	jsonData, err := json.Marshal(parseRequest)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	services := map[string]string{
		"bitshop": "http://bitshop:9001/parse",
		"jetman":  "http://jetman:9002/parse",
		"xcore":   "http://xcore:9003/parse",
		"ram":     "http://ram:9004/parse",
	}

	response := AllParsersResponse{
		Status: "success",
		Data: struct {
			BitShop []ParseResult     `json:"bitshop,omitempty"`
			JetMan  []ParseResult     `json:"jetman,omitempty"`
			XCore   []ParseResult     `json:"xcore,omitempty"`
			Ram     []ParseResult     `json:"ram,omitempty"`
			Errors  map[string]string `json:"errors,omitempty"`
		}{
			Errors: make(map[string]string),
		},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, url := range services {
		wg.Add(1)
		go func(serviceName, serviceURL string) {
			defer wg.Done()
			log.Printf("Sending request to %s (%s)", serviceName, serviceURL)

			results, err := sendParseRequest(serviceURL, jsonData)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				log.Printf("Error from %s service: %v", serviceName, err)
				response.Data.Errors[serviceName] = err.Error()
				return
			}

			log.Printf("Received %d results from %s", len(results), serviceName)

			switch serviceName {
			case "bitshop":
				response.Data.BitShop = results
			case "jetman":
				response.Data.JetMan = results
			case "xcore":
				response.Data.XCore = results
			case "ram":
				response.Data.Ram = results
			}
		}(name, url)
	}

	wg.Wait()
	response.TimeTaken = time.Since(startTime).String()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding final response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	log.Printf("Request completed in %s", response.TimeTaken)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	parseRequest := ParseRequest{Query: query}
	jsonData, err := json.Marshal(parseRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var allResults []ParseResult

	services := map[string]string{
		"bitshop": "http://bitshop:9001/parse",
		"jetman":  "http://jetman:9002/parse",
		"ram":     "http://ram:9003/parse",
		"xcore":   "http://xcore:9004/parse",
	}

	for name, url := range services {
		results, err := sendParseRequest(url, jsonData)
		if err != nil {
			log.Printf("Error from %s service: %v", name, err)
			continue
		}
		allResults = append(allResults, results...)
	}

	response := Response{
		Status: "success",
		Data:   allResults,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendParseRequest(url string, jsonData []byte) ([]ParseResult, error) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response struct {
		Status string        `json:"status"`
		Data   []ParseResult `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := Response{
		Status: "success",
		Data:   "API Gateway is running",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
