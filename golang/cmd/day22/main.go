package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"gonum.org/v1/gonum/stat/combin"
	"slices"
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

func (s secretNum) price() int {
	return int(s) % 10
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

type priceChangeQuad struct {
	first  int
	second int
	third  int
	fourth int
}

var possiblePriceChanges = []int{
	-9, -8, -7, -6, -5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
}

func generateAllPossibleSequences() []priceChangeQuad {
	indices := combin.Cartesian([]int{19, 19, 19, 19})
	sequences := make([]priceChangeQuad, len(indices))
	for i := range indices {
		sequences[i] = priceChangeQuad{
			first:  possiblePriceChanges[indices[i][0]],
			second: possiblePriceChanges[indices[i][1]],
			third:  possiblePriceChanges[indices[i][2]],
			fourth: possiblePriceChanges[indices[i][3]],
		}
	}

	return sequences
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

	// Part 2
	prices := make([][]int, len(secretNums))
	priceChanges := make([][]int, len(secretNums))
	for i, sn := range secretNums {
		prices[i] = make([]int, 2001)
		prices[i][0] = sn.price()
		priceChanges[i] = make([]int, 2001)
		priceChanges[i][0] = -100

		for j := range 2000 {
			sn = snf.findNext(sn)
			prices[i][j+1] = sn.price()
			priceChanges[i][j+1] = prices[i][j+1] - prices[i][j]
		}
	}

	containsQuadMap := make([]map[priceChangeQuad]int, len(priceChanges))
	for i := range priceChanges {
		containsQuadMap[i] = make(map[priceChangeQuad]int)
		// Start this loop at 4 instead of 3, as price change 0 is invalid, so we skip it.
		for j := 4; j < len(priceChanges[i]); j += 1 {
			pcq := priceChangeQuad{
				first:  priceChanges[i][j-3],
				second: priceChanges[i][j-2],
				third:  priceChanges[i][j-1],
				fourth: priceChanges[i][j],
			}
			if _, ok := containsQuadMap[i][pcq]; ok {
				// We already have this quad.
				continue
			}
			containsQuadMap[i][pcq] = j
		}
	}

	sequences := generateAllPossibleSequences()
	payoutPerSequence := make([]int, len(sequences))
	for i := range sequences {
		for j := range prices {
			if idx, ok := containsQuadMap[j][sequences[i]]; ok {
				payoutPerSequence[i] += prices[j][idx]
			}
		}
	}
	maxPayout := slices.Max(payoutPerSequence)
	index := slices.Index(payoutPerSequence, maxPayout)
	fmt.Println(sequences[index])
	fmt.Println(maxPayout)
}
