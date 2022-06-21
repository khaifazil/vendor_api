package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
)

func (d *doublyLinkedList) gatherVouchers(w http.ResponseWriter) ([]Voucher, error) {
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
	allVouchers, err := d.gatherVouchers(w)
	if err != nil {
		errorResponse(w, "No vouchers to redeem. Redemption", http.StatusBadRequest)
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
	allVouchers, err := d.gatherVouchers(w)
	if err != nil {
		errorResponse(w, "No vouchers to redeem. Redemption", http.StatusBadRequest)
		ErrorLogger.Println("400 - No vouchers to redeem")
		return
	}

	//process url queries
	values := r.URL.Query()
	pageIndex, err := strconv.Atoi(values["page_index"][0])
	if err != nil {
		errorResponse(w, "Processing url query", http.StatusBadRequest)
		ErrorLogger.Println("unable to process url query")
		return
	}

	recordsPerPage, err := strconv.Atoi(values["records_per_page"][0])
	if err != nil {
		errorResponse(w, "Processing url query", http.StatusBadRequest)
		ErrorLogger.Println("unable to process url query")
		return
	}
	//totalPages := len(allVouchers) / recordsPerPage
	var sliceToSend []Voucher

	if (pageIndex * recordsPerPage) > len(allVouchers) {
		errorResponse(w, "Page_index is out of bounds total pages, get unclaimed vouchers", http.StatusBadRequest)
		ErrorLogger.Println("400 - Page_index is out of bounds total pages")
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

//db := openDatabase()
//defer closeDatabase(db)
//
//for _, branch := range branchSlice {
//	if branch.UnclaimedVouchers == nil {
//		continue
//	}
//	for _, voucher := range branch.UnclaimedVouchers {
//		_, err := db.Exec("UPDATE consumed_vouchers SET Is_Consumed = true WHERE VID = ?", voucher.VID)
//		if err != nil {
//			errorResponse(w, "Error updating database. Fund redemption", 500)
//			ErrorLogger.Println("500 - Error updating database.")
//			return
//		}
//	}
//	//update merchant_branches with correct amounts
//	_, err := db.Exec("UPDATE merchant_branches SET Amount_owed = 0, Amount_claimed = Amount_claimed + ? WHERE Branch_ID = ?", branch.AmountOwed, branch.BranchID)
//	if err != nil {
//		errorResponse(w, "Error updating database. Fund redemption", 500)
//		ErrorLogger.Println("500 - Error updating database.")
//		return
//	}
//}
//
////reset linked list
//err := branchList.deleteAllNodes()
//if err != nil {
//	ErrorLogger.Println(err)
//	return
//}
//
//reply := struct {
//	Ok   bool     `json:"ok"`
//	Msg  string   `json:"msg"`
//	Data []branch `json:"data"`
//}{true, "[MS-MERCHANTS]: Voucher claims, successful", branchSlice}
//w.Header().Set("Content-Type", "application/json")
//w.WriteHeader(http.StatusAccepted)
//json.NewEncoder(w).Encode(reply)
