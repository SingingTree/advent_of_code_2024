package main

import (
	"bufio"
	_ "embed"
	"github.com/samber/lo"
	"log"
	"slices"
	"strconv"
	"strings"
)

//go:embed input
var input string

type ordering struct {
	before int
	after  int
}

type pageUpdate struct {
	pageNums []int
}

func (pu *pageUpdate) sort(orderings []ordering) {
	slices.SortFunc(pu.pageNums, func(lhs int, rhs int) int {
		if lo.Contains(orderings, ordering{lhs, rhs}) {
			return -1
		} else if lo.Contains(orderings, ordering{rhs, lhs}) {
			return 1
		} else {
			return 0
		}
	})
}

func (pu *pageUpdate) isSorted(orderings []ordering) bool {
	return slices.IsSortedFunc(pu.pageNums, func(lhs int, rhs int) int {
		if lo.Contains(orderings, ordering{lhs, rhs}) {
			return -1
		} else if lo.Contains(orderings, ordering{rhs, lhs}) {
			return 1
		} else {
			return 0
		}
	})
}

func (pu *pageUpdate) middlePage() int {
	pageCount := len(pu.pageNums)
	index := pageCount / 2
	return pu.pageNums[index]
}

func handleLineFirstSection(line string) (int, int) {
	numStrings := strings.Split(line, "|")

	lhs, err := strconv.Atoi(numStrings[0])
	if err != nil {
		log.Fatal(err)
	}
	rhs, err := strconv.Atoi(numStrings[1])
	if err != nil {
		log.Fatal(err)
	}

	return lhs, rhs
}

func handleLineSecondSection(line string) []int {
	pageNumStrings := strings.Split(line, ",")

	pageNums := make([]int, len(pageNumStrings))

	for i := range pageNumStrings {
		num, err := strconv.Atoi(pageNumStrings[i])
		if err != nil {
			log.Fatal(err)
		}
		pageNums[i] = num
	}

	return pageNums
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))
	section := 0

	orderings := make([]ordering, 0)
	pageNumUpdates := make([]pageUpdate, 0)

	for scanner.Scan() {
		if scanner.Text() == "" {
			section += 1
			continue
		}
		if section == 0 {
			lhs, rhs := handleLineFirstSection(scanner.Text())
			orderings = append(orderings, ordering{lhs, rhs})
		} else if section == 1 {
			nums := handleLineSecondSection(scanner.Text())

			pageNumUpdates = append(pageNumUpdates, pageUpdate{pageNums: nums})
		}
	}

	partOneSum := 0
	partTwoSum := 0
	for _, pageNumUpdate := range pageNumUpdates {
		if pageNumUpdate.isSorted(orderings) {
			println(pageNumUpdate.middlePage())
			partOneSum += pageNumUpdate.middlePage()
		} else {
			pageNumUpdate.sort(orderings)
			partTwoSum += pageNumUpdate.middlePage()
		}

	}
	println(partOneSum)
	println(partTwoSum)
}
