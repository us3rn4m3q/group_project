package database

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type Site struct {
	ID        int
	Name      string
	BaseURL   string
	CreatedAt time.Time
}

type ParsingHistory struct {
	ID           int
	SiteID       int
	Status       string
	StartedAt    time.Time
	FinishedAt   *time.Time
	ErrorMessage *string
	ItemsParsed  int
}

type ParsedItem struct {
	ID               int
	ParsingHistoryID int
	SiteID           int
	Title            string
	URL              string
	ParsedData       []byte // JSONB data
	CreatedAt        time.Time
}

// NewDB подключение к бд
func NewDB(connStr string) (*DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// CreateParsingHistory Запись о начале парсинга
func (db *DB) CreateParsingHistory(siteID int) (*ParsingHistory, error) {
	var history ParsingHistory
	err := db.QueryRow(`
        INSERT INTO parsing_history (site_id, status)
        VALUES ($1, 'IN_PROGRESS')
        RETURNING id, site_id, status, started_at
    `, siteID).Scan(&history.ID, &history.SiteID, &history.Status, &history.StartedAt)

	if err != nil {
		return nil, err
	}

	return &history, nil
}

// UpdateParsingHistory Обновление статуса парсинга
func (db *DB) UpdateParsingHistory(id int, status string, itemsParsed int, errorMsg *string) error {
	_, err := db.Exec(`
        UPDATE parsing_history 
        SET status = $1, 
            finished_at = CURRENT_TIMESTAMP, 
            items_parsed = $2,
            error_message = $3
        WHERE id = $4
    `, status, itemsParsed, errorMsg, id)

	return err
}

// SaveParsedItem сохраняет спарсенный элемент
func (db *DB) SaveParsedItem(item *ParsedItem) error {
	_, err := db.Exec(`
        INSERT INTO parsed_items 
        (parsing_history_id, site_id, title, url, parsed_data)
        VALUES ($1, $2, $3, $4, $5)
    `, item.ParsingHistoryID, item.SiteID, item.Title, item.URL, item.ParsedData)

	return err
}

// GetParsingHistory получает историю парсинга для сайта
func (db *DB) GetParsingHistory(siteID int) ([]ParsingHistory, error) {
	rows, err := db.Query(`
        SELECT id, site_id, status, started_at, finished_at, error_message, items_parsed
        FROM parsing_history
        WHERE site_id = $1
        ORDER BY started_at DESC
    `, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []ParsingHistory
	for rows.Next() {
		var h ParsingHistory
		err := rows.Scan(
			&h.ID, &h.SiteID, &h.Status, &h.StartedAt,
			&h.FinishedAt, &h.ErrorMessage, &h.ItemsParsed,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	return history, nil
}
