package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	_ "github.com/lib/pq"
)

type Item struct {
	Name    string `json:"name"`
	ItemUrl string `json:"itemUrl"`
	Price   string `json:"price"`
	ImgUrl  string `json:"imgUrl"`
}

type ParseRequest struct {
	Query string `json:"query"`
}

type Response struct {
	Status string `json:"status"`
	Data   []Item `json:"data"`
}

var db *sql.DB

func main() {
	var err error
	db, err = initDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	http.HandleFunc("/parse", handleParse)
	fmt.Println("jetman.by parser is running on :9002")
	log.Fatal(http.ListenAndServe(":9002", nil))
}

func initDB() (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	return db, nil
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received parse request")

	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Received request body: %s", string(body))

	var req ParseRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Error unmarshaling request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Starting parsing for query: %s", req.Query)

	results := WebScraper(req.Query)
	log.Printf("Parsing completed. Found %d items", len(results))

	response := Response{
		Status: "success",
		Data:   results,
	}

	w.Header().Set("Content-Type", "application/json")
	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Sending response: %s", string(responseJSON))
	w.Write(responseJSON)
}

func createParsingHistory(query string) (int, error) {
	var id int
	err := db.QueryRow(`
        INSERT INTO parsing_history (search_query, status)
        VALUES ($1, 'IN_PROGRESS')
        RETURNING id
    `, query).Scan(&id)
	return id, err
}

func saveItem(historyID int, item Item) error {
	_, err := db.Exec(`
        INSERT INTO items (parsing_history_id, name, item_url, price, img_url)
        VALUES ($1, $2, $3, $4, $5)
    `, historyID, item.Name, item.ItemUrl, item.Price, item.ImgUrl)
	return err
}

func updateParsingHistory(id int, itemsCount int) error {
	_, err := db.Exec(`
        UPDATE parsing_history
        SET status = 'SUCCESS',
            items_found = $2,
            finished_at = CURRENT_TIMESTAMP
        WHERE id = $1
    `, id, itemsCount)
	return err
}

func WebScraper(s string) []Item {

	var searchItems []Item
	s = url.QueryEscape(s)
	scrapeURL := "https://jetman.by/search?q=" + s

	if strings.Contains(s, " ") {
		strings.Replace(s, " ", "+", -1)
		scrapeURL = "https://jetman.by/search?q=" + s
	} else {
		scrapeURL = "https://jetman.by/search?q=" + s
	}
	c := colly.NewCollector(colly.AllowedDomains("jetman.by/", "www.jetman.by", "jetman.by", "www.jetman.by/"))

	c.OnHTML("div.product-item", func(e *colly.HTMLElement) {

		price := e.ChildText("div[class=standart]")
		price, _ = strings.CutPrefix(price, "Стандарт")

		price, _, _ = strings.Cut(price, ",")
		price = price + " р"

		itemUrls := e.ChildAttrs("a", "href")
		itemUrl := ""
		if len(itemUrls) > 0 {
			itemUrl = itemUrls[0]
		}

		if itemUrl != "" {
			parsedUrl, err := url.Parse(itemUrl)
			if err == nil && !parsedUrl.IsAbs() {
				baseUrl, _ := url.Parse("https://jetman.by")
				itemUrl = baseUrl.ResolveReference(parsedUrl).String()
			}
		}

		imgUrl := ""

		e.ForEach("div.picture a", func(_ int, el *colly.HTMLElement) {
			style := el.Attr("style")
			re := regexp.MustCompile(`url\(["']?(.*?)["']?\)`)
			matches := re.FindStringSubmatch(style)
			if len(matches) > 1 {
				imgUrl = matches[1]
			}

			oneItem := Item{
				Name:    e.ChildText("div[class=product-title]"),
				ItemUrl: itemUrl,
				Price:   price,
				ImgUrl:  imgUrl,
			}

			if len(searchItems) < 5 {
				searchItems = append(searchItems, oneItem)
			}

		})
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "*/*")
		log.Println("visiting:", r.URL, c)
	})
	c.OnResponse(func(r *colly.Response) {
		log.Println("Got: ", r.Request.URL, c)
	})
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Error: ", r.Request.URL, err, c)
	})

	err := c.Visit(scrapeURL)
	if err != nil {
		log.Println("something went wrong:", err, c)
	}

	return searchItems
}
