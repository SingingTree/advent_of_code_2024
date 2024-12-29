package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"github.com/samber/lo"
	"gonum.org/v1/gonum/stat/combin"
	"slices"
	"strconv"
	"strings"
)

//go:embed input
var input string

func handleLine(line string) []string {
	return lo.ChunkString(line, 1)
}

type move int

const (
	up move = iota
	down
	left
	right
	press
)

func (m *move) toString() string {
	switch *m {
	case up:
		return "^"
	case down:
		return "v"
	case left:
		return "<"
	case right:
		return ">"
	case press:
		return "A"
	default:
		panic("invalid move")
	}

}

type coordinate struct {
	row int
	col int
}

func (c *coordinate) applyMove(m move) *coordinate {
	switch m {
	case up:
		return &coordinate{c.row - 1, c.col}
	case down:
		return &coordinate{c.row + 1, c.col}
	case left:
		return &coordinate{c.row, c.col - 1}
	case right:
		return &coordinate{c.row, c.col + 1}
	default:
		panic("invalid move")

	}
}

// Returns a list of list of moves. Each inner list is a set of viable moves to
// reach the desired location.
func (c *coordinate) movesToOther(other coordinate, unsafeCoordinate coordinate) [][]move {
	rowDiff := c.row - other.row
	goingDown := rowDiff < 0
	if goingDown {
		rowDiff = -rowDiff
	}
	rowMoves := make([]move, rowDiff)
	for i := range rowMoves {
		if goingDown {
			rowMoves[i] = down
		} else {
			rowMoves[i] = up
		}
	}

	colDiff := c.col - other.col
	goingRight := colDiff < 0
	if goingRight {
		colDiff = -colDiff
	}
	colMoves := make([]move, colDiff)
	for i := range colMoves {
		if goingRight {
			colMoves[i] = right
		} else {
			colMoves[i] = left
		}
	}

	unsortedMoves := append(slices.Clone(colMoves), rowMoves...)
	moveCombinations := combin.Permutations(len(unsortedMoves), len(unsortedMoves))
	moves := make([][]move, len(moveCombinations))
	for _, moveCombination := range moveCombinations {
		currentMove := make([]move, len(moveCombination))
		for i, idx := range moveCombination {
			currentMove[i] = unsortedMoves[idx]
		}
		moves = append(moves, currentMove)
	}

	touchesUnsafe := func(moves []move) bool {
		currentCoord := c
		for _, m := range moves {
			currentCoord = currentCoord.applyMove(m)
			if currentCoord.row == unsafeCoordinate.row && currentCoord.col == unsafeCoordinate.col {
				return true
			}
		}
		return false
	}

	moveChainsEq := func(lhs []move, rhs []move) bool {
		if len(lhs) != len(rhs) {
			return false
		}
		for i := range lhs {
			if lhs[i] != rhs[i] {
				return false
			}
		}
		return true
	}

	// Filter out any moves that go via the unsafe coordinate.
	filteredMoves := lo.Filter(moves, func(moves []move, _ int) bool {
		if moves == nil {
			return false
		}
		if touchesUnsafe(moves) {
			return false
		}
		return true
	})
	uniqMoves := make([][]move, 0)
	for _, lhs := range filteredMoves {
		dupe := false
		for _, rhs := range uniqMoves {
			if moveChainsEq(lhs, rhs) {
				dupe = true
				break
			}
		}
		if !dupe {
			uniqMoves = append(uniqMoves, lhs)
		}
	}

	return uniqMoves
}

type fromToStringPair struct {
	from string
	to   string
}

func getNumericKeypadMoveMap() map[fromToStringPair][][]move {
	m := make(map[fromToStringPair][][]move)
	keys := [][]string{
		{"7", "8", "9"},
		{"4", "5", "6"},
		{"1", "2", "3"},
		{"?", "0", "A"},
	}

	unsafeCoordinate := coordinate{
		row: 3,
		col: 0,
	}

	for row, line := range keys {
		for col, key := range line {
			for row2, line2 := range keys {
				for col2, key2 := range line2 {
					if key == "?" || key2 == "?" {
						// Don't figure out mappings to unsafes.
						continue
					}
					keyCoordinate := coordinate{row: row, col: col}
					key2Coordinate := coordinate{row: row2, col: col2}
					fromTo := fromToStringPair{key, key2}
					moveCandidates := keyCoordinate.movesToOther(key2Coordinate, unsafeCoordinate)
					m[fromTo] = moveCandidates
				}
			}
		}
	}

	return m
}

