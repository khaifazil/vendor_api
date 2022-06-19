package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"net/http"
)

//var wg sync.WaitGroup

func main() {

	//testVoucher := `{"ID": "dsfadfa", "Amount": 2, "VendorUsed": "NTUC"}`
	//
	//if readV, err := readVoucher([]byte(testVoucher)); err != nil {
	//	ErrorLogger.Println("unable to read voucher:", err)
	//} else {
	//	fmt.Printf("%+v\n", readV)
	//	validateVoucher(readV)
	//}

	//fmt.Println(genID("Sheng Siong"))

	router := mux.NewRouter()
	router.HandleFunc("/", home)
	//router.HandleFunc("/vendorAPI/v1/process_voucher", processVoucher).Methods("POST")
	router.HandleFunc("/api/v1/merchants/consume_voucher", consumeVoucher).Methods("POST")
	router.HandleFunc("/api/v1/merchants/", CreateMerchant).Methods("POST")
	router.HandleFunc("/api/v1/merchants/{merchantID}/branches", addBranches).Methods("POST")
	router.HandleFunc("/api/v1/merchants/", getAllMerchants).Methods("GET")

	fmt.Println("Listening on port 8080")
	err := http.ListenAndServeTLS("localhost:9091", "./SSL/localhost.cert.pem", "./SSL/localhost.key.pem", router)
	if err != nil {
		ErrorLogger.Fatal("Error:", err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to Vendor API!")
}
