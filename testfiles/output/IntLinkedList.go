package main

import (
	"fmt"
)

type Node struct {
	data int
	next *Node
}

func NewNode(data int) *Node {
	return &Node{
		data: data,
		next: nil,
	}
}

type IntLinkedList struct {
	size int
	head *Node
	tail *Node
}

func NewIntLinkedList() *IntLinkedList {
	return &IntLinkedList{
		head: nil,
		tail: nil,
		size: 0,
	}
}

func (l IntLinkedList) Size() int {
	return l.size
}

func (l *IntLinkedList) Add(element int) {
	if l.head == nil {
		l.size++
		l.head = NewNode(element)
		l.tail = l.head
	} else {
		l.size++
		l.tail.next = NewNode(element)
		l.tail = l.tail.next
	}
}

func (l IntLinkedList) Get(index int) int {
	currentNode := l.head

	for i := 1; i <= index; i++ {
		currentNode = currentNode.next
	}
	return currentNode.data
}

func (l *IntLinkedList) Remove(index int) {
	currentNode := l.head
	previousNode := l.head

	if index == 0 {
		l.head = l.head.next
	}

	for i := 1; i <= index; i++ {
		previousNode = currentNode
		currentNode = previousNode.next
	}

	if currentNode.next == nil {
		l.tail = previousNode
	} else {
		previousNode.next = currentNode.next
	}

	l.size--
}

func (l IntLinkedList) String() string {
	result := "{"
	if l.Size() > 0 {
		result += string(l.Get(0))
	}
	for i := 1; i < l.size; i++ {
		result += ", " + string(l.Get(i))
	}
	result += "}"
	return result
}

func AssertArraysEqual(list *IntLinkedList, answer []int) {
	if list.Size() != len(answer) {
		panic("Expected list of length " + string(len(answer)) + " but got " + string(list.Size()))
	}
	for i := 0; i < len(answer); i++ {
		if list.Get(i) != answer[i] {
			panic("Expected " + string(answer[i]) + " but got " + string(list.Get(i)) + " at index " + string(i))
		}
	}
}

func Test1() {
	list := NewIntLinkedList()
	answer := []int{}
	AssertArraysEqual(list, answer)
}

func Test2() {
	list := NewIntLinkedList()
	for i := 0; i < 3; i++ {
		list.Add(i * i)
	}
	answer := []int{0, 1, 4}
	AssertArraysEqual(list, answer)
}

func Test3() {
	list := NewIntLinkedList()
	for i := 0; i < 5; i++ {
		list.Add(i * i)
	}
	list.Remove(1)
	list.Remove(2)
	answer := []int{0, 4, 16}
	AssertArraysEqual(list, answer)
}

func Test4() {
	list := NewIntLinkedList()
	for i := 0; i < 5; i++ {
		list.Add(i * i)
	}

	list.Remove(1)
	list.Remove(2)

	list.Remove(0)
	list.Remove(0)
	list.Remove(0)
	answer1 := []int{}
	AssertArraysEqual(list, answer1)

	for i := 0; i < 5; i++ {
		list.Add(i * i)
	}
	list.Remove(4)
	list.Add(-1)
	answer2 := []int{0, 1, 4, 9, -1}
	AssertArraysEqual(list, answer2)
}

func Test5() {
	list := NewIntLinkedList();
	for i := 0; i < 12; i++ {
		list.Add(i * i)
	}
	answer := []int{0, 1, 4, 9, 16, 25, 36, 49, 64, 81, 100, 121}
	AssertArraysEqual(list, answer)
}

func main() {
	Test1()
	Test2()
	Test3()
	Test4()
	Test5()

	fmt.Println("pass")
}
