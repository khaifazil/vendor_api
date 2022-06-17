package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func openDatabase() *sql.DB {
	//Use mysql as driverName and a valid DSN as dataSourceName:
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/vendor_api_db")

	//handle error
	if err != nil {
		ErrorLogger.Fatalf("Unable to open database: %v", err)
	} else {
		fmt.Println("Database opened")
	}
	return db
}

func insertVoucherDB(v Voucher, db *sql.DB) error {
	query := fmt.Sprintf("INSERT INTO consumed_vouchers (Voucher_ID, CustomerID, Merchant_name, Amount, is_Consumed, is_Claimed, is_Validated) VALUES ('%s','%s','%s',%d,%v,%v,%v)", v.VID, v.CustomerID, v.MerchantID, v.Amount, v.IsConsumed, v.IsClaimed, v.IsValidated)
	//insert used voucher into table
	_, err := db.Query(query)
	if err != nil {
		return err
	}
	return nil
}
