package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Voucher struct {
	VID         string
	CustomerID  string
	Amount      int
	MerchantID  string
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

			branchList.storeConsumed(voucher)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(voucher)
			//fmt.Printf("%+v", *merchantsList.Head)
		}
	}

}

func (d *doublyLinkedList) storeConsumed(v Voucher) { //TODO: add concurrency
	//Open database
	db := openDatabase()
	defer db.Close()
	defer fmt.Println("Database closed")

	if err := insertVoucherDB(v, db); err != nil {
		ErrorLogger.Panicf("unable to insert into database: %v", err)
	}

	//merchant, err := d.searchListForMerchant(v.MerchantID)
	//if err != nil {
	//	ErrorLogger.Println(err) //TODO: create new merchant if no merchant found
	//}

	//merchant.UnclaimedVouchers = append(merchant.UnclaimedVouchers, v)
	//merchant.AmountOwed += v.Amount
}
