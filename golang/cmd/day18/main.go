package main

import (
	"bufio"
	_ "embed"
	"github.com/samber/lo"
	"slices"
	"strconv"
	"strings"
)

//go:embed input
var input string

type coordinate struct {
	row int
	col int
}

type memory struct {
	memory [][]int
}

func newMemory(width int, height int, corruptedBytes []coordinate, numBytes int) memory {
	memoryRows := make([][]int, height)
	for i := range memoryRows {
		memoryRows[i] = make([]int, width)
	}
	for i := range numBytes {
		memoryRows[corruptedBytes[i].row][corruptedBytes[i].col] = -1
	}
	return memory{
		memory: memoryRows,
	}
}

func (m *memory) inBounds(c coordinate) bool {
	if c.row < 0 || c.col < 0 {
		return false
	}
	if c.row >= len(m.memory) || c.col >= len(m.memory[0]) {
		return false
	}
	return true
}

func (m *memory) canMove(c coordinate) bool {
	if !m.inBounds(c) {
		return false
	}
	return !m.isCorrupted(c)
}

func (m *memory) nextMoves(c coordinate) []coordinate {
	north := coordinate{row: c.row - 1, col: c.col}
	east := coordinate{row: c.row, col: c.col + 1}
	south := coordinate{row: c.row + 1, col: c.col}
	west := coordinate{row: c.row, col: c.col - 1}

	results := make([]coordinate, 0)

	if m.canMove(north) {
		results = append(results, north)
	}
	if m.canMove(east) {
		results = append(results, east)
	}
	if m.canMove(south) {
		results = append(results, south)
	}
	if m.canMove(west) {
		results = append(results, west)
	}
	return results
}

func (m *memory) isCorrupted(c coordinate) bool {
	return m.memory[c.row][c.col] < 0
}

type moveChain struct {
	coordinates []coordinate
}

func (mc *moveChain) cost() int {
	// Offset by -1 because the start coordinate doesn't cost a move.
	return len(mc.coordinates) - 1

}

func (mc *moveChain) addMove(c coordinate) moveChain {
	return moveChain{
		coordinates: append(slices.Clone(mc.coordinates), c),
	}
}

func (m *memory) findPath() moveChain {
	coordToCost := make(map[coordinate]int)
	coordToCost[coordinate{row: 0, col: 0}] = 0

	goal := coordinate{
		row: len(m.memory) - 1,
		col: len(m.memory[0]) - 1,
	}

	start := coordinate{0, 0}
	initialMoveChain := moveChain{coordinates: []coordinate{start}}
	frontier := []moveChain{initialMoveChain}
	goalPaths := make([]moveChain, 0)
	for len(frontier) > 0 {
		newFrontier := make([]moveChain, 0)

		for i := range frontier {
			last, found := lo.Last(frontier[i].coordinates)
			if !found {
				panic("Should always have a last")
			}

			for _, nextMove := range m.nextMoves(last) {
				nextChain := frontier[i].addMove(nextMove)
				mappedCost, found := coordToCost[nextMove]
				if found && mappedCost <= nextChain.cost() {
					continue
				}

				coordToCost[nextMove] = nextChain.cost()

				if nextMove == goal {
					goalPaths = append(goalPaths, nextChain)
					continue
				}

				newFrontier = append(newFrontier, nextChain)
			}
		}

		frontier = newFrontier
	}

	slices.SortFunc(goalPaths, func(a moveChain, b moveChain) int {
		return a.cost() - b.cost()
	})
	return goalPaths[0]
}

func (m *memory) print() {
	for i := range m.memory {
		for j := range m.memory[i] {
			if m.isCorrupted(coordinate{row: i, col: j}) {
				print("#")
			} else {
				print(".")
			}
		}
		println()
	}

}

func handleLine(line string) coordinate {
	tokens := strings.Split(line, ",")
	x, err := strconv.Atoi(tokens[0])
	if err != nil {
		panic(err)
	}
	y, err := strconv.Atoi(tokens[1])
	if err != nil {
		panic(err)
	}
	return coordinate{
		row: y,
		col: x,
	}
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	badBytes := make([]coordinate, 0)
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		badBytes = append(badBytes, handleLine(scanner.Text()))
	}

	dimension := 71
	mem := newMemory(dimension, dimension, badBytes, 1024)
	mem.print()
	path := mem.findPath()
	println(path.cost())
}
