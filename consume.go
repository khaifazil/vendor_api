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

//func validateVoucher(w http.ResponseWriter, v Voucher) error {
//	url := "https://localhost:5000/voucherAPI/validate"
//
//	jsonValue, err := json.Marshal(v)
//	if err != nil {
//		ErrorLogger.Println("unable to marshal JSON: ", err)
//		return err
//	}
//
//	_, err = http.Post(url, "application/json", bytes.NewReader(jsonValue))
//	if err != nil {
//		return err
//	} else {
//		w.WriteHeader(http.StatusAccepted)
//	}
//	return nil
//}

//func processVoucher(w http.ResponseWriter, r *http.Request) {
//
//	reqBody, err := ioutil.ReadAll(r.Body)
//	if err != nil {
//		http.Error(w, "Please supply voucher information in JSON format", http.StatusUnprocessableEntity)
//		ErrorLogger.Println("unable to read request body:", err)
//		return
//	}
//
//	voucher, err := readVoucher(reqBody)
//	if err != nil {
//		http.Error(w, "Please supply voucher information in JSON format", http.StatusUnprocessableEntity)
//		ErrorLogger.Println("unable to read voucher:", err)
//		return
//	}
//
//	err = validateVoucher(w, voucher)
//	if err != nil {
//		http.Error(w, "Unable to send voucher validation request", http.StatusUnprocessableEntity)
//		ErrorLogger.Println("Unable to send voucher validation POST request:", err)
//		return
//	}
//	w.WriteHeader(http.StatusAccepted)
//	w.Write([]byte("Voucher sent for validation"))
//}

func consumeVoucher(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("System recovered")
		}
	}()

	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ErrorLogger.Println("Unable to read request body:", err)
		http.Error(w, "Unable to read request body", http.StatusUnprocessableEntity) //TODO: change error
		return
	}

	if r.Header.Get("Content-Type") == "application/json" {
		voucher, err := readVoucher(resp)
		if err != nil {
			ErrorLogger.Println("unable to marshal JSON: ", err)
			http.Error(w, "unable to marshal JSON", http.StatusUnprocessableEntity) //TODO: change error
			return
		}

		db := openDatabase()
		defer closeDatabase(db)

		var isActive bool
		err = db.QueryRow("SELECT is_active FROM merchants AS m JOIN merchant_branches AS mb ON m.Merchant_ID = mb.MerchantID WHERE Branch_ID = ?;", voucher.BranchID).Scan(&isActive)
		if err != nil {
			http.Error(w, "500 - unable to query database", http.StatusInternalServerError) //TODO: change error
			ErrorLogger.Println("500 - unable to query database", err)
			return
		}

		if !isActive {
			http.Error(w, "403 - Merchant is not active", http.StatusForbidden) //TODO: change error
			ErrorLogger.Println("403 - Merchant is not active")
			return
		}

		if voucher.IsValidated == true {
			voucher.IsConsumed = true

			err := branchList.storeConsumed(voucher)
			if err != nil {
				w.WriteHeader(http.StatusUnprocessableEntity) //TODO: change error
				w.Write([]byte(err.Error()))
				ErrorLogger.Println(err)
				return
			}

			db := openDatabase()
			defer db.Close()
			defer fmt.Println("Database Closed")

			//update merchant database amount owed
			db.Query("UPDATE merchant_branches SET Amount_owed = Amount_owed + ? WHERE Branch_ID = ?", voucher.Amount, voucher.BranchID)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(struct {
				Ok   bool
				Msg  string
				Data Voucher
			}{Ok: true, Msg: "[MS-MERCHANTS]: consume voucher, successful", Data: voucher})
		}
	}

}

func (d *doublyLinkedList) storeConsumed(v Voucher) error { //TODO: add concurrency

	//Open database
	db := openDatabase()
	defer db.Close()
	defer fmt.Println("Database closed")

	//check if branch is in linked-list
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
				return fmt.Errorf("unable to query db: %v", err)
			}
			return fmt.Errorf("branch does not exists in database: %v", err)
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
	//stores voucher into DB
	if err := insertVoucherDB(v, db); err != nil {
		ErrorLogger.Panicf("unable to insert into database: %v", err)
	}

	return nil
}
