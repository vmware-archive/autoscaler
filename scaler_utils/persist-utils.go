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
	dbEnabled bool = true
	TABLE_NAME = "AppDetail"
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
	
	if (err != nil) {
		fmt.Println("Error: ", err)
		dbEnabled = false
		return
	}
		
	fmt.Printf("Persistence: Vcap Services: %v\n", vcapServices.MySqlServiceDefn)
	
    if (len(vcapServices.MySqlServiceDefn) == 0 ) {
   		dbEnabled = false
		return
	}
	
	mysqlServiceDefn := vcapServices.MySqlServiceDefn[0]
	
	dbString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",				
	            mysqlServiceDefn.Credentials.Username,
				mysqlServiceDefn.Credentials.Password,
				mysqlServiceDefn.Credentials.Hostname,
				mysqlServiceDefn.Credentials.Port,
				mysqlServiceDefn.Credentials.Name)
	
//	if (dbType == "MongoDb") {
//		dbString = vcapServices.MongoDbServiceDefn[0].Credentials.Uri
//	} 
	InitDB(dbString)
}

func IsDBEnabled() bool {
	return dbEnabled
}

func enableDB() {
	dbEnabled = true
	fmt.Printf("Persistence: MySQL Persistence Enabled\n")
}


func InitDB(dbConnectionInfo string) {
	
	fmt.Printf("Persistence: Connecting to DB using parameters: %s", dbConnectionInfo)
	
	db, err := sql.Open("mysql", dbConnectionInfo)
    if err != nil {
   		log.Fatalln("\nFatal!! Error connecting to DB provided: " + dbConnectionInfo, err)
   	}

	// construct a gorp DbMap
	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	// JSON struct seems to allow null and that has problems with setKeys..
	//table := dbmap.AddTableWithName(AppDetail{}, "appdetails").SetKeys(true, "Target")
	dbmap.AddTable(AppDetail{})	
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")
		
	enableDB()	
}

func CloseDB() {
	if (dbmap != nil ) {
		dbmap.Db.Close()
	}
} 


func checkErr(err error, msg string) {
    if err != nil {
        log.Fatalln(msg, err)
    }
}

func Get(target string) AppDetail {
	var nilApp AppDetail
	if dbmap == nil {
		log.Fatalln("Persistence: Call InitDb() first!")
		return nilApp
	}
	
	//appDetail, err := dbmap.Get(AppDetail{}, target)
	var appDetail AppDetail
	_, err := dbmap.Select(&appDetail, "select * from " + TABLE_NAME + " where trader = ?", target)
	if err != nil {
		log.Fatalln("Persistence: Error selecting AppDetail with target name: %s from DB, %s", target, err)
		return nilApp
	}
	return appDetail
}

func Load() []AppDetail {
	var nilApps []AppDetail
	if dbmap == nil {
		log.Fatalln("Persistence: Call InitDb() first!")
		return nilApps
	}
	
	var appDetails []AppDetail
	_, err := dbmap.Select(&appDetails, "select * from " + TABLE_NAME + " LIMIT 50")
	if err != nil {
		log.Fatalln("Persistence: Error loading AppDetail from DB: %s", err)
		return nilApps
	}
	return appDetails
}

func Insert(appDetail AppDetail) {
	if dbmap == nil {
		log.Fatalln("Persistence: Call InitDb() first!")
		return
	}
	
	appDetail.Updated = time.Now().Unix() 
	err := dbmap.Insert(&appDetail)
	if err != nil {
		log.Fatalln("Persistence: Error persisting AppDetail: %v into DB, %s", appDetail, err)
	}
}

func Update(appDetail AppDetail) {
	if dbmap == nil {
		log.Fatalln("Persistence: Call InitDb() first!")
		return
	}
	
	appDetail.Updated = time.Now().Unix() 
	_, err := dbmap.Update(appDetail)
	if err != nil {
		log.Fatalln("Persistence: Error updating AppDetail: %v into DB, %s", appDetail, err)
	}
}


func Delete(appDetail AppDetail) {
	if dbmap == nil {
		log.Fatalln("\nCall InitDb() first!")
		return
	}
	
	_, err := dbmap.Delete(appDetail)
	if err != nil {
		log.Fatalln("Persistence: Error deleting AppDetail: %v from DB, %s", appDetail, err)
	}
}

func DeleteAll() error {
	if dbmap == nil {
		log.Fatalln("Persistence: Call InitDb() first!")
	}
	return dbmap.TruncateTables()
}

