package main

import (
	"bufio"
	_ "embed"
	"github.com/samber/lo"
	"log"
	"strconv"
	"strings"
)

//go:embed input
var input string

func handleLine(line string) []int {
	stones := make([]int, 0)
	stoneNumStrings := strings.Fields(line)
	for i := range stoneNumStrings {
		stone, err := strconv.Atoi(stoneNumStrings[i])
		if err != nil {
			log.Fatal(err)
		}
		stones = append(stones, stone)
	}

	return stones
}

func blinkStone(stone int) []int {
	if stone == 0 {
		return []int{1}
	}
	stoneStr := strconv.Itoa(stone)
	if len(stoneStr)%2 == 0 {
		// We can get away with just operating on the string because we have ascii chars, but in general this is wonky.
		lhsStr := stoneStr[:len(stoneStr)/2]
		// lhs should require dropping, but do it just in case.
		for rune(lhsStr[0]) == '0' {
			lhsStr = lhsStr[1:]
		}
		lhs, err := strconv.Atoi(lhsStr)
		if err != nil {
			log.Fatal(err)
		}
		rhsStr := stoneStr[len(stoneStr)/2:]
		for len(rhsStr) > 1 && rune(rhsStr[0]) == '0' {
			rhsStr = rhsStr[1:]
		}
		rhs, err := strconv.Atoi(rhsStr)
		if err != nil {
			log.Fatal(err)
		}
		return []int{lhs, rhs}
	}

	return []int{stone * 2024}
}

func blinkToDepth(
	stones []int,
	searchDepth int,
	stoneToNext map[int][]int,
	stoneToMaxDepth map[int]int,
) {
	frontier := stones
	depthLeft := searchDepth
	for depthLeft > 0 {
		newFrontier := make([]int, 0)
		for i := range frontier {
			mapDepth, ok := stoneToMaxDepth[frontier[i]]
			if !ok {
				stoneToMaxDepth[frontier[1]] = depthLeft
			} else if mapDepth > depthLeft {
				// We've already seen this stone, skip it.
				continue
			}

			_, ok = stoneToNext[frontier[i]]
			if ok {
				// We've already seen this stone.
				continue
			}

			blinkedStones := blinkStone(frontier[i])
			stoneToNext[frontier[i]] = blinkedStones
			for _, blinkedStone := range blinkedStones {
				newFrontier = append(newFrontier, blinkedStone)
			}
			newFrontier = lo.Uniq(newFrontier)
		}
		frontier = newFrontier
		depthLeft -= 1
	}
}

type stoneAndDepthLeft struct {
	stone     int
	depthLeft int
}

func countChildren(
	stone int,
	depthLeft int,
	stoneToNext map[int][]int,
	stoneAtDepthLeftToChildCount map[stoneAndDepthLeft]int,
) int {
	if depthLeft == 1 {
		return len(stoneToNext[stone])
	}

	// Check for cached count.
	cachedCount, ok := stoneAtDepthLeftToChildCount[stoneAndDepthLeft{stone, depthLeft}]
	if ok {
		return cachedCount
	}

	// No cached count, get calculating.
	count := 0
	for _, child := range stoneToNext[stone] {
		count += countChildren(child, depthLeft-1, stoneToNext, stoneAtDepthLeftToChildCount)
	}
	// Cache our count before returning.
	stoneAtDepthLeftToChildCount[stoneAndDepthLeft{stone: stone, depthLeft: depthLeft}] = count
	return count
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	var stones []int
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		stones = handleLine(scanner.Text())
	}

	stoneToNext := make(map[int][]int)
	stoneToMaxDepth := make(map[int]int)

	blinkToDepth(stones, 75, stoneToNext, stoneToMaxDepth)

	stoneAndDepthLeftToChildCount := make(map[stoneAndDepthLeft]int)
	count := 0
	for _, stone := range stones {
		count += countChildren(stone, 25, stoneToNext, stoneAndDepthLeftToChildCount)
	}

	println(count)

	count = 0
	for _, stone := range stones {
		count += countChildren(stone, 75, stoneToNext, stoneAndDepthLeftToChildCount)
	}

	println(count)

	// -- part 2

}