func getArrowKeypadMoveMap() map[fromToStringPair][][]move {
	m := make(map[fromToStringPair][][]move)
	keys := [][]string{
		{"?", "^", "A"},
		{"<", "v", ">"},
	}

	unsafeCoordinate := coordinate{
		row: 0,
		col: 0,
	}

	for row, line := range keys {
		for col, key := range line {
			for row2, line2 := range keys {
				for col2, key2 := range line2 {
					if key == "?" || key2 == "?" {
						// Don't figure out mappings to unsafes.
						continue
					}
					keyCoordinate := coordinate{row: row, col: col}
					key2Coordinate := coordinate{row: row2, col: col2}
					fromTo := fromToStringPair{key, key2}
					moveCandidates := keyCoordinate.movesToOther(key2Coordinate, unsafeCoordinate)
					m[fromTo] = moveCandidates
				}
			}
		}
	}

	return m
}

type moveFinder struct {
	currentNumericKey      string
	numArrowKeypads        int
	currentArrowKeypadKeys []string
	numericKeypadMoves     map[fromToStringPair][][]move
	arrowKeypadMoves       map[fromToStringPair][][]move
	// optimalMoves contains the cost for a given string of moves at a given
	// level. Level 0 is the arrow keypad closest to the numeric keypad, while
	// level n == numArrowKeypads - 1 is the arrow keypad furthest from the
	// numeric keypad.
	optimalMoves []map[string]int
}

func newMoveFinder(numArrows int) moveFinder {
	currentArrowKeypadKeys := make([]string, numArrows)
	for i := range numArrows {
		currentArrowKeypadKeys[i] = "A"
	}
	optimalMoves := make([]map[string]int, numArrows)
	for i := range optimalMoves {
		optimalMoves[i] = make(map[string]int)
	}

	mf := moveFinder{
		currentNumericKey:      "A",
		numArrowKeypads:        numArrows,
		currentArrowKeypadKeys: currentArrowKeypadKeys,
		numericKeypadMoves:     getNumericKeypadMoveMap(),
		arrowKeypadMoves:       getArrowKeypadMoveMap(),
		optimalMoves:           optimalMoves,
	}

	return mf
}

func (mf *moveFinder) clone() moveFinder {
	return moveFinder{
		currentNumericKey:      mf.currentNumericKey,
		currentArrowKeypadKeys: slices.Clone(mf.currentArrowKeypadKeys),
		numericKeypadMoves:     mf.numericKeypadMoves,
		arrowKeypadMoves:       mf.arrowKeypadMoves,
	}
}

func (mf *moveFinder) pressNumeric(numericKey string) [][]move {
	fromTo := fromToStringPair{
		from: mf.currentNumericKey,
		to:   numericKey,
	}
	numericMoveCandidates, ok := mf.numericKeypadMoves[fromTo]
	if !ok {
		panic("Unrecognized keys")
	}
	mf.currentNumericKey = numericKey

	return slices.Clone(numericMoveCandidates)
}

func (mf *moveFinder) pressNumericMultipleTimes(numericKeys []string) [][]move {
	if mf.currentNumericKey != "A" {
		panic("Should always start on A")
	}
	initialMoves := mf.pressNumeric(numericKeys[0])
	moves := make([][]move, len(initialMoves))
	for i := range initialMoves {
		moves[i] = slices.Clone(initialMoves[i])
		moves[i] = append(moves[i], press)
	}
	for _, m := range numericKeys[1:] {
		newMoves := make([][]move, 0)
		moreMoves := mf.pressNumeric(m)
		for _, moreMoveChain := range moreMoves {
			for _, existingMoveChain := range moves {
				newMoves = append(newMoves, append(slices.Clone(existingMoveChain), moreMoveChain...))
			}
		}
		for i := range newMoves {
			newMoves[i] = append(newMoves[i], press)
		}
		moves = newMoves
	}

	return moves
}

func (mf *moveFinder) pressArrows(arrowKey string, keypadIndex int) [][]move {
	fromTo := fromToStringPair{
		from: mf.currentArrowKeypadKeys[keypadIndex],
		to:   arrowKey,
	}
	arrowMoveCandidates, ok := mf.arrowKeypadMoves[fromTo]
	if !ok {
		panic("Unrecognized keys")
	}
	mf.currentArrowKeypadKeys[keypadIndex] = arrowKey

	return slices.Clone(arrowMoveCandidates)
}

