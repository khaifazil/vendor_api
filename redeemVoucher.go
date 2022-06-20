package main

import (
	"encoding/json"
	"net/http"
)

func (d *doublyLinkedList) claimVoucher(w http.ResponseWriter, r *http.Request) {
	var branchSlice []branch
	c := make(chan branch)

	if d.Size == 0 {
		errorResponse(w, "No vouchers to redeem. Redemption", http.StatusBadRequest)
		ErrorLogger.Println("400 - No vouchers to redeem")
		return
	} else if d.Size == 1 {
		branchSlice = append(branchSlice, d.Head.Data)
	} else {
		wg.Add(1)
		go func() {
			currentNode := d.Head
			for i := 1; i <= d.Size/2; i++ {
				c <- currentNode.Data
				currentNode = currentNode.Next
			}
			wg.Done()
		}()
		wg.Add(1)
		go func() {
			currentNode := d.Tail
			for i := 1; i <= d.Size-(d.Size/2); i++ {
				c <- currentNode.Data
				currentNode = currentNode.Prev
			}
			wg.Done()
		}()

		go func() {
			wg.Wait()
			close(c)
		}()
		for elem := range c {
			branchSlice = append(branchSlice, elem)
		}
	}

	db := openDatabase()
	defer closeDatabase(db)

	for _, branch := range branchSlice {
		if branch.UnclaimedVouchers == nil {
			continue
		}
		for _, voucher := range branch.UnclaimedVouchers {
			_, err := db.Exec("UPDATE consumed_vouchers SET Is_Consumed = true WHERE VID = ?", voucher.VID)
			if err != nil {
				errorResponse(w, "Error updating database. Fund redemption", 500)
				ErrorLogger.Println("500 - Error updating database.")
				return
			}
		}
		//update merchant_branches with correct amounts
		_, err := db.Exec("UPDATE merchant_branches SET Amount_owed = 0, Amount_claimed = Amount_claimed + ? WHERE Branch_ID = ?", branch.AmountOwed, branch.BranchID)
		if err != nil {
			errorResponse(w, "Error updating database. Fund redemption", 500)
			ErrorLogger.Println("500 - Error updating database.")
			return
		}
	}

	//reset linked list
	err := branchList.deleteAllNodes()
	if err != nil {
		ErrorLogger.Println(err)
		return
	}

	reply := struct {
		Ok   bool     `json:"ok"`
		Msg  string   `json:"msg"`
		Data []branch `json:"data"`
	}{true, "[MS-MERCHANTS]: Voucher claims, successful", branchSlice}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(reply)
}
