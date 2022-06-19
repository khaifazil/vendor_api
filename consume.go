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
	BranchCode  string
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
		http.Error(w, "Unable to read request body", http.StatusUnprocessableEntity)
		return
	}

	if r.Header.Get("Content-Type") == "application/json" {
		voucher, err := readVoucher(resp)
		if err != nil {
			ErrorLogger.Println("unable to marshal JSON: ", err)
			http.Error(w, "unable to marshal JSON", http.StatusUnprocessableEntity)
			return
		}

		if voucher.IsValidated == true {
			voucher.IsConsumed = true

			err := branchList.storeConsumed(voucher)
			if err != nil {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte(err.Error()))
				ErrorLogger.Println(err)
				return
			}

			db := openDatabase()
			defer db.Close()
			defer fmt.Println("Database Closed")

			db.Query("UPDATE merchant_branches SET Amount_owed = Amount_owed + ? WHERE Branch_Code = ?", voucher.Amount, voucher.BranchCode)

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
	b, exist := d.searchListForNode(v.BranchCode)
	if !exist { // if branch does not exist, check DB
		var (
			name       string
			merchantID string
			amount     int
		)
		//query row and returns data needed
		err := db.QueryRow("SELECT Name, MerchantID, Amount_owed FROM merchant_branches WHERE Branch_Code = ?", v.BranchCode).Scan(&name, &merchantID, &amount)
		if err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("unable to query db: %v", err)
			}
			return fmt.Errorf("branch does not exists in database: %v", err)
		}
		//creates new branch node
		newBranch := &branch{
			Code:              v.BranchCode,
			Name:              name,
			MerchantID:        merchantID,
			AmountOwed:        amount,
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
