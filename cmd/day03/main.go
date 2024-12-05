package main

import (
	"bufio"
	_ "embed"
	"log"
	"regexp"
	"strconv"
	"strings"
)

//go:embed input
var input string

type mul struct {
	lhs int
	rhs int
}

var mulRegex = regexp.MustCompile(`mul\((\d+),(\d+)\)|do\(\)|don't\(\)`)

type inputHandler struct {
	do bool
}

func (ih *inputHandler) handleLine(line string) ([]mul, []mul) {
	muls := make([]mul, 0)
	filteredMuls := make([]mul, 0)

	matches := mulRegex.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		println(match[0])
		if match[0] == "do()" {
			ih.do = true
		} else if match[0] == "don't()" {
			ih.do = false
		} else {
			lhs, err := strconv.Atoi(match[1])
			if err != nil {
				log.Fatal(err)
			}
			rhs, err := strconv.Atoi(match[2])
			if err != nil {
				log.Fatal(err)
			}
			muls = append(muls, mul{lhs: lhs, rhs: rhs})
			if ih.do {
				filteredMuls = append(filteredMuls, mul{lhs: lhs, rhs: rhs})
			}
		}
	}

	return muls, filteredMuls
}

func main() {
	muls1 := make([]mul, 0)
	muls2 := make([]mul, 0)

	ih := &inputHandler{do: true}

	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		if scanner.Text() == "" {
			// Skip blank lines.
			continue
		}
		part1Muls, part2Muls := ih.handleLine(scanner.Text())
		muls1 = append(muls1, part1Muls...)
		muls2 = append(muls2, part2Muls...)
	}

	count := 0
	for _, mul := range muls1 {
		count += mul.lhs * mul.rhs
	}
	println(count)

	count = 0
	for _, mul := range muls2 {
		count += mul.lhs * mul.rhs
	}
	println(count)

}
