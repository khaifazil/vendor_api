package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func (d *doublyLinkedList) gatherVouchers() ([]Voucher, error) {
	var allVouchers []Voucher
	c := make(chan Voucher)

	if d.Size == 0 {
		return allVouchers, errors.New("no vouchers to redeem")
	} else if d.Size == 1 {
		for _, v := range d.Head.Data.UnclaimedVouchers {
			allVouchers = append(allVouchers, v)
		}
	} else {
		wg.Add(1)
		go func() {
			currentNode := d.Head
			for i := 1; i <= d.Size/2; i++ {
				for _, v := range currentNode.Data.UnclaimedVouchers {
					c <- v
				}
				currentNode = currentNode.Next
			}
			wg.Done()
		}()
		wg.Add(1)
		go func() {
			currentNode := d.Tail
			for i := 1; i <= d.Size-(d.Size/2); i++ {
				for _, v := range currentNode.Data.UnclaimedVouchers {
					c <- v
				}
				currentNode = currentNode.Prev
			}
			wg.Done()
		}()

		go func() {
			wg.Wait()
			close(c)
		}()
		for elem := range c {
			allVouchers = append(allVouchers, elem)
		}
	}

	sort.SliceStable(allVouchers, func(i, j int) bool {
		return allVouchers[i].VID < allVouchers[j].VID
	})
	return allVouchers, nil
}

func (d *doublyLinkedList) totalUnclaimedVoucher(w http.ResponseWriter, r *http.Request) {

	//check validation of apikey in header
	if !validateAPIKey(r) {
		errorResponse(w, "API key is unauthorized", http.StatusUnauthorized)
		ErrorLogger.Println("API key is unauthorized")
		return
	}

	allVouchers, err := d.gatherVouchers()
	if err != nil {
		errorResponse(w, "No vouchers to redeem", http.StatusBadRequest)
		ErrorLogger.Println("400 - No vouchers to redeem")
		return
	}
	type data struct {
		TotalVouchers int `json:"total_vouchers"`
	}
	reply := struct {
		Ok   bool   `json:"ok"`
		Msg  string `json:"msg"`
		Data data   `json:"data"`
	}{true, "[MS-MERCHANTS]: Voucher claims, successful", data{len(allVouchers)}}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(reply)

	//for i, v := range allVouchers {
	//	fmt.Printf("%v - %+v\n", i, v)
	//}
}

//handler to send over unclaimed vouchers for processing
func (d *doublyLinkedList) sendVouchers(w http.ResponseWriter, r *http.Request) {

	//check validation of apikey in header
	if !validateAPIKey(r) {
		errorResponse(w, "API key is unauthorized", http.StatusUnauthorized)
		ErrorLogger.Println("API key is unauthorized")
		return
	}

	allVouchers, err := d.gatherVouchers()
	if err != nil {
		errorResponse(w, "No vouchers to redeem", http.StatusBadRequest)
		ErrorLogger.Println("400 - No vouchers to redeem")
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

	var sliceToSend []Voucher

	if (pageIndex * recordsPerPage) > len(allVouchers) {
		errorResponse(w, "Page_index is out of bounds of total pages", http.StatusBadRequest)
		ErrorLogger.Println("400 - Page_index is out of bounds of total pages")
		return
	} else if (pageIndex*recordsPerPage)+recordsPerPage > len(allVouchers) {
		sliceToSend = allVouchers[(pageIndex * recordsPerPage):]
	} else {
		sliceToSend = allVouchers[(pageIndex * recordsPerPage) : (pageIndex*recordsPerPage)+recordsPerPage]
	}

	reply := struct {
		Ok   bool      `json:"ok"`
		Msg  string    `json:"msg"`
		Data []Voucher `json:"data"`
	}{true, "[MS-MERCHANTS]: Get unclaimed vouchers, successful", sliceToSend}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(reply)

	//fmt.Printf("%+v", sliceToSend)
}

func claimVoucher(w http.ResponseWriter, r *http.Request) {

	//check validation of apikey in header
	if !validateAPIKey(r) {
		errorResponse(w, "API key is unauthorized", http.StatusUnauthorized)
		ErrorLogger.Println("API key is unauthorized")
		return
	}

	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResponse(w, "unable to read request body", http.StatusUnprocessableEntity)
		ErrorLogger.Println("unable to read request body:", err)
		return
	}

	var results []*Voucher
	//decodes request body into results
	json.NewDecoder(strings.NewReader(string(resp))).Decode(&results)

	db := openDatabase()
	defer closeDatabase(db)

	for _, voucher := range results {

		_, err := db.Exec("UPDATE consumed_vouchers SET Is_Claimed = true WHERE VID = ?", voucher.VID)
		if err != nil {
			errorResponse(w, "Error updating database", 500)
			ErrorLogger.Println("500 - Error updating database.")
			return
		}
		//update merchant_branches with correct amounts
		_, err = db.Exec("UPDATE merchant_branches SET Amount_owed = Amount_owed - ?, Amount_claimed = Amount_claimed + ? WHERE Branch_ID = ?", voucher.Amount, voucher.Amount, voucher.BranchID)
		if err != nil {
			errorResponse(w, "Error updating database", 500)
			ErrorLogger.Println("500 - Error updating database.")
			return
		}
		voucher.IsClaimed = true
		//fmt.Println(voucher)
	}

	reply := struct {
		Ok   bool       `json:"ok"`
		Msg  string     `json:"msg"`
		Data []*Voucher `json:"data"`
	}{true, "[MS-MERCHANTS]: Voucher claims, successful", results}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(reply)
}

func reloadLocalCache(w http.ResponseWriter, r *http.Request) {

	//check validation of apikey in header
	if !validateAPIKey(r) {
		errorResponse(w, "API key is unauthorized", http.StatusUnauthorized)
		ErrorLogger.Println("API key is unauthorized")
		return
	}

	//reset linked list
	err := branchList.deleteAllNodes()
	if err != nil {
		errorResponse(w, "Error deleting local cache", 500)
		ErrorLogger.Println(err)
		return
	}

	initFromDatabase()
	reply := struct {
		Ok   bool     `json:"ok"`
		Msg  string   `json:"msg"`
		Data struct{} `json:"data"`
	}{true, "[MS-MERCHANTS]: Reload local cache, successful", struct{}{}}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(reply)
}
