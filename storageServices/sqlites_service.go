package storageServices

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/rainbowmga/timetravel/entity"
	"github.com/rainbowmga/timetravel/service"
)

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

    _ = sqlsvc.db.QueryRow("SELECT id, data FROM CustomerData WHERE id = ?", id).Scan(&record.ID, &dataJSON)
	if record.ID == 0 {
		return entity.Record{}, service.ErrRecordDoesNotExist
	}

    _ = json.Unmarshal([]byte(dataJSON), &record.Data)
    // if err != nil {
    //     return nil, err
    // }

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

    _, err = sqlsvc.db.Exec("INSERT INTO CustomerData (id, data) VALUES (?, ?)", id, string(dataJSON))
    return err
}

func (sqlsvc *SqliteRecordService) UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error) {
	fmt.Println("UpdateRecord")
	return entity.Record{}, nil
}