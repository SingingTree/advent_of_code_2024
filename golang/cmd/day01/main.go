package main

import (
	"bufio"
	_ "embed"
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

//go:embed input
var input string

type listHolder struct {
	lhsNumbers []int
	rhsNumbers []int
}

func (holder *listHolder) addEntries(lhs int, rhs int) {
	holder.lhsNumbers = append(holder.lhsNumbers, lhs)
	holder.rhsNumbers = append(holder.rhsNumbers, rhs)
}

func (holder *listHolder) getDifferences() int {
	slices.Sort(holder.lhsNumbers)
	slices.Sort(holder.rhsNumbers)

	differences := 0

	for i := range holder.lhsNumbers {
		lhs := holder.lhsNumbers[i]
		rhs := holder.rhsNumbers[i]
		if lhs > rhs {
			differences += lhs - rhs
		} else {
			differences += rhs - lhs
		}
	}
	return differences
}

func (holder *listHolder) getSimilarityScore() int {
	rhsCounts := lo.CountValues(holder.rhsNumbers)
	score := 0

	for _, lhs := range holder.lhsNumbers {
		score += lhs * rhsCounts[lhs]
	}

	return score
}

func handleLine(line string, holder *listHolder) {
	tokens := strings.Split(line, "   ")
	lhs, err := strconv.Atoi(tokens[0])
	if err != nil {
		log.Fatal(err)
	}
	rhs, err := strconv.Atoi(tokens[1])
	if err != nil {
		log.Fatal(err)
	}
	holder.addEntries(lhs, rhs)
}

func main() {
	holder := listHolder{
		lhsNumbers: make([]int, 0),
		rhsNumbers: make([]int, 0),
	}
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		if scanner.Text() == "" {
			// Skip blank lines.
			continue
		}
		handleLine(scanner.Text(), &holder)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Answer to part 1.
	println(holder.getDifferences())

	// Answer to part 2.
	println(holder.getSimilarityScore())
}
