package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

func openDatabase() *sql.DB {

	var (
		UserName = os.Getenv("DB_USER")
		Password = os.Getenv("DB_PW")
		DBIP     = os.Getenv("DB_IP")
		DBPort   = os.Getenv("DB_PORT")
		DBName   = os.Getenv("DB_NAME")
	)

	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", UserName, Password, DBIP, DBPort, DBName)

	//Use mysql as driverName and a valid DSN as dataSourceName:
	db, err := sql.Open("mysql", dsn)
	//handle error
	if err != nil {
		ErrorLogger.Fatalf("Unable to open database: %v", err)
	} else {
		fmt.Println("Database opened")
	}
	return db
}

func closeDatabase(db *sql.DB) {
	db.Close()
	fmt.Println("Database Closed")
}

func insertVoucherDB(v Voucher, db *sql.DB) error {
	//insert used voucher into table
	_, err := db.Exec("INSERT INTO consumed_vouchers (VID, Branch_ID, Customer_ID, Amount, Is_Consumed, Is_Claimed, Is_Validated) VALUES (?, ?,?,?,?,?,?)", v.VID, v.BranchID, v.CustomerID, v.Amount, v.IsConsumed, v.IsClaimed, v.IsValidated)
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

func insertNewBranchDB(branchID, code, merchantID, name string, db *sql.DB) error {
	//insert used voucher into table
	_, err := db.Exec("INSERT INTO merchant_Branches (Branch_ID, Branch_Code, Name, MerchantID) VALUES (?, ?, ?, ?)", branchID, code, name, merchantID)
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

func branchExists(db *sql.DB, target, field string) (bool, error) {
	sqlStmt := `SELECT ? FROM merchant_branches WHERE ? = ?`
	err := db.QueryRow(sqlStmt, field, field, target).Scan(&target)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
		return false, nil
	}
	return true, nil
}
