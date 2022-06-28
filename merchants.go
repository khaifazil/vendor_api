package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type MerchantData struct {
	MerchantID   string `json:"merchantID"`
	MerchantName string `json:"merchantName"`
	IsActive     bool   `json:"isActive"`
}

type data struct {
	MerchantID   string   `json:"merchantID"`
	MerchantName string   `json:"merchantName"`
	Branches     []branch `json:"branches"`
}

type branch struct {
	BranchID          string    `json:"branchID"`
	Name              string    `json:"name"`
	BranchCode        string    `json:"branchCode"`
	MerchantID        string    `json:"merchantID"`
	AmountOwed        int       `json:"amountOwed"`
	UnclaimedVouchers []Voucher `json:"unclaimedVouchers"`
}

var branchList doublyLinkedList

func genID(name string) string {
	tn := time.Now().Format("20060102150405")
	name = strings.ReplaceAll(strings.ToUpper(name), " ", "")
	return tn + name
}

func CreateMerchant(w http.ResponseWriter, r *http.Request) {

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
		errorResponse(w, "Send in merchant details in JSON format", http.StatusUnprocessableEntity)
		ErrorLogger.Println("unable to read request body:", err)
		return
	}

	var result map[string]interface{}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		errorResponse(w, "Send in merchant details in JSON format", http.StatusUnprocessableEntity)
		ErrorLogger.Println("unable to read request body:", err)
		return
	}

	db := openDatabase()
	defer db.Close()
	defer fmt.Println("Database closed")

	//check for duplicates
	exists, err := merchantExistsName(db, result["name"].(string))
	if err != nil {
		errorResponse(w, "unable to query database", http.StatusInternalServerError)
		ErrorLogger.Panicln("unable to query database:", err)
	}

	if exists {
		errorResponse(w, "409 - Duplicate Merchant Name", http.StatusConflict)
		ErrorLogger.Println("409 - Duplicate Merchant Name")
		return
	}

	branches := result["branches"]

	//generate merchantID
	newID := genID(result["name"].(string))

	for _, b := range branches.([]interface{}) {
		branchID := newID + "-" + b.(map[string]interface{})["code"].(string)
		//check for duplicates
		exists, err = branchExists(db, branchID, "Branch_ID")
		if err != nil {
			errorResponse(w, "unable to query database:", http.StatusInternalServerError)
			ErrorLogger.Panicln("unable to query database:", err)
		}
		if exists {
			errorResponse(w, "409 - Duplicate Branch Code", http.StatusConflict)
			ErrorLogger.Println("409 - Duplicate Branch code")
			return
		}
	}

	//save merchant to database
	if err = insertNewMerchantDB(newID, result["name"].(string), db); err != nil {
		errorResponse(w, "unable to query database", http.StatusInternalServerError)
		ErrorLogger.Panicf("unable to insert new Merchant:", err)
	}

	//loop over slices
	var temp []branch
	for _, b := range branches.([]interface{}) {
		code := b.(map[string]interface{})["code"].(string)
		name := b.(map[string]interface{})["name"].(string)
		branchID := newID + "-" + code

		//save branch to database
		if err = insertNewBranchDB(branchID, code, newID, name, db); err != nil {
			errorResponse(w, "unable to query database", http.StatusInternalServerError)
			ErrorLogger.Panicf("unable to insert new branch:", err)
		}
		//save branch into linked list
		newBranch := branch{
			BranchID:          branchID,
			Name:              name,
			BranchCode:        code,
			MerchantID:        newID,
			AmountOwed:        0,
			UnclaimedVouchers: nil,
		}
		//branchList.addEndNode(newBranch)
		temp = append(temp, newBranch)
	}

	//send post request to web portal
	reply := struct {
		Ok   bool   `json:"ok"`
		Msg  string `json:"msg"`
		Data data   `json:"data"`
	}{
		Ok:  true,
		Msg: "[MS-MERCHANTS]: created merchant data, successful",
		Data: data{
			MerchantID:   newID,
			MerchantName: result["name"].(string),
			Branches:     temp,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reply)

}

