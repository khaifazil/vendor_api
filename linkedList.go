package main

import (
	"errors"
	"fmt"
)

type ListNode struct {
	Data branch
	Next *ListNode
	Prev *ListNode
}

type doublyLinkedList struct {
	Head *ListNode
	Tail *ListNode
	Size int
}

func initDoublyLinkedList() *doublyLinkedList {
	return &doublyLinkedList{}
}

func (d *doublyLinkedList) addFrontNode(b branch) {
	newNode := &ListNode{
		Data: b,
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
	d.Size++
}

func (d *doublyLinkedList) addEndNode(b branch) {
	newNode := &ListNode{
		Data: b,
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
	d.Size++
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
	return d.Size
}

func (d *doublyLinkedList) removeNode(node *ListNode) error {

	if d.Size == 0 { //if list is empty
		return errors.New("linkedList is empty")
	} else if d.Size == 1 { //if list has only 1 node
		d.Head = nil
		d.Tail = nil
		d.Size--
	} else if d.Head == node { //if node is head of list
		d.Head = node.Next
		node.Next = nil
		d.Head.Prev = nil
		d.Size--
	} else if d.Tail == node { // if node is tail of list
		d.Tail = node.Prev
		node.Prev = nil
		d.Tail.Next = nil
		d.Size--
	} else { // if node is somewhere in the middle
		node.Next.Prev = node.Prev
		node.Prev.Next = node.Next
		node.Next = nil
		node.Prev = nil
		d.Size--
	}

	return nil
}

func (d *doublyLinkedList) searchListForNode(branchCode string) (*ListNode, bool) {
	currentNode := d.Head
	for currentNode != nil {

		if currentNode.Data.Code == branchCode {
			return currentNode, true
		}
		currentNode = currentNode.Next
	}
	return nil, false
}
