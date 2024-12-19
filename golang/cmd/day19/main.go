package main

import (
	"bufio"
	_ "embed"
	"github.com/samber/lo"
	"strings"
)

//go:embed input
var input string

type towel struct {
	colors string
}

func handleLineTowel(line string) []towel {
	towels := make([]towel, 0)
	towelColorStrings := lo.Map(strings.Split(line, ","), func(item string, index int) string {
		return strings.TrimSpace(item)
	})
	for _, colorString := range towelColorStrings {
		towels = append(towels, towel{colors: colorString})
	}

	return towels
}

func handlePatternLineTowel(line string) string {
	return strings.TrimSpace(line)
}

type towelSolutionFinder struct {
	towels                      []towel
	knownSolvableCombinations   map[string]struct{}
	knownUnsolvableCombinations map[string]struct{}
	combinationToSolutionCount  map[string]int
}

func newTowelSolutionFinder(towels []towel) towelSolutionFinder {
	return towelSolutionFinder{
		towels:                      towels,
		knownSolvableCombinations:   make(map[string]struct{}),
		knownUnsolvableCombinations: make(map[string]struct{}),
		combinationToSolutionCount:  make(map[string]int),
	}
}

type patternSearchState struct {
	patternSolved string
	patternLeft   string
}

// DFS iterative can represent finder.
func (tsf *towelSolutionFinder) canRepresent(chain string) bool {
	if _, ok := tsf.knownSolvableCombinations[chain]; ok {
		return true
	}
	if _, ok := tsf.knownUnsolvableCombinations[chain]; ok {
		return false
	}

	frontier := []patternSearchState{
		{patternSolved: "", patternLeft: chain},
	}
	for len(frontier) > 0 {
		state, _ := lo.Last(frontier)
		frontier = frontier[:len(frontier)-1]
		if _, ok := tsf.knownSolvableCombinations[state.patternLeft]; ok {
			return true
		}
		if _, ok := tsf.knownUnsolvableCombinations[state.patternLeft]; ok {
			continue
		}
		noPrefixMatch := true
		for _, t := range tsf.towels {
			if strings.HasPrefix(state.patternLeft, t.colors) {
				noPrefixMatch = false
				patternLeft, _ := strings.CutPrefix(state.patternLeft, t.colors)
				if _, ok := tsf.knownSolvableCombinations[patternLeft]; ok {
					return true
				}
				patternSolved := state.patternSolved + t.colors
				tsf.knownSolvableCombinations[patternSolved] = struct{}{}
				if len(patternLeft) == 0 {
					return true
				}
				frontier = append(frontier, patternSearchState{
					patternSolved, patternLeft,
				})
			}
		}
		if noPrefixMatch {
			tsf.knownUnsolvableCombinations[state.patternLeft] = struct{}{}
		}
	}

	return false
}

// DFS recursive count finder.
func (tsf *towelSolutionFinder) numRepresentations(chain string) int {
	if count, ok := tsf.combinationToSolutionCount[chain]; ok {
		return count
	}

	count := 0
	for _, t := range tsf.towels {
		if strings.HasPrefix(chain, t.colors) {
			patternLeft, _ := strings.CutPrefix(chain, t.colors)
			if len(patternLeft) == 0 {
				count += 1
				continue
			}
			count += tsf.numRepresentations(patternLeft)
		}
	}

	tsf.combinationToSolutionCount[chain] = count
	return count
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	towels := make([]towel, 0)
	patterns := make([]string, 0)
	parsedTowels := false
	for scanner.Scan() {
		if scanner.Text() == "" {
			parsedTowels = true
			continue
		}
		if !parsedTowels {
			towels = append(towels, handleLineTowel(scanner.Text())...)
			continue
		}
		patterns = append(patterns, handlePatternLineTowel(scanner.Text()))
	}

	// Part 1.
	tsf := newTowelSolutionFinder(towels)
	count := 0
	for _, pattern := range patterns {
		if tsf.canRepresent(pattern) {
			count++
		}
	}
	println(count)

	// Part 2.
	count = 0
	for _, pattern := range patterns {
		count += tsf.numRepresentations(pattern)
	}
	println(count)
}
