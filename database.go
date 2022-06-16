package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func openDatabase() *sql.DB {
	//Use mysql as driverName and a valid DSN as dataSourceName:
	db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/vendor_api_db")

	//handle error
	if err != nil {
		ErrorLogger.Fatalf("Unable to open database: %v", err)
	} else {
		fmt.Println("Database opened")
	}
	return db
}
