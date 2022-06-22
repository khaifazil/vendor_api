package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"net/http"
	"sync"
)

var wg sync.WaitGroup

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/", home)
	//router.HandleFunc("/vendorAPI/v1/process_voucher", processVoucher).Methods("POST")
	router.HandleFunc("/api/v1/merchants/consume_voucher", consumeVoucher).Methods("POST")
	router.HandleFunc("/api/v1/merchants/", CreateMerchant).Methods("POST")
	router.HandleFunc("/api/v1/merchants/{merchantID}/branches", addBranches).Methods("POST")
	router.HandleFunc("/api/v1/merchants/{merchantID}", getMerchant).Methods("GET")
	router.HandleFunc("/api/v1/merchants/", getAllMerchants).Methods("GET")
	router.HandleFunc("/api/v1/merchants/{merchantID}/{branchID}/remove", removeBranch).Methods("GET")
	router.HandleFunc("/api/v1/merchants/{merchantID}/toggle_active", updateMerchantIsActive).Methods("PUT")
	router.HandleFunc("/api/v1/merchants/get_vouchers/total", branchList.totalUnclaimedVoucher).Methods("GET")
	router.HandleFunc("/api/v1/merchants/get_vouchers/", branchList.sendVouchers).Methods("GET")
	router.HandleFunc("/api/v1/merchants/claim_vouchers", claimVoucher).Methods("PUT")
	router.HandleFunc("/api/v1/merchants/reload", reloadLocalCache).Methods("PUT")

	fmt.Println("Listening on port 9091")
	err := http.ListenAndServeTLS("localhost:9091", "./SSL/localhost.cert.pem", "./SSL/localhost.key.pem", router)
	if err != nil {
		ErrorLogger.Fatal("Error:", err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to Vendor API!")
}
