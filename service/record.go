package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rainbowmga/timetravel/entity"
)

var ErrRecordDoesNotExist = errors.New("record with that id does not exist")
var ErrRecordIDInvalid = errors.New("record id must >= 0")
var ErrRecordAlreadyExists = errors.New("record already exists")

// Implements method to get, create, and update record data.
type RecordService interface {

	// GetRecord will retrieve an record.
	GetRecord(ctx context.Context, id int) (entity.Record, error)

	// CreateRecord will insert a new record.
	//
	// If it a record with that id already exists it will fail.
	CreateRecord(ctx context.Context, record entity.Record) error

	// UpdateRecord will change the internal `Map` values of the record if they exist.
	// if the update[key] is null it will delete that key from the record's Map.
	//
	// UpdateRecord will error if id <= 0 or the record does not exist with that id.
	UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error)
}

// InMemoryRecordService is an in-memory implementation of RecordService.
type InMemoryRecordService struct {
	data map[int]entity.Record
}

type SqliteRecordService struct {
	db *sql.DB
}

func NewInMemoryRecordService() InMemoryRecordService {
	return InMemoryRecordService{
		data: map[int]entity.Record{},
	}
}

func NewSqliteRecordService(db *sql.DB) SqliteRecordService {
	return SqliteRecordService{
		db: db,
	}
}

func (sqlsvc *SqliteRecordService) GetRecord(ctx context.Context, id int) (entity.Record, error) {
	fmt.Println("GetRecord")
	var record entity.Record
    var dataJSON string

    _ = sqlsvc.db.QueryRow("SELECT id, data FROM CustomerData WHERE id = ?", id).Scan(&record.ID, &dataJSON)
	if record.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
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
		return ErrRecordIDInvalid
	}

	existingRecord, _ := sqlsvc.GetRecord(ctx, id)
	if existingRecord.ID != 0 {
		return ErrRecordAlreadyExists
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


func (s *InMemoryRecordService) GetRecord(ctx context.Context, id int) (entity.Record, error) {
	record := s.data[id]
	if record.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}

	record = record.Copy() // copy is necessary so modifations to the record don't change the stored record
	return record, nil
}

func (s *InMemoryRecordService) CreateRecord(ctx context.Context, record entity.Record) error {
	id := record.ID
	if id <= 0 {
		return ErrRecordIDInvalid
	}

	existingRecord := s.data[id]
	if existingRecord.ID != 0 {
		return ErrRecordAlreadyExists
	}

	s.data[id] = record
	return nil
}

func (s *InMemoryRecordService) UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error) {
	entry := s.data[id]
	if entry.ID == 0 {
		return entity.Record{}, ErrRecordDoesNotExist
	}

	for key, value := range updates {
		if value == nil { // deletion update
			delete(entry.Data, key)
		} else {
			entry.Data[key] = *value
		}
	}

	return entry.Copy(), nil
}
