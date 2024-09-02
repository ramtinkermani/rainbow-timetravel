package storageServices

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/rainbowmga/timetravel/entity"
	"github.com/rainbowmga/timetravel/service"
)

var tableName string = "CustomerData"

func initDB() (*sql.DB, error) {
    db, err := sql.Open("sqlite3", "./data/data.db")
    if err != nil {
        return nil, err
    }

    createTableQuery := `
    CREATE TABLE IF NOT EXISTS CustomerData (
        id INTEGER PRIMARY KEY,
        data TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT NULL
    );`

    _, err = db.Exec(createTableQuery)
    if err != nil {
        return nil, err
    }

    return db, nil
}

type SqliteRecordService struct {
	db *sql.DB
}

func NewSqliteRecordService(db *sql.DB) *SqliteRecordService {
	return &SqliteRecordService{
		db: db,
	}
}

func (sqlsvc *SqliteRecordService) GetRecord(ctx context.Context, id int) (entity.Record, error) {
	fmt.Println("GetRecord")
	var record entity.Record
    var dataJSON string

	query := fmt.Sprintf("SELECT id, data FROM %s WHERE id = ?", tableName)
    _ = sqlsvc.db.QueryRow(query, id).Scan(&record.ID, &dataJSON)
	if record.ID == 0 {
		return entity.Record{}, service.ErrRecordDoesNotExist
	}

    err := json.Unmarshal([]byte(dataJSON), &record.Data)
    if err != nil {
        return entity.Record{}, err
    }

	record = record.Copy() // copy is necessary so modifations to the record don't change the stored record
	return record, nil
}

func (sqlsvc *SqliteRecordService) CreateRecord(ctx context.Context, record entity.Record) error {
	fmt.Println("CreateRecord")

	id := record.ID
	if id <= 0 {
		return service.ErrRecordIDInvalid
	}

	existingRecord, _ := sqlsvc.GetRecord(ctx, id)
	if existingRecord.ID != 0 {
		return service.ErrRecordAlreadyExists
	}

	dataJSON, err := json.Marshal(record.Data)
    if err != nil {
        return err
    }

	query := fmt.Sprintf("INSERT INTO %s (id, data) VALUES (?, ?)", tableName)
    _, err = sqlsvc.db.Exec(query, id, string(dataJSON))
    return err
}

func (sqlsvc *SqliteRecordService) UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error) {
	entry, _ := sqlsvc.GetRecord(ctx, id)
	
	if entry.ID == 0 {
		return entity.Record{}, service.ErrRecordDoesNotExist
	}

	for key, value := range updates {
		if value == nil { // deletion update
			delete(entry.Data, key)
		} else {
			entry.Data[key] = *value
		}
	}

	dataJSON, err := json.Marshal(entry.Data)
    if err != nil {
        return entity.Record{},nil
    }

	query := fmt.Sprintf("UPDATE %s SET data=? WHERE id=?", tableName)
    _, err = sqlsvc.db.Exec(query, string(dataJSON), id)

	return entity.Record{},nil
}