func (mf *moveFinder) pressArrowsMultipleTimes(arrowKeys []string, keypadIndex int) [][]move {
	if mf.currentArrowKeypadKeys[keypadIndex] != "A" {
		panic("Should always start on A")
	}
	initialMoves := mf.pressArrows(arrowKeys[0], keypadIndex)
	moves := make([][]move, len(initialMoves))
	for i := range initialMoves {
		moves[i] = slices.Clone(initialMoves[i])
		moves[i] = append(moves[i], press)
	}
	for _, m := range arrowKeys[1:] {
		newMoves := make([][]move, 0)
		moreMoves := mf.pressArrows(m, keypadIndex)
		for _, moreMoveChain := range moreMoves {
			for _, existingMoveChain := range moves {
				newMoves = append(newMoves, append(slices.Clone(existingMoveChain), moreMoveChain...))
			}
		}
		for i := range newMoves {
			newMoves[i] = append(newMoves[i], press)
		}
		moves = newMoves
	}

	return moves
}

func (mf *moveFinder) findOptimalMoveOnArrowKeypad(
	keypadIndex int,
	// candidateMoves are the move strings coming from lower levels.
	moves string,
) int {
	if cost, ok := mf.optimalMoves[keypadIndex][moves]; ok {
		return cost
	}
	if keypadIndex == mf.numArrowKeypads-1 {
		mf.optimalMoves[keypadIndex][moves] = len(moves)
		return len(moves)
	}
	// Chunk moves smaller so we can cache with finer granularity.
	subMoves := lo.Filter(strings.SplitAfter(moves, "A"), func(s string, _ int) bool {
		if s == "" {
			return false
		}
		return true
	})

	totalCost := 0
	for _, subMove := range subMoves {
		chunkedMoves := lo.ChunkString(subMove, 1)
		nextLevelMoves := mf.pressArrowsMultipleTimes(chunkedMoves, keypadIndex+1)
		nextLevelMoveStrings := lo.Map(nextLevelMoves, func(moves []move, _ int) string {
			return lo.Reduce(moves, func(agg string, item move, _ int) string {
				return agg + item.toString()
			}, "")
		})
		nextMoveCosts := make([]int, len(nextLevelMoveStrings))
		for i, moveString := range nextLevelMoveStrings {
			nextMoveCosts[i] = mf.findOptimalMoveOnArrowKeypad(keypadIndex+1, moveString)
		}
		totalCost += lo.Min(nextMoveCosts)
	}
	mf.optimalMoves[keypadIndex][moves] = totalCost
	return totalCost
}

func (mf *moveFinder) deriveNumericPresses(desiredNumericKeys []string) int {
	// Handle first level of presses.
	numericMoveCandidates := mf.pressNumericMultipleTimes(desiredNumericKeys)
	firstArrowPresses := make([][]move, 0)
	for _, nmc := range numericMoveCandidates {
		nmcString := lo.Map(nmc, func(m move, _ int) string {
			return m.toString()
		})
		someFirstArrowPresses := mf.pressArrowsMultipleTimes(nmcString, 0)
		firstArrowPresses = append(firstArrowPresses, someFirstArrowPresses...)
	}
	firstArrowPressesAsStrings := lo.Map(firstArrowPresses, func(moves []move, _ int) string {
		return lo.Reduce(moves, func(agg string, item move, _ int) string {
			return agg + item.toString()
		}, "")
	})
	lengths := make([]int, len(firstArrowPressesAsStrings))
	for i := range lengths {
		lengths[i] = mf.findOptimalMoveOnArrowKeypad(0, firstArrowPressesAsStrings[i])
	}
	return lo.Min(lengths)
}

func (mf *moveFinder) pressForKeys(presses []string) int {
	return mf.deriveNumericPresses(presses)
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	keyPresses := make([][]string, 0)
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}

		keyPresses = append(keyPresses, handleLine(scanner.Text()))
	}

	mf := newMoveFinder(2)

	score := 0
	for i := range keyPresses {
		moveCost := mf.pressForKeys(keyPresses[i])
		code := ""
		for _, s := range keyPresses[i] {
			code += s
		}
		code, ok := strings.CutSuffix(code, "A")
		if !ok {
			panic("Should always have A suffix!")
		}
		codeNum, err := strconv.Atoi(code)
		if err != nil {
			panic(err)
		}
		score += moveCost * codeNum
	}

	fmt.Println(score)

	// Part 2
	mf = newMoveFinder(25)

	score = 0
	for i := range keyPresses {
		moveCost := mf.pressForKeys(keyPresses[i])
		code := ""
		for _, s := range keyPresses[i] {
			code += s
		}
		code, ok := strings.CutSuffix(code, "A")
		if !ok {
			panic("Should always have A suffix!")
		}
		codeNum, err := strconv.Atoi(code)
		if err != nil {
			panic(err)
		}
		score += moveCost * codeNum
	}

	// 9806252 too low
	fmt.Println(score)
}