func getMerchant(w http.ResponseWriter, r *http.Request) { //TODO: retrieve branches as well

	//check validation of apikey in header
	if !validateAPIKey(r) {
		errorResponse(w, "API key is unauthorized", http.StatusUnauthorized)
		ErrorLogger.Println("API key is unauthorized")
		return
	}

	var (
		merchantID string
		name       string
		isActive   bool
	)

	params := mux.Vars(r)
	ID := params["merchantID"]
	fmt.Println(ID)

	db := openDatabase()
	defer db.Close()
	defer fmt.Println("Database closed")

	err := db.QueryRow("SELECT * FROM merchants WHERE Merchant_ID = ?", ID).Scan(&merchantID, &name, &isActive)
	if err != nil {
		if err != sql.ErrNoRows {
			errorResponse(w, "unable to query database", http.StatusInternalServerError)
			ErrorLogger.Println("500 - unable to query database", err)
			return
		}
		errorResponse(w, "404 - Merchant ID not found in database", http.StatusNotFound)
		ErrorLogger.Println("404 - Merchant ID not found in database", err)
		return
	}

	result := MerchantData{
		MerchantID:   merchantID,
		MerchantName: name,
		IsActive:     isActive,
	}

	reply := struct {
		OK   bool         `json:"ok"`
		Msg  string       `json:"msg"`
		Data MerchantData `json:"data"`
	}{true, "[MS-MERCHANTS]: retrieved merchant data, successful", result}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reply)
}

func getAllMerchants(w http.ResponseWriter, r *http.Request) {

	//check validation of apikey in header
	if !validateAPIKey(r) {
		errorResponse(w, "API key is unauthorized", http.StatusUnauthorized)
		ErrorLogger.Println("API key is unauthorized")
		return
	}

	//process url queries
	values := r.URL.Query()

	pageIndex, err := strconv.Atoi(values.Get("page_index"))
	if err != nil {
		errorResponse(w, "URL query not found", http.StatusBadRequest)
		ErrorLogger.Println("400 - URL query not found")
		return
	}
	recordsPerPage, err := strconv.Atoi(values.Get("records_per_page"))
	if err != nil {
		errorResponse(w, "URL query not found", http.StatusBadRequest)
		ErrorLogger.Println("400 - URL query not found")
		return
	}

	db := openDatabase()
	defer db.Close()

	rows, err := db.Query("SELECT COUNT(*) FROM merchants")
	if err != nil {
		errorResponse(w, "unable to query database", http.StatusInternalServerError)
		ErrorLogger.Println("unable to scan rows", err)
		return
	}
	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(count)

	rows, err = db.Query("SELECT * FROM merchants")
	if err != nil {
		errorResponse(w, "unable to query database", http.StatusInternalServerError)
		ErrorLogger.Println("unable to scan rows", err)
		return
	}
	defer rows.Close()

	var (
		merchant_ID   string
		merchantName  string
		isActive      bool
		branchID      string
		branchName    string
		branchCode    string
		amountOwed    int
		amountClaimed int
	)

	type tempBranch struct {
		BranchID      string `json:"branchID"`
		Name          string `json:"name"`
		BranchCode    string `json:"branchCode"`
		MerchantID    string `json:"merchantID"`
		AmountOwed    int    `json:"amountOwed"`
		AmountClaimed int    `json:"amountClaimed"`
	}

	type merchantAndBranch struct {
		MerchantID   string       `json:"merchantID"`
		MerchantName string       `json:"merchantName"`
		IsActive     bool         `json:"isActive"`
		Branches     []tempBranch `json:"branches"`
	}

	var merchantsAndBranches []merchantAndBranch

	wg.Add(count)
	go func() {
		for rows.Next() {

			if err = rows.Scan(&merchant_ID, &merchantName, &isActive); err != nil {
				errorResponse(w, "unable to query database", http.StatusInternalServerError)
				ErrorLogger.Println("unable to scan rows", err)
				return
			}

			branch, err := db.Query("SELECT * FROM merchant_branches WHERE MerchantID = ?", merchant_ID)
			if err != nil {
				errorResponse(w, "unable to query database", http.StatusInternalServerError)
				ErrorLogger.Println("unable to scan rows", err)
				return
			}
			defer branch.Close()

			var branchSlice []tempBranch

			for branch.Next() {
				if err = branch.Scan(&branchID, &branchName, &branchCode, &merchant_ID, &amountOwed, &amountClaimed); err != nil {
					errorResponse(w, "unable to query database", http.StatusInternalServerError)
					ErrorLogger.Println("unable to scan rows", err)
					return
				}

				tempBranch := tempBranch{
					BranchID:      branchID,
					Name:          branchName,
					BranchCode:    branchCode,
					MerchantID:    merchant_ID,
					AmountOwed:    amountOwed,
					AmountClaimed: amountClaimed,
				}

				branchSlice = append(branchSlice, tempBranch)
			}

			temp := merchantAndBranch{
				MerchantID:   merchant_ID,
				MerchantName: merchantName,
				IsActive:     isActive,
				Branches:     branchSlice,
			}
			merchantsAndBranches = append(merchantsAndBranches, temp)
			wg.Done()

		}
	}()
	wg.Wait()

	if merchantsAndBranches == nil {
		errorResponse(w, "there are no merchants in database", http.StatusBadRequest)
		ErrorLogger.Println("there are no merchants in database", err)
		return
	}

	var sliceToSend []merchantAndBranch
	if (pageIndex * recordsPerPage) > len(merchantsAndBranches) {
		errorResponse(w, "Page_index is out of bounds of total pages", http.StatusBadRequest)
		ErrorLogger.Println("400 - Page_index is out of bounds of total pages")
		return
	} else if (pageIndex*recordsPerPage)+recordsPerPage > len(merchantsAndBranches) {
		sliceToSend = merchantsAndBranches[(pageIndex * recordsPerPage):]
	} else {
		sliceToSend = merchantsAndBranches[(pageIndex * recordsPerPage) : (pageIndex*recordsPerPage)+recordsPerPage]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Ok   bool                `json:"ok"`
		Msg  string              `json:"msg"`
		Data []merchantAndBranch `json:"data"`
	}{true, "[MS-MERCHANTS]: retrieval of list of merchants, successful", sliceToSend})
}

