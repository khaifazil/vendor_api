package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"net/http"
)

type Voucher struct {
	VID         string
	CustomerID  string
	Amount      int
	BranchID    string
	IsValidated bool
	IsConsumed  bool
	IsClaimed   bool
}

func readVoucher(data []byte) (Voucher, error) {
	newVoucher := Voucher{}
	if err := json.Unmarshal(data, &newVoucher); err != nil {
		return Voucher{}, err
	}
	return newVoucher, nil
}

func consumeVoucher(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("System recovered")
		}
	}()

	//check validation of apikey in header
	if !validateAPIKey(r) {
		errorResponse(w, "API key is unauthorized", http.StatusUnauthorized)
		ErrorLogger.Println("API key is unauthorized")
		return
	}

	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrorLogger.Println("Unable to read request body:", err)
		errorResponse(w, "Unable to read request body", http.StatusUnprocessableEntity)
		return
	}

	if r.Header.Get("Content-Type") == "application/json" {
		voucher, err := readVoucher(resp)
		if err != nil {
			ErrorLogger.Println("unable to marshal JSON: ", err)
			errorResponse(w, "Unable to marshal JSON", http.StatusUnprocessableEntity)
			return
		}

		db := openDatabase()
		defer closeDatabase(db)

		var isActive bool
		err = db.QueryRow("SELECT is_active FROM merchants AS m JOIN merchant_branches AS mb ON m.Merchant_ID = mb.MerchantID WHERE Branch_ID = ?;", voucher.BranchID).Scan(&isActive)
		if err != nil {
			errorResponse(w, "500 - unable to query database", http.StatusInternalServerError)
			ErrorLogger.Println("500 - unable to query database", err)
			return
		}

		if !isActive {
			errorResponse(w, "403 - Merchant is not active", http.StatusForbidden)
			ErrorLogger.Println("403 - Merchant is not active")
			return
		}

		if voucher.IsValidated == true {
			voucher.IsConsumed = true

			branchList.storeConsumed(voucher)

			db := openDatabase()
			defer db.Close()
			defer fmt.Println("Database Closed")

			//update merchant database amount owed
			db.Query("UPDATE merchant_branches SET Amount_owed = Amount_owed + ? WHERE Branch_ID = ?", voucher.Amount, voucher.BranchID)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(struct {
				Ok   bool    `json:"ok"`
				Msg  string  `json:"msg"`
				Data Voucher `json:"data"`
			}{Ok: true, Msg: "[MS-MERCHANTS]: consume voucher, successful", Data: voucher})
		}
	}

}

func (d *doublyLinkedList) storeConsumed(v Voucher) {

	//Open database
	db := openDatabase()
	defer db.Close()
	defer fmt.Println("Database closed")

	//check if branch is in linked-list
	wg.Add(1)
	go func() {
		b, exist := d.searchListForNode(v.BranchID)
		if !exist { // if branch does not exist, check DB
			var (
				name       string
				merchantID string
				amount     int
				branchCode string
			)
			//query row and returns data needed
			err := db.QueryRow("SELECT Name, MerchantID, Amount_owed, Branch_Code FROM merchant_branches WHERE Branch_ID = ?", v.BranchID).Scan(&name, &merchantID, &amount, &branchCode)
			if err != nil {
				if err != sql.ErrNoRows {
					ErrorLogger.Panicln("unable to query db: %v", err)
				}
				ErrorLogger.Panicln("unable to query db: %v", err)
			}
			//creates new branch node
			newBranch := &branch{
				BranchID:          v.BranchID,
				Name:              name,
				BranchCode:        branchCode,
				MerchantID:        merchantID,
				AmountOwed:        0,
				UnclaimedVouchers: nil,
			}
			newBranch.UnclaimedVouchers = append(newBranch.UnclaimedVouchers, v)
			newBranch.AmountOwed += v.Amount
			branchList.addEndNode(*newBranch) //adds to linked list
		} else {
			b.Data.UnclaimedVouchers = append(b.Data.UnclaimedVouchers, v)
			b.Data.AmountOwed += v.Amount
		}
		wg.Done()
	}()
	//stores voucher into DB
	wg.Add(1)
	go func() {
		if err := insertVoucherDB(v, db); err != nil {
			ErrorLogger.Panicf("unable to insert into database: %v", err)
		}
		wg.Done()
	}()
	wg.Wait()
}
