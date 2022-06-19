package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type MerchantData struct {
	MerchantID   string
	MerchantName string
	IsActive     bool
}

type FormData struct {
	Name     string   `json:"name"`
	Branches []branch `json:"branches"`
	IsActive bool     `json:"is_active"`
}

type JsonReply struct {
	Ok   bool   `json:"ok"`
	Msg  string `json:"msg"`
	Data data   `json:"data"`
}

type data struct {
	MerchantID   string   `json:"merchantID"`
	MerchantName string   `json:"merchantName"`
	Branches     []branch `json:"branches"`
}

type branch struct {
	Code              string `json:"code"`
	Name              string `json:"name"`
	MerchantID        string
	AmountOwed        int
	UnclaimedVouchers []Voucher
}

var branchList doublyLinkedList

func init() {

}

func (d *doublyLinkedList) searchListForBranch(target string) (*branch, bool) {
	currentNode := d.Head
	for currentNode != nil {

		if currentNode.Data.Code == target {
			return &currentNode.Data, true
		}
		currentNode = currentNode.Next
	}
	return &branch{}, false
}

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

	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Send in merchant details in JSON format", http.StatusUnprocessableEntity)
		ErrorLogger.Println("unable to read request body:", err)
		return
	}

	var result map[string]interface{}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		http.Error(w, "Send in merchant details in JSON format", http.StatusUnprocessableEntity)
		ErrorLogger.Println("unable to unmarshal JSON:", err)
		return
	}

	db := openDatabase()
	defer db.Close()
	defer fmt.Println("Database closed")

	//check for duplicates
	exists, err := merchantExistsName(db, result["name"].(string))
	if err != nil {
		http.Error(w, "unable to query database:", http.StatusInternalServerError)
		ErrorLogger.Panicln("unable to query database:", err)
	}

	if exists {
		http.Error(w, "409 - Duplicate Merchant Name", http.StatusConflict)
		ErrorLogger.Println("409 - Duplicate Merchant Name")
		return
	}

	branches := result["branches"]

	for _, b := range branches.([]interface{}) {

		//check for duplicates
		exists, err = branchExists(db, b.(map[string]interface{})["code"].(string))
		if err != nil {
			http.Error(w, "unable to query database:", http.StatusInternalServerError)
			ErrorLogger.Panicln("unable to query database:", err)
		}
		if exists {
			http.Error(w, "409 - Duplicate Branch Code", http.StatusConflict)
			ErrorLogger.Println("409 - Duplicate Branch code")
			return
		}
	}

	//generate merchantID
	newID := genID(result["name"].(string))

	//open database

	//save merchant to database
	if err = insertNewMerchantDB(newID, result["name"].(string), db); err != nil {
		ErrorLogger.Panicf("unable to insert new Merchant:", err)
	}

	//loop over slices
	var temp []branch
	for _, b := range branches.([]interface{}) {
		code := b.(map[string]interface{})["code"]
		name := b.(map[string]interface{})["name"]

		//save branch to database
		if err = insertNewBranchDB(code.(string), newID, name.(string), db); err != nil {
			ErrorLogger.Panicf("unable to insert new branch:", err)
		}
		//save branch into linked list
		newBranch := branch{
			Code:              code.(string),
			Name:              name.(string),
			MerchantID:        newID,
			AmountOwed:        0,
			UnclaimedVouchers: nil,
		}
		branchList.addEndNode(newBranch)
		temp = append(temp, newBranch)
	}

	//send post request to web portal
	newJson := JsonReply{
		Ok:  true,
		Msg: "[MS-MERCHANTS]: created merchant data, successful",
		Data: data{
			MerchantID:   newID,
			MerchantName: result["name"].(string),
			Branches:     temp,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newJson)

	////_, err = http.Post("https://localhost:5001/api/v1/voucher", "application/json", bytes.NewReader(jsonValue))
	////if err != nil {
	////	http.Error(w, "unable to send POST request", http.StatusBadGateway)
	////	ErrorLogger.Println("unable to send post request", err)
	////}

}

