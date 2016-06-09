package main

import (
	"strconv"
)

func CountIslands(grid [][]int) int {

	//visited map
	visited := make(map[string]bool)

	//count of elements in region
	count := 0

	//region count
	regionCount := 0

	var rows int = len(grid)
	var cols int = len(grid[0])

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			numberRegions(&grid, rows, cols, i, j, visited, &count)
			//count number of regions
			if count > 0 {
				regionCount++
			}
			count = 0

		}
	}


	return regionCount
}

func numberRegions(matrix *[][]int, m int, n int, x int, y int, visited map[string]bool, count *int) {

	if x < 0 || x >= m {
		return
	}

	if y < 0 || y >= n {
		return
	}

	var is_visited bool = visited[strconv.Itoa(x)+strconv.Itoa(y)] || false

	if is_visited == true {
		return
	}

	if (*matrix)[x][y] == 0 {
		return
	}

	*count = *count + 1

	visited[strconv.Itoa(x)+strconv.Itoa(y)] = true

	//left
	numberRegions(matrix, m, n, x-1, y, visited, count)
	//top
	numberRegions(matrix, m, n, x, y-1, visited, count)
	//right
	numberRegions(matrix, m, n, x+1, y, visited, count)
	//bottom
	numberRegions(matrix, m, n, x, y+1, visited, count)

}
