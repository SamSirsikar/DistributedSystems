package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCountIslands(t *testing.T) {
	fmt.Println("Lab 1 - Part I \n---- Counting Islands ----")
	assert := assert.New(t)
	rows := [][]int{}

	row1 := []int{1, 1, 0, 0, 0}
	row2 := []int{1, 1, 0, 0, 0}
	row3 := []int{0, 0, 1, 0, 0}
	row4 := []int{0, 0, 0, 1, 1}

	rows = append(rows, row1)
	rows = append(rows, row2)
	rows = append(rows, row3)
	rows = append(rows, row4)

	fmt.Println(rows)
	assert.Equal(3, CountIslands(rows), "Number of islands should be 3")
}