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

type equation struct {
	target        int
	candidateNums []int
}

func (eq *equation) checkForSolution() bool {
	solutionCandidates := make([]int, 0)

	solutionCandidates = append(solutionCandidates, eq.candidateNums[0])
	for i := 1; i < len(eq.candidateNums); i++ {
		newSolutionCandidates := make([]int, 0)
		for _, solutionCandidate := range solutionCandidates {
			newSolutionCandidates = append(newSolutionCandidates, solutionCandidate+eq.candidateNums[i])
			newSolutionCandidates = append(newSolutionCandidates, solutionCandidate*eq.candidateNums[i])
		}
		solutionCandidates = newSolutionCandidates
	}

	if slices.Contains(solutionCandidates, eq.target) {
		return true
	}
	return false
}

func (eq *equation) checkForSolutionWithConcat() bool {
	solutionCandidates := make([]int, 0)

	solutionCandidates = append(solutionCandidates, eq.candidateNums[0])
	for i := 1; i < len(eq.candidateNums); i++ {
		newSolutionCandidates := make([]int, 0)
		for _, solutionCandidate := range solutionCandidates {
			newSolutionCandidates = append(newSolutionCandidates, solutionCandidate+eq.candidateNums[i])
			newSolutionCandidates = append(newSolutionCandidates, solutionCandidate*eq.candidateNums[i])
			solutionCandidateStr := strconv.Itoa(solutionCandidate)
			candidateNumStr := strconv.Itoa(eq.candidateNums[i])
			concatStr := solutionCandidateStr + candidateNumStr
			concatInt, err := strconv.Atoi(concatStr)
			if err != nil {
				log.Fatal(err)
			}
			newSolutionCandidates = append(newSolutionCandidates, concatInt)

		}
		solutionCandidates = newSolutionCandidates
	}

	if slices.Contains(solutionCandidates, eq.target) {
		return true
	}
	return false
}

func handleLine(line string) equation {
	tokens := strings.Fields(line)

	// Trim colon from first string.
	tokens[0] = strings.Trim(tokens[0], ":")

	target, err := strconv.Atoi(tokens[0])
	if err != nil {
		log.Fatal(err)
	}

	candidateNums := make([]int, len(tokens)-1)
	for i := 1; i < len(tokens); i++ {
		candidateNum, err := strconv.Atoi(tokens[i])
		if err != nil {
			log.Fatal(err)
		}
		candidateNums[i-1] = candidateNum
	}

	return equation{
		target:        target,
		candidateNums: candidateNums,
	}
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))
	equations := make([]equation, 0)

	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		equations = append(equations, handleLine(scanner.Text()))
	}

	solutionSumPart1 := 0
	solutionSumPart2 := 0
	for _, equation := range equations {
		if equation.checkForSolution() {
			solutionSumPart1 += equation.target
		}
		if equation.checkForSolutionWithConcat() {
			solutionSumPart2 += equation.target
		}
	}

	println(solutionSumPart1)
	println(solutionSumPart2)
}
