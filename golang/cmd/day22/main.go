package main

import (
	"bufio"
	_ "embed"
	"strconv"
	"strings"
)

//go:embed input
var input string

func handleLine(line string) int {
	num, err := strconv.Atoi(line)
	if err != nil {
		panic(err)
	}
	return num
}

type secretNum int

func (s secretNum) mix(n secretNum) secretNum {
	return s ^ n
}

func (s secretNum) prune() secretNum {
	return s % 16777216
}

func (s secretNum) iterate() secretNum {
	newNum := s
	newNum = s.mix(s * 64)
	newNum = newNum.prune()
	newNum = newNum.mix(newNum / 32)
	newNum = newNum.prune()
	newNum = newNum.mix(newNum * 2048)
	newNum = newNum.prune()
	return newNum
}

type secretNumFinder struct {
	nextLookup map[secretNum]secretNum
}

func newSecretNumFinder() secretNumFinder {
	return secretNumFinder{
		nextLookup: make(map[secretNum]secretNum),
	}
}

func (snf *secretNumFinder) findNext(sn secretNum) secretNum {
	next, found := snf.nextLookup[sn]
	if !found {
		next = sn.iterate()
		snf.nextLookup[sn] = next
	}
	return next
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	secretNums := make([]secretNum, 0)
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}

		secretNums = append(secretNums, secretNum(handleLine(scanner.Text())))
	}

	snf := newSecretNumFinder()

	sum := 0
	for _, sn := range secretNums {
		for range 2000 {
			sn = snf.findNext(sn)
		}
		sum += int(sn)
	}
	println(sum)
}