func addBranches(w http.ResponseWriter, r *http.Request) {

	//check validation of apikey in header
	if !validateAPIKey(r) {
		errorResponse(w, "API key is unauthorized", http.StatusUnauthorized)
		ErrorLogger.Println("API key is unauthorized")
		return
	}

	params := mux.Vars(r)
	var result map[string]interface{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResponse(w, "422 - Unable to read request body", http.StatusUnprocessableEntity)
		ErrorLogger.Println("Unable to read request body:", err)
		return
	}

	if err := json.Unmarshal(body, &result); err != nil {
		errorResponse(w, "422 - Send in branch details in JSON format", http.StatusUnprocessableEntity)
		ErrorLogger.Println("unable to unmarshal JSON:", err)
		return
	}

	db := openDatabase()
	defer db.Close()
	defer fmt.Println("Database closed")

	//check if merchant exist in database
	exist, err := merchantExistsID(db, params["merchantID"])
	if err != nil {
		errorResponse(w, "500 - unable to query database", http.StatusInternalServerError)
		ErrorLogger.Println("unable to query database", err)
		return
	}

	//if does not exist return error
	if !exist {
		errorResponse(w, "404 - Merchant ID does not exist in database", http.StatusNotFound)
		ErrorLogger.Println("Merchant ID does not exist in database")
		return
	}
	//if exists check for branch duplicates
	branches := result["branches"]
	for _, b := range branches.([]interface{}) {

		//check for duplicates
		exist, err := branchExists(db, b.(map[string]interface{})["code"].(string), "Branch_Code")
		if err != nil {
			errorResponse(w, "500 - unable to query database:", http.StatusInternalServerError)
			ErrorLogger.Panicln("unable to query database:", err)
		}
		if exist {
			errorResponse(w, "409 - Duplicate Branch Code", http.StatusConflict)
			ErrorLogger.Println("409 - Duplicate Branch code")
			return
		}
	}
	//if no duplicates create branches
	var temp []branch
	for _, b := range branches.([]interface{}) {
		code := b.(map[string]interface{})["code"].(string)
		name := b.(map[string]interface{})["name"].(string)
		branchID := params["merchantID"] + "-" + code

		//save branch to database
		if err = insertNewBranchDB(branchID, code, params["merchantID"], name, db); err != nil {
			errorResponse(w, "unable to query database", http.StatusInternalServerError)
			ErrorLogger.Panicf("unable to insert new branch:", err)
		}
		//save branch into linked list
		newBranch := branch{
			BranchID:          branchID,
			Name:              name,
			BranchCode:        code,
			MerchantID:        params["merchantID"],
			AmountOwed:        0,
			UnclaimedVouchers: nil,
		}
		//branchList.addEndNode(newBranch)
		temp = append(temp, newBranch)
	}

	//send post request to web portal
	reply := struct {
		OK   bool     `json:"ok"`
		Msg  string   `json:"msg"`
		Data []branch `json:"data"`
	}{true, "[MS-MERCHANTS]: branches added to merchant data, successful", temp}

	//send back results
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reply)
}

