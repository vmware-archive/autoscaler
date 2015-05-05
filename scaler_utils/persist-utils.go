package scaler_utils

import (
	"fmt"
	"log"
	"time"
	"os"
	"encoding/json"
	"database/sql"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
)

// Code referenced from
// http://nathanleclaire.com/blog/2013/11/04/want-to-work-with-databases-in-golang-lets-try-some-gorp/

var ( 
	dbEnabled bool
	dbType = "MySQL"
	dbmap *gorp.DbMap
	vcap_services string
)

func init() {
	vcap_services = os.Getenv("VCAP_SERVICES")
	if vcap_services == "" {
		dbEnabled = false
		return
	}
	var jsonBlob = []byte(vcap_services)
	var vcapServices VcapServices
	err := json.Unmarshal(jsonBlob, &vcapServices)
	
	//var creds Credentials
 	//err := json.Unmarshal(jsonBlob, &creds)
	if (err != nil) {
		fmt.Println("Error: ", err)
	}
	//fmt.Printf("Unmarshalled: %#v\n", creds)
	fmt.Printf("Vcap Services: %#v\n", vcapServices)
	
	dbString := vcapServices.MySqlServiceDefn[0].Credentials.Uri
	if (dbType == "MongoDb") {
		dbString = vcapServices.MongoDbServiceDefn[0].Credentials.Uri
	} 
	InitDB(dbString)
	
	if (dbmap != nil) {
		dbEnabled = true
		defer CloseDB()
	}
}

func DBEnabled() bool {
	return dbEnabled
}

func InitDB(dbConnectionInfo string) {
	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	db, err := sql.Open("mysql", dbConnectionInfo)
    if err != nil {
   		log.Fatalln("\nFatal!! Error connecting to DB provided: " + dbConnectionInfo, err)
   	}

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	table := dbmap.AddTable(AppDetail{}).SetKeys(true, "Target")
	fmt.Printf("Created Table: %#v\n", table)
}

func CloseDB() {
	if (dbmap != nil) {
		dbmap.Db.Close()
	}
}

func Insert(appDetail AppDetail) {
	if dbmap == nil {
		log.Fatalln("\nCall InitDb() first!")
		return
	}
	
	appDetail.Updated = time.Now().Unix() 
	err := dbmap.Insert(appDetail)
	if err != nil {
		log.Fatalln("Error persisting AppDetail: %v into DB, %s", appDetail, err)
	}
}

func Update(appDetail AppDetail) {
	if dbmap == nil {
		log.Fatalln("\nCall InitDb() first!")
		return
	}
	
	appDetail.Updated = time.Now().Unix() 
	_, err := dbmap.Update(appDetail)
	if err != nil {
		log.Fatalln("Error updating AppDetail: %v into DB, %s", appDetail, err)
	}
}


func Delete(appDetail AppDetail) {
	if dbmap == nil {
		log.Fatalln("\nCall InitDb() first!")
		return
	}
	
	_, err := dbmap.Delete(appDetail)
	if err != nil {
		log.Fatalln("Error deleting AppDetail: %v from DB, %s", appDetail, err)
	}
}

func DeleteAll() error {
	if dbmap == nil {
		log.Fatalln("\nCall InitDb() first!")
	}
	return dbmap.TruncateTables()
}

