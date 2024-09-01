package storageServices

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/rainbowmga/timetravel/service"
)

func GetStorageType() string{

	storageType := flag.String( "storage_type", "sqlite", "Options are: ['sqlite', 'memory']" )
	flag.Parse()
	if *storageType != "memory" && *storageType != "sqlite"{
		log.Fatalf("Invalid Storage Type. Options are: ['sqlite', 'memory']")
	}
	
	fmt.Printf("Selected Storage type is %s\n", *storageType)
	return *storageType
}

func BuildStorageService(storageType string) (service.RecordService, *sql.DB){
	var storageService service.RecordService
	var db *sql.DB

	if storageType == "sqlite"{
		db, err := initDB()
		if err != nil {
			log.Fatal((err))
		}
		storageService = NewSqliteRecordService(db)
	} else
	{
		storageService = NewInMemoryRecordService()
	}
	return storageService, db
}