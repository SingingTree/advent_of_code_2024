package main

import (
	"bufio"
	_ "embed"
	"github.com/samber/lo"
	"strings"
)

//go:embed input
var input string

func handleLine(line string) []string {
	return lo.ChunkString(line, 1)
}

func checkHorizontal(
	grid [][]string,
	row int,
	col int,
	word []string,
) bool {
	for i := range word {
		checkCol := col + i
		if checkCol >= len(grid[row]) {
			return false
		}
		if grid[row][checkCol] != word[i] {
			return false
		}
	}
	return true
}

func checkVertical(
	grid [][]string,
	row int,
	col int,
	word []string,
) bool {

	for i := range word {
		checkRow := row + i
		if checkRow >= len(grid) {
			return false
		}
		if grid[checkRow][col] != word[i] {
			return false
		}
	}
	return true
}

func checkDiagonalLeft(
	grid [][]string,
	row int,
	col int,
	word []string,
) bool {
	for i := range word {
		checkRow := row + i
		if checkRow >= len(grid) {
			return false
		}
		checkCol := col - i
		if checkCol < 0 {
			return false
		}
		if grid[checkRow][checkCol] != word[i] {
			return false
		}
	}
	return true
}

func checkDiagonalRight(
	grid [][]string,
	row int,
	col int,
	word []string,
) bool {
	for i := range word {
		checkRow := row + i
		if checkRow >= len(grid) {
			return false
		}
		checkCol := col + i
		if checkCol >= len(grid[checkRow]) {
			return false
		}
		if grid[checkRow][checkCol] != word[i] {
			return false
		}
	}
	return true
}

func countXmas(grid [][]string) int {
	count := 0
	for row := 0; row < len(grid); row++ {
		for col := 0; col < len(grid[row]); col++ {
			if checkHorizontal(grid, row, col, []string{"X", "M", "A", "S"}) {
				count += 1
			}
			if checkVertical(grid, row, col, []string{"X", "M", "A", "S"}) {
				count += 1
			}
			if checkDiagonalLeft(grid, row, col, []string{"X", "M", "A", "S"}) {
				count += 1
			}
			if checkDiagonalRight(grid, row, col, []string{"X", "M", "A", "S"}) {
				count += 1
			}

			if checkHorizontal(grid, row, col, []string{"S", "A", "M", "X"}) {
				count += 1
			}
			if checkVertical(grid, row, col, []string{"S", "A", "M", "X"}) {
				count += 1
			}
			if checkDiagonalLeft(grid, row, col, []string{"S", "A", "M", "X"}) {
				count += 1
			}
			if checkDiagonalRight(grid, row, col, []string{"S", "A", "M", "X"}) {
				count += 1
			}
		}
	}

	return count
}

func countMas(grid [][]string) int {
	count := 0
	for row := 0; row < len(grid)-2; row++ {
		for col := 0; col < len(grid[row])-2; col++ {
			hasLeft := false
			hasRight := false
			if checkDiagonalRight(grid, row, col, []string{"M", "A", "S"}) {
				hasLeft = true
			}
			if checkDiagonalRight(grid, row, col, []string{"S", "A", "M"}) {
				hasLeft = true
			}

			if checkDiagonalLeft(grid, row, col+2, []string{"M", "A", "S"}) {
				hasRight = true
			}
			if checkDiagonalLeft(grid, row, col+2, []string{"S", "A", "M"}) {
				hasRight = true
			}

			if hasLeft && hasRight {
				count += 1
			}
		}
	}

	return count
}

func main() {
	grid := make([][]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		if scanner.Text() == "" {
			// Skip blank lines.
			continue
		}
		gridLine := handleLine(scanner.Text())
		grid = append(grid, gridLine)
	}

	// Part 1
	count := countXmas(grid)
	println(count)

	count = countMas(grid)
	println(count)
}
