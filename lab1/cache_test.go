package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCache(t *testing.T) {
	assert := assert.New(t)
	fmt.Println("Lab 1 - Part II \n---- LRU Cache ----")
	// Populates entries into the Cache
	Set(1, 10)
	Set(2, 20)
	Set(3, 30)

	assert.Equal(10, Get(1), "Get(1) should return 10")
	assert.Equal(20, Get(2), "Get(2) should return 20")
	assert.Equal(30, Get(3), "Get(3) should return 30")

	Set(4, 40)
	assert.Equal(40, Get(4), "Get(4) should return 40")
	// Checks Cache Invalidation
	assert.Equal(-1, Get(1), "Get(1) should return -1 as it was the least recently used entry")
}