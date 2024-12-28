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
	//movesCandidates := make([][]move, 0)
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

	// Filter out any moves that go via the unsfae coordinate.
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

	//sameRowAsUnsafe := c.row == unsafeCoordinate.row
	//sameColAsUnsafe := c.col == unsafeCoordinate.col
	//if sameRowAsUnsafe && sameColAsUnsafe {
	//	panic("Unreachable, this means we're on the unsafe coordinate already!")
	//}
	//if !sameColAsUnsafe {
	//	// Because we're not on an unsafe col, it's safe to do the row moves
	//	// first. This is because no amount of row moves will put us onto the
	//	// unsafe coordinate.
	//	moveCandidate := append(slices.Clone(rowMoves), colMoves...)
	//	movesCandidates = append(movesCandidates, moveCandidate)
	//}
	//if !sameRowAsUnsafe {
	//	// As above, since we're not on the unsafe row, it's safe for us to do
	//	// col moves before row moves.
	//	moveCandidate := append(slices.Clone(colMoves), rowMoves...)
	//	movesCandidates = append(movesCandidates, moveCandidate)
	//}
	//
	//return movesCandidates
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
	currentNumericKey           string
	currentFirstArrowKeypadKey  string
	currentSecondArrowKeypadKey string
	currentThirdArrowKeypadKey  string
	numericKeypadMoves          map[fromToStringPair][][]move
	arrowKeypadMoves            map[fromToStringPair][][]move
}

func newMoveFinder() moveFinder {
	return moveFinder{
		currentNumericKey:           "A",
		currentFirstArrowKeypadKey:  "A",
		currentSecondArrowKeypadKey: "A",
		currentThirdArrowKeypadKey:  "A",
		numericKeypadMoves:          getNumericKeypadMoveMap(),
		arrowKeypadMoves:            getArrowKeypadMoveMap(),
	}
}

func (mf *moveFinder) clone() moveFinder {
	return moveFinder{
		currentNumericKey:           mf.currentNumericKey,
		currentFirstArrowKeypadKey:  mf.currentFirstArrowKeypadKey,
		currentSecondArrowKeypadKey: mf.currentSecondArrowKeypadKey,
		currentThirdArrowKeypadKey:  mf.currentThirdArrowKeypadKey,
		numericKeypadMoves:          mf.numericKeypadMoves,
		arrowKeypadMoves:            mf.arrowKeypadMoves,
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

func (mf *moveFinder) pressFirstArrows(arrowKey string) [][]move {
	fromTo := fromToStringPair{
		from: mf.currentFirstArrowKeypadKey,
		to:   arrowKey,
	}
	arrowMoveCandidates, ok := mf.arrowKeypadMoves[fromTo]
	if !ok {
		panic("Unrecognized keys")
	}
	mf.currentFirstArrowKeypadKey = arrowKey

	return slices.Clone(arrowMoveCandidates)
}

func (mf *moveFinder) pressFirstArrowsMultipleTimes(arrowKeys []string) [][]move {
	if mf.currentFirstArrowKeypadKey != "A" {
		panic("Should always start on A")
	}
	initialMoves := mf.pressFirstArrows(arrowKeys[0])
	moves := make([][]move, len(initialMoves))
	for i := range initialMoves {
		moves[i] = slices.Clone(initialMoves[i])
		moves[i] = append(moves[i], press)
	}
	for _, m := range arrowKeys[1:] {
		newMoves := make([][]move, 0)
		moreMoves := mf.pressFirstArrows(m)
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

func (mf *moveFinder) pressSecondArrows(arrowKey string) [][]move {
	fromTo := fromToStringPair{
		from: mf.currentSecondArrowKeypadKey,
		to:   arrowKey,
	}
	arrowMoveCandidates, ok := mf.arrowKeypadMoves[fromTo]
	if !ok {
		panic("Unrecognized keys")
	}
	mf.currentSecondArrowKeypadKey = arrowKey

	return arrowMoveCandidates
}

func (mf *moveFinder) pressSecondArrowsMultipleTimes(arrowKeys []string) [][]move {
	if mf.currentSecondArrowKeypadKey != "A" {
		panic("Should always start on A")
	}
	initialMoves := mf.pressSecondArrows(arrowKeys[0])
	moves := make([][]move, len(initialMoves))
	for i := range initialMoves {
		moves[i] = slices.Clone(initialMoves[i])
		moves[i] = append(moves[i], press)
	}
	for _, m := range arrowKeys[1:] {
		newMoves := make([][]move, 0)
		moreMoves := mf.pressSecondArrows(m)
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

func (mf *moveFinder) pressThirdArrows(arrowKey string) [][]move {
	fromTo := fromToStringPair{
		from: mf.currentThirdArrowKeypadKey,
		to:   arrowKey,
	}
	arrowMoveCandidates, ok := mf.arrowKeypadMoves[fromTo]
	if !ok {
		panic("Unrecognized keys")
	}
	mf.currentThirdArrowKeypadKey = arrowKey

	return arrowMoveCandidates
}

func (mf *moveFinder) deriveNumericPresses(desiredNumericKeys []string) []move {
	numericMoveCandidates := mf.pressNumericMultipleTimes(desiredNumericKeys)
	firstArrowPresses := make([][]move, 0)
	for _, nmc := range numericMoveCandidates {
		nmcString := lo.Map(nmc, func(m move, _ int) string {
			return m.toString()
		})
		someFirstArrowPresses := mf.pressFirstArrowsMultipleTimes(nmcString)
		firstArrowPresses = append(firstArrowPresses, someFirstArrowPresses...)
	}
	shortestLen := len(firstArrowPresses[0])
	for _, p := range firstArrowPresses {
		if len(p) < shortestLen {
			shortestLen = len(p)
		}
	}
	firstArrowPresses = lo.Filter(firstArrowPresses, func(presses []move, _ int) bool {
		if len(presses) > shortestLen {
			return false
		}
		return true
	})
	secondArrowPresses := make([][]move, 0)
	for _, fap := range firstArrowPresses {
		fapString := lo.Map(fap, func(m move, _ int) string {
			return m.toString()
		})
		someSecondArrowPresses := mf.pressSecondArrowsMultipleTimes(fapString)
		secondArrowPresses = append(secondArrowPresses, someSecondArrowPresses...)
	}
	minSecondPresses := lo.MinBy(secondArrowPresses, func(a []move, b []move) bool {
		return len(a) < len(b)
	})
	return minSecondPresses
}

func (mf *moveFinder) pressForKeys(presses []string) []move {
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

	mf := newMoveFinder()

	score := 0
	for i := range keyPresses {
		moves := mf.pressForKeys(keyPresses[i])
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
		score += len(moves) * codeNum
	}

	fmt.Println(score)
}
