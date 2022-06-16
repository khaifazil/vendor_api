package main

import (
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func main() {

	//testVoucher := `{"ID": "dsfadfa", "Amount": 2, "VendorUsed": "NTUC"}`
	//
	//if readV, err := readVoucher([]byte(testVoucher)); err != nil {
	//	ErrorLogger.Println("unable to reading voucher:", err)
	//} else {
	//	fmt.Printf("%+v\n", readV)
	//	validateVoucher(readV)
	//}

	router := mux.NewRouter()
	router.HandleFunc("/", home)
	router.HandleFunc("/vendorAPI/v1/process_voucher", processVoucher).Methods("POST")

	fmt.Println("Listening on port 8080")
	err := http.ListenAndServeTLS("localhost:8080", "./SSL/cert03.pem", "./SSL/key03.pem", router)
	if err != nil {
		ErrorLogger.Fatal("Error:", err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to Vendor API!")
}
