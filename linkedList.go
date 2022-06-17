package main

import (
	"errors"
	"fmt"
)

type ListNode struct {
	Data MerchantData
	Next *ListNode
	Prev *ListNode
}

type doublyLinkedList struct {
	Head *ListNode
	Tail *ListNode
	size int
}

func initDoublyLinkedList() *doublyLinkedList {
	return &doublyLinkedList{}
}

func (d *doublyLinkedList) addFrontNode(merchant MerchantData) {
	newNode := &ListNode{
		Data: merchant,
	}

	if d.Head == nil { //if list is empty
		d.Head = newNode
		d.Tail = newNode
	} else { // if list is not empty
		newNode.Next = d.Head
		d.Head.Prev = newNode
		d.Head = newNode
		newNode.Prev = nil
	}
	d.size++
}

func (d *doublyLinkedList) addEndNode(merchant MerchantData) {
	newNode := &ListNode{
		Data: merchant,
	}

	if d.Head == nil {
		d.Head = newNode
		d.Tail = newNode
	} else {
		d.Tail.Next = newNode
		newNode.Prev = d.Tail
		d.Tail = newNode
		newNode.Next = nil
	}
	d.size++
}

func (d *doublyLinkedList) traverseForward() error {
	if d.Head == nil {
		return errors.New("traversal error: List is empty")
	}

	currentNode := d.Head
	for currentNode != nil {
		fmt.Println(currentNode) //TODO: change to send query or marshal to json
		currentNode = currentNode.Next
	}
	return nil
}

func (d *doublyLinkedList) traverseBackwards() error {
	if d.Head == nil {
		return errors.New("traversal error: List is empty")
	}

	currentNode := d.Tail
	for currentNode != nil {
		fmt.Println(currentNode) //TODO: change to send query or marshal to json
		currentNode = currentNode.Prev
	}
	return nil
}

func (d *doublyLinkedList) length() int {
	return d.size
}
