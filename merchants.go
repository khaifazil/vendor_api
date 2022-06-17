package main

import "errors"

type MerchantData struct {
	MerchantID        string
	MerchantName      string
	AmountOwed        int
	UnclaimedVouchers []Voucher
}

var merchantsList doublyLinkedList

func init() {

	merchantsList.addEndNode(struct {
		MerchantID        string
		MerchantName      string
		AmountOwed        int
		UnclaimedVouchers []Voucher
	}{MerchantID: "1234586", MerchantName: "NTUC", AmountOwed: 0, UnclaimedVouchers: nil})
}

func (d *doublyLinkedList) searchListForMerchant(merchantID string) (*MerchantData, error) {
	currentNode := d.Head
	for currentNode != nil {

		if currentNode.Data.MerchantID == merchantID {
			return &currentNode.Data, nil
		}
		currentNode = currentNode.Next
	}
	return &MerchantData{}, errors.New("merchant not found in list")
}
