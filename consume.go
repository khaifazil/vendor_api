package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Voucher struct {
	ID         string
	Amount     int
	VendorUsed string
}

func readVoucher(data []byte) (Voucher, error) {
	newVoucher := Voucher{}
	if err := json.Unmarshal(data, &newVoucher); err != nil {
		return Voucher{}, err
	}
	return newVoucher, nil
}

func validateVoucher(w http.ResponseWriter, v Voucher) error {
	url := "https://localhost:5000/voucherAPI/validate"

	jsonValue, err := json.Marshal(v)
	if err != nil {
		ErrorLogger.Println("unable to marshal JSON: ", err)
		return err
	}

	_, err = http.Post(url, "application/json", bytes.NewReader(jsonValue))
	if err != nil {
		ErrorLogger.Println("Unable to send POST request:", err)
		return err
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
	return nil
}

func processVoucher(w http.ResponseWriter, r *http.Request) {

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Please supply voucher information in JSON format", http.StatusUnprocessableEntity)
		ErrorLogger.Println("unable to read request body:", err)
		return
	}

	voucher, err := readVoucher(reqBody)
	if err != nil {
		http.Error(w, "Please supply voucher information in JSON format", http.StatusUnprocessableEntity)
		ErrorLogger.Println("unable to read voucher:", err)
		return
	}

	err = validateVoucher(w, voucher)
	if err != nil {
		http.Error(w, "Unable to send voucher validation request", http.StatusUnprocessableEntity)
		ErrorLogger.Println("Unable to send voucher validation request:", err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Voucher sent for validation"))

}