func removeBranch(w http.ResponseWriter, r *http.Request) {

	//check validation of apikey in header
	if !validateAPIKey(r) {
		errorResponse(w, "API key is unauthorized", http.StatusUnauthorized)
		ErrorLogger.Println("API key is unauthorized")
		return
	}

	params := mux.Vars(r)

	merchantID := params["merchantID"]
	branchID := params["branchID"]

	var (
		name       string
		amountOwed int
		branchCode string
	)

	db := openDatabase()
	defer db.Close()
	defer fmt.Println("Database closed")

	//check if merchantID exists
	exists, err := merchantExistsID(db, merchantID)
	if err != nil {
		errorResponse(w, "500 - unable to query database", http.StatusInternalServerError)
		ErrorLogger.Println("unable to query database", err)
		return
	}

	if !exists {
		errorResponse(w, "404 - Merchant ID not found in database", http.StatusNotFound)
		ErrorLogger.Println("404 - Merchant ID not found in database")
		return
	}
	//check if branch exists
	err = db.QueryRow("SELECT Name, Amount_owed, Branch_Code FROM merchant_branches WHERE Branch_ID = ?", branchID).Scan(&name, &amountOwed, &branchCode)
	if err != nil {
		if err != sql.ErrNoRows {
			errorResponse(w, "500 - unable to query database", http.StatusInternalServerError)
			ErrorLogger.Println("500 - unable to query database", err)
			return
		}
		errorResponse(w, "404 - Branch not found in database", http.StatusNotFound)
		ErrorLogger.Println("404 - Branch not found in database", err)
		return
	}
	//check if there is outstanding balance for the branch
	//if balance is not 0 reject
	if amountOwed > 0 {
		errorResponse(w, "403 - Request unsuccessful, there are unclaimed funds tied to branch", http.StatusForbidden)
		ErrorLogger.Println("403 - Request unsuccessful, there are unclaimed funds tied to branch")
		return
	}
	//if all checks pass delete branch from database
	_, err = db.Exec("DELETE FROM merchant_branches WHERE Branch_ID = ?", branchID)

	deletedBranch := branch{
		BranchID:          branchID,
		Name:              name,
		BranchCode:        branchCode,
		MerchantID:        merchantID,
		AmountOwed:        amountOwed,
		UnclaimedVouchers: nil,
	}

	reply := struct {
		Ok   bool   `json:"ok"`
		Msg  string `json:"msg"`
		Data branch `json:"data"`
	}{true, "[MS-MERCHANTS]: Branch removed from merchant data, successful", deletedBranch}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(202)
	json.NewEncoder(w).Encode(reply)
}

func updateMerchantIsActive(w http.ResponseWriter, r *http.Request) {

	//check validation of apikey in header
	if !validateAPIKey(r) {
		errorResponse(w, "API key is unauthorized", http.StatusUnauthorized)
		ErrorLogger.Println("API key is unauthorized")
		return
	}

	params := mux.Vars(r)
	merchantID := params["merchantID"]

	var (
		name     string
		isActive bool
	)

	db := openDatabase()
	defer closeDatabase(db)

	err := db.QueryRow("SELECT Name, is_active FROM merchants WHERE Merchant_ID = ?", merchantID).Scan(&name, &isActive)
	if err != nil {
		if err != sql.ErrNoRows {
			errorResponse(w, "500 - unable to query database", http.StatusInternalServerError)
			ErrorLogger.Println("500 - unable to query database", err)
			return
		}
		errorResponse(w, "404 - Merchant ID not found in database", http.StatusNotFound)
		ErrorLogger.Println("404 - Merchant ID not found in database")
		return
	}

	var msg string
	if isActive {
		_, err = db.Exec("UPDATE merchants SET is_active = FALSE WHERE Merchant_ID = ?", merchantID)
		if err != nil {
			errorResponse(w, "500 - unable to query database", http.StatusInternalServerError)
			ErrorLogger.Println("500 - unable to query database", err)
			return
		}
		isActive = false
		msg = "[MS-MERCHANTS]: deactivated merchant data, successful"
	} else {
		_, err = db.Exec("UPDATE merchants SET is_active = TRUE WHERE Merchant_ID = ?", merchantID)
		if err != nil {
			errorResponse(w, "500 - unable to query database", http.StatusInternalServerError)
			ErrorLogger.Println("500 - unable to query database", err)
			return
		}
		isActive = true
		msg = "[MS-MERCHANTS]: activated merchant data, successful"
	}

	m := MerchantData{
		MerchantID:   merchantID,
		MerchantName: name,
		IsActive:     isActive,
	}
	reply := struct {
		Ok   bool         `json:"ok"`
		Msg  string       `json:"msg"`
		Data MerchantData `json:"data"`
	}{true, msg, m}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reply)
}
