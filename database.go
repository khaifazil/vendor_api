package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func openDatabase() *sql.DB {
	//Use mysql as driverName and a valid DSN as dataSourceName:
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/merchant_api_db")

	//handle error
	if err != nil {
		ErrorLogger.Fatalf("Unable to open database: %v", err)
	} else {
		fmt.Println("Database opened")
	}
	return db
}

func insertVoucherDB(v Voucher, db *sql.DB) error {
	query := fmt.Sprintf("INSERT INTO consumed_vouchers (Voucher_ID, Customer_ID, Merchant_name, Amount, Is_Consumed, Is_Claimed, Is_Validated) VALUES ('%s','%s','%s',%d,%v,%v,%v)", v.VID, v.CustomerID, v.BranchCode, v.Amount, v.IsConsumed, v.IsClaimed, v.IsValidated)
	//insert used voucher into table
	_, err := db.Query(query)
	if err != nil {
		return err
	}
	return nil
}

func insertNewMerchantDB(ID string, name string, db *sql.DB) error {
	query := fmt.Sprintf("INSERT INTO merchants (Merchant_ID, Name) VALUES ('%s','%s')", ID, name)
	//insert used voucher into table
	_, err := db.Query(query)
	if err != nil {
		return err
	}
	return nil
}

func insertNewBranchDB(code string, ID string, name string, db *sql.DB) error {
	query := fmt.Sprintf("INSERT INTO merchant_Branches (Branch_Code, Name, MerchantID) VALUES ('%s','%s','%s')", code, name, ID)
	//insert used voucher into table
	_, err := db.Query(query)
	if err != nil {
		return err
	}
	return nil
}

func merchantExistsName(db *sql.DB, name string) (bool, error) {
	sqlStmt := `SELECT name FROM merchants WHERE name = ?`
	err := db.QueryRow(sqlStmt, name).Scan(&name)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func merchantExistsID(db *sql.DB, ID string) (bool, error) {
	sqlStmt := `SELECT Merchant_ID FROM merchants WHERE Merchant_ID = ?`
	err := db.QueryRow(sqlStmt, ID).Scan(&ID)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func branchExists(db *sql.DB, code string) (bool, error) {
	sqlStmt := `SELECT Branch_Code FROM merchant_branches WHERE Branch_Code = ?`
	err := db.QueryRow(sqlStmt, code).Scan(&code)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
		return false, nil
	}
	return true, nil
}
