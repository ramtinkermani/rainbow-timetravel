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
        _id INTEGER PRIMARY KEY AUTOINCREMENT,
        id INTEGER,
        data TEXT,
		updates text,
		effective_date DATETIME DEFAULT CURRENT_TIMESTAMP,
        created_date DATETIME DEFAULT CURRENT_TIMESTAMP
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


func (sqlsvc *SqliteRecordService) GetRecordVersions(ctx context.Context, id int) ([]entity.Record, error) {
	fmt.Println("Getting All versions")
	var records []entity.Record

	query := fmt.Sprintf("SELECT _id, id, data, updates, created_date, effective_date FROM %s WHERE id = ? ORDER BY _id DESC", tableName)

	rows, err := sqlsvc.db.Query(query, id)
	if err != nil {
		return []entity.Record{}, err
	}
	defer rows.Close()


	// Iterate over the result set and scan each row into a Record struct
	for rows.Next() {
		var record entity.Record
		var dataJSON sql.NullString
		var updatesJSON sql.NullString

		err := rows.Scan(&record.I_id, &record.ID, &dataJSON, &updatesJSON, &record.CreatedDate, &record.EffectiveDate)
		if err != nil {
			return []entity.Record{}, err
		}

		// Unmarshal the JSON data into the Record's Data field
		if dataJSON.Valid{
			err = json.Unmarshal([]byte(dataJSON.String), &record.Data)
			if err != nil {
				return []entity.Record{}, err
			}
		}

		// Same for the Updates 
		if updatesJSON.Valid {
			err = json.Unmarshal([]byte(updatesJSON.String), &record.Updates)
			if err != nil {
				return []entity.Record{}, err
			}
		}

		record = record.Copy()

		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		return []entity.Record{}, err
	}

	if len(records) == 0 {
		return []entity.Record{}, service.ErrRecordDoesNotExist
	}

	// Return the slice of records
	return records, nil

}

func (sqlsvc *SqliteRecordService) GetRecord(ctx context.Context, id int) (entity.Record, error) {
	fmt.Println("GetRecord")
	var record entity.Record
    var dataJSON string

	// Get the latest version of the data from the DB
	query := fmt.Sprintf("SELECT id, data, created_date, effective_date FROM %s WHERE id = ? order by _id desc limit 1", tableName)
    _ = sqlsvc.db.QueryRow(query, id).Scan(&record.ID, &dataJSON, record.EffectiveDate, record.CreatedDate)

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

	dataJSON, err_data := json.Marshal(entry.Data)
	updatesJSON, err_updates := json.Marshal(updates)

    if err_data != nil || err_updates != nil {
        return entity.Record{},nil
    }

	var err error
	// Eventhough this is an Update, since we now have versioning,
	// we insert a new row with the same ID as well as the update 
	// that 'updated' this version
	query := fmt.Sprintf("INSERT INTO %s (id, data, updates) VALUES (?, ?, ?)", tableName)
    _, err = sqlsvc.db.Exec(query, id, string(dataJSON), string(updatesJSON))

	return entity.Record{}, err
}