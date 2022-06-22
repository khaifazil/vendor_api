package main

import "fmt"

func init() {

	//fmt.Printf("%+v", branchList)
}

func initFromDatabase() {
	db := openDatabase()
	defer closeDatabase(db)

	var (
		branchID   string
		name       string
		branchCode string
		merchantID string
		amountOwed int
	)

	var (
		VID         string
		customerID  string
		amount      int
		isConsumed  bool
		isClaimed   bool
		isValidated bool
	)
	//get merchant branches with outstanding amount owed
	rows, err := db.Query("SELECT Branch_ID, Name, Branch_Code, MerchantID, Amount_owed FROM merchant_branches WHERE Amount_owed > 0 ORDER BY Branch_ID")
	if err != nil {
		ErrorLogger.Fatal("unable to initialize data from database:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var vouchers []Voucher
		if err := rows.Scan(&branchID, &name, &branchCode, &merchantID, &amountOwed); err != nil {
			ErrorLogger.Fatal("rows.Scan:", err)
		}

		vouchersDB, err := db.Query("SELECT * FROM consumed_vouchers WHERE Is_Claimed = false && Branch_ID = ?", branchID)
		if err != nil {
			ErrorLogger.Fatal("unable to initialize data from database:", err)
		}
		defer vouchersDB.Close()

		for vouchersDB.Next() {
			if err := vouchersDB.Scan(&VID, &branchID, &customerID, &amount, &isConsumed, &isClaimed, &isValidated); err != nil {
				ErrorLogger.Fatal("unable to initialize data from database:", err)
			}
			voucher := Voucher{
				VID:         VID,
				CustomerID:  customerID,
				Amount:      amount,
				BranchID:    branchID,
				IsValidated: isValidated,
				IsConsumed:  isConsumed,
				IsClaimed:   isClaimed,
			}
			vouchers = append(vouchers, voucher)
		}

		branch := branch{
			BranchID:          branchID,
			Name:              name,
			BranchCode:        branchCode,
			MerchantID:        merchantID,
			AmountOwed:        amountOwed,
			UnclaimedVouchers: vouchers,
		}

		branchList.addEndNode(branch)
		fmt.Println("Local cache initialized")
	}

}
