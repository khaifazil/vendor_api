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

	//branchList.addEndNode(branch{
	//	BranchID:          "test",
	//	BranchCode:        "test",
	//	Name:              "test1",
	//	MerchantID:        "DSFSDF",
	//	AmountOwed:        4,
	//	UnclaimedVouchers: nil,
	//})
	//branchList.addEndNode(branch{
	//	BranchID:          "test",
	//	BranchCode:        "test",
	//	Name:              "test2",
	//	MerchantID:        "DSFSDF",
	//	AmountOwed:        4,
	//	UnclaimedVouchers: nil,
	//})
	//branchList.addEndNode(branch{
	//	BranchID:          "test",
	//	BranchCode:        "test",
	//	Name:              "test3",
	//	MerchantID:        "DSFSDF",
	//	AmountOwed:        4,
	//	UnclaimedVouchers: nil,
	//})
	//branchList.addEndNode(branch{
	//	BranchID:          "test",
	//	BranchCode:        "test",
	//	Name:              "test4",
	//	MerchantID:        "DSFSDF",
	//	AmountOwed:        4,
	//	UnclaimedVouchers: nil,
	//})
	//branchList.addEndNode(branch{
	//	BranchID:          "test",
	//	BranchCode:        "test",
	//	Name:              "test5",
	//	MerchantID:        "DSFSDF",
	//	AmountOwed:        4,
	//	UnclaimedVouchers: nil,
	//})

	//branchList.redeemVoucher()

	//var branchSlice []branch
	//c := make(chan branch)
	//go branchList.traverseForward(c)
	//go branchList.traverseBackwards(c)
	//
	//for elem := range c {
	//	branchSlice = append(branchSlice, elem)
	//}
	//
	//fmt.Println(branchSlice)

	router := mux.NewRouter()
	router.HandleFunc("/", home)
	//router.HandleFunc("/vendorAPI/v1/process_voucher", processVoucher).Methods("POST")
	router.HandleFunc("/api/v1/merchants/consume_voucher", consumeVoucher).Methods("POST")
	router.HandleFunc("/api/v1/merchants/", CreateMerchant).Methods("POST")
	router.HandleFunc("/api/v1/merchants/{merchantID}/branches", addBranches).Methods("POST")
	router.HandleFunc("/api/v1/merchants/{merchantID}", getMerchant).Methods("GET")
	router.HandleFunc("/api/v1/merchants/", getAllMerchants).Methods("GET")
	router.HandleFunc("/api/v1/merchants/{merchantID}/{branchID}", removeBranch).Methods("GET")
	router.HandleFunc("/api/v1/merchants/{merchantID}/toggle_active", updateMerchantIsActive).Methods("PUT")
	router.HandleFunc("/api/v1/merchants/redeem", branchList.claimVoucher).Methods("PUT")

	fmt.Println("Listening on port 9091")
	err := http.ListenAndServeTLS("localhost:9091", "./SSL/localhost.cert.pem", "./SSL/localhost.key.pem", router)
	if err != nil {
		ErrorLogger.Fatal("Error:", err)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to Vendor API!")
}
