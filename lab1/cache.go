package main

import (
	"container/list"
	"fmt"
	"strconv"
	"strings"
)

// DO NOT CHANGE THIS CACHE SIZE VALUE
const CACHE_SIZE int = 3

type LRUCache struct {
	maxCapacity     int
	currentCapacity int
	hash            map[int]*list.Element
	queue           *list.List
}

var c = New(CACHE_SIZE)

func New(capacity int) *LRUCache {
	// initialize the cache and return
	return &LRUCache{
		maxCapacity:     capacity,
		currentCapacity: capacity,
		hash:            make(map[int]*list.Element),
		queue:           list.New(),
	}
}

func Set(key int, value int) {
	// TODO: add your code here!
	if c.currentCapacity == 0 {
		// get the last element
		lastElement := c.queue.Back()

		// delete the last element from the list
		elementValue := c.queue.Remove(lastElement)

		// fetch the actual value from the element value
		s := strings.Split(elementValue.(string), ":")
		tempkey := s[0]

		// delete the key from hashmap
		var intTempKey int
		fmt.Sscan(tempkey, &intTempKey)
		delete(c.hash, intTempKey)

		c.currentCapacity++
	}

	newElement := c.queue.PushFront(strconv.Itoa(key) + ":" + strconv.Itoa(value))
	c.hash[key] = newElement
	c.currentCapacity--
}

func Get(key int) int {
	if valPointer, ok := c.hash[key]; ok {
		valElement := *(*list.Element)(valPointer)
		c.queue.MoveToFront(&valElement)
		c.hash[key] = &valElement
		firstElement := c.queue.Front()
		temp := firstElement.Value
		s := strings.Split(temp.(string), ":")
		tempValue := s[1]
		var value int
		fmt.Sscan(tempValue, &value)
		return value
	}
	return -1
}
