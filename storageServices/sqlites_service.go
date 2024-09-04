package storageServices

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rainbowmga/timetravel/contexthelpers"
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

	// Retrieve the effectiveDate from the context
	effective_date_str, ok := ctx.Value(contexthelpers.EffectiveDateKey).(string)

	var effectiveDate string
	if ok{
		// Parse the date string into a time.Time object
		effective_date_parsed, err := time.Parse("2006-01-02", effective_date_str)
		if err != nil {
			effectiveDate = time.Now().UTC().Format("2006-01-02 15:04:05")
		} else {
			// Format the date as a string suitable for SQLite (SQLite uses ISO8601 format for datetime)
			effectiveDate = effective_date_parsed.Format("2006-01-02 15:04:05")
		}
	}

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
        return entity.Record{}, nil
    }

	var err error
	// Eventhough this is an Update, since we now have versioning,
	// we insert a new row with the same ID as well as the update 
	// that 'updated' this version
	query := fmt.Sprintf("INSERT INTO %s (id, data, updates, effective_date) VALUES (?, ?, ?, ?)", tableName)
    _, err = sqlsvc.db.Exec(query, id, string(dataJSON), string(updatesJSON), effectiveDate)

	return entity.Record{}, err
}