//func getMerchant(ID string, db *sql.DB) (MerchantData, error) {
//
//	var (
//		merchant_ID string
//		name        string
//		isActive    bool
//	)
//
//	err := db.QueryRow("SELECT * FROM merchants WHERE Merchant_ID = ?", ID).Scan(&merchant_ID, &name, isActive)
//	if err != nil {
//		if err != sql.ErrNoRows {
//			return MerchantData{}, err
//		}
//		return MerchantData{}, errors.New("Merchant not found in database")
//	}
//
//	result := MerchantData{
//		MerchantID:   merchant_ID,
//		MerchantName: name,
//		IsActive:     isActive,
//	}
//	return result, nil
//}

func getAllMerchants(w http.ResponseWriter, r *http.Request) {
	var (
		merchant_ID string
		name        string
		isActive    bool
	)

	var merchants []MerchantData
	db := openDatabase()
	defer db.Close()
	defer fmt.Println("Database closed")

	rows, err := db.Query("SELECT * FROM merchants")
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "there are no merchants in database", http.StatusBadRequest)
			ErrorLogger.Println("there are no merchants in database", err)
			return
		}
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&merchant_ID, &name, &isActive); err != nil {
			http.Error(w, "unable to scan rows", http.StatusBadRequest)
			ErrorLogger.Println("unable to scan rows", err)
			return
		}

		temp := MerchantData{
			MerchantID:   merchant_ID,
			MerchantName: name,
			IsActive:     isActive,
		}

		merchants = append(merchants, temp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Ok   bool
		Msg  string
		Data []MerchantData
	}{true, "[MS-MERCHANTS]: retrieval of list of merchants, successful", merchants})
}

func addBranches(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var result map[string]interface{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "422 - Unable to read request body", http.StatusUnprocessableEntity)
		ErrorLogger.Println("Unable to read request body:", err)
		return
	}

	if err := json.Unmarshal(body, &result); err != nil {
		http.Error(w, "422 - Send in branch details in JSON format", http.StatusUnprocessableEntity)
		ErrorLogger.Println("unable to unmarshal JSON:", err)
		return
	}

	db := openDatabase()
	defer db.Close()
	defer fmt.Println("Database closed")

	//check if merchant exist in database
	exist, err := merchantExistsID(db, params["merchantID"])
	if err != nil {
		http.Error(w, "500 - unable to query database", http.StatusInternalServerError)
		ErrorLogger.Println("unable to query database", err)
		return
	}

	//if does not exist return error
	if !exist {
		http.Error(w, "404 - Merchant ID does not exist in database", http.StatusNotFound)
		ErrorLogger.Println("Merchant ID does not exist in database")
	}
	//if exists check for branch duplicates
	branches := result["branches"]
	for _, b := range branches.([]interface{}) {

		//check for duplicates
		exist, err := branchExists(db, b.(map[string]interface{})["code"].(string))
		if err != nil {
			http.Error(w, "500 - unable to query database:", http.StatusInternalServerError)
			ErrorLogger.Panicln("unable to query database:", err)
		}
		if exist {
			http.Error(w, "409 - Duplicate Branch Code", http.StatusConflict)
			ErrorLogger.Println("409 - Duplicate Branch code")
			return
		}
	}
	//if no duplicates create branches
	var temp []branch
	for _, b := range branches.([]interface{}) {
		code := b.(map[string]interface{})["code"]
		name := b.(map[string]interface{})["name"]

		//save branch to database
		if err = insertNewBranchDB(code.(string), params["merchantID"], name.(string), db); err != nil {
			ErrorLogger.Panicf("unable to insert new branch:", err)
		}
		//save branch into linked list
		newBranch := branch{
			Code:              code.(string),
			Name:              name.(string),
			MerchantID:        params["merchantID"],
			AmountOwed:        0,
			UnclaimedVouchers: nil,
		}
		branchList.addEndNode(newBranch)
		temp = append(temp, newBranch)
	}

	//send post request to web portal
	reply := struct {
		OK   bool
		Msg  string
		Data []branch
	}{true, "[MS-MERCHANTS]: branches added to merchant data, successful", temp}

	//send back results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reply)
}
