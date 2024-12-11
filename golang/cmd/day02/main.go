package main

import (
	"bufio"
	_ "embed"
	"log"
	"slices"
	"strconv"
	"strings"
)

//go:embed input
var input string

type levelHandler struct {
	levelsList [][]int
}

func (handler *levelHandler) addLevels(levels []int) {
	handler.levelsList = append(handler.levelsList, levels)
}

func isSafe(levels []int) bool {
	is_increasing := false
	is_decreasing := false

	prev := levels[0]
	for _, level := range levels[1:] {
		if prev == level {
			// Must change
			return false
		}
		if prev < level {
			is_increasing = true
			if level-prev > 3 {
				// Difference is too great.
				return false
			}
		}
		if prev > level {
			is_decreasing = true
			if prev-level > 3 {
				return false
			}
		}
		if is_increasing && is_decreasing {
			return false
		}

		prev = level
	}

	return true
}

func isSafeWithLevelModulator(levels []int) bool {
	if isSafe(levels) {
		return true
	}
	for i := range levels {
		newLevels := slices.Clone(levels)
		newLevels = slices.Delete(newLevels, i, i+1)
		if isSafe(newLevels) {
			return true
		}
	}

	return false
}

func (handler *levelHandler) getSafeCount() int {
	count := 0
	for _, levels := range handler.levelsList {
		if isSafe(levels) {
			count += 1
		}
	}

	return count
}

func (handler *levelHandler) getSafeCountWithModulator() int {
	count := 0
	for _, levels := range handler.levelsList {
		if isSafeWithLevelModulator(levels) {
			count += 1
		}
	}

	return count
}

func handleLine(line string, handler *levelHandler) {
	tokens := strings.Split(line, " ")
	levels := make([]int, len(tokens))
	for i, token := range tokens {
		reading, err := strconv.Atoi(token)
		if err != nil {
			log.Fatal(err)
		}

		levels[i] = reading
	}
	handler.addLevels(levels)
}

func main() {
	handler := levelHandler{
		make([][]int, 0),
	}

	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		if scanner.Text() == "" {
			// Skip blank lines.
			continue
		}
		handleLine(scanner.Text(), &handler)
	}

	println(handler.getSafeCount())
	println(handler.getSafeCountWithModulator())
}
