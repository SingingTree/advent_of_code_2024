package main

import (
	"bufio"
	_ "embed"
	"github.com/samber/lo"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

//go:embed input
var input string

const (
	buttonALine = iota
	buttonBLine
	prizeLine
)

const (
	buttonACost = 3
	buttonBCost = 1
)

var buttonARegex = regexp.MustCompile(`Button A: X\+(\d+), Y\+(\d+)`)
var buttonBRegex = regexp.MustCompile(`Button B: X\+(\d+), Y\+(\d+)`)
var prizeRegex = regexp.MustCompile(`Prize: X=(\d+), Y=(\d+)`)

// Returns (x, y, lineType)
func handleLine(line string) (int, int, int) {
	matches := buttonARegex.FindStringSubmatch(line)
	if matches != nil {
		x, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Fatal(err)
		}
		y, err := strconv.Atoi(matches[2])
		if err != nil {
			log.Fatal(err)
		}
		return x, y, buttonALine
	}
	matches = buttonBRegex.FindStringSubmatch(line)
	if matches != nil {
		x, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Fatal(err)
		}
		y, err := strconv.Atoi(matches[2])
		if err != nil {
			log.Fatal(err)
		}
		return x, y, buttonBLine
	}
	matches = prizeRegex.FindStringSubmatch(line)
	if matches != nil {
		x, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Fatal(err)
		}
		y, err := strconv.Atoi(matches[2])
		if err != nil {
			log.Fatal(err)
		}
		return x, y, prizeLine
	}
	return -1, -1, -1
}

type coordinate struct {
	x int
	y int
}

func (c *coordinate) add(other coordinate) coordinate {
	return coordinate{c.x + other.x, c.y + other.y}
}

func (c *coordinate) sub(other coordinate) coordinate {
	return coordinate{c.x - other.x, c.y - other.y}
}

func (c *coordinate) mul(x int) coordinate {
	return coordinate{c.x * x, c.y * x}
}

type clawMachine struct {
	buttonAMove   coordinate
	buttonBMove   coordinate
	prizeLocation coordinate
}

func (cm *clawMachine) subGame(c coordinate) clawMachine {
	prizeLocation := cm.prizeLocation.sub(c)
	return clawMachine{
		buttonAMove:   cm.buttonAMove,
		buttonBMove:   cm.buttonBMove,
		prizeLocation: prizeLocation,
	}
}

type clawMachineMoveChain struct {
	aPressCount         int
	bPressCount         int
	currentClawPosition coordinate
	cost                int
}

type presses struct {
	aPresses int
	bPresses int
}

func (cm *clawMachine) bruteForceSolutions(maxPresses int) []clawMachineMoveChain {
	viableSolutions := make([]clawMachineMoveChain, 0)

	// BFS.
	remainingDepth := maxPresses - 1
	frontier := []clawMachineMoveChain{
		{aPressCount: 1, bPressCount: 0, currentClawPosition: cm.buttonAMove, cost: buttonACost},
		{aPressCount: 0, bPressCount: 1, currentClawPosition: cm.buttonBMove, cost: buttonBCost},
	}
	// Test just in case 1 button press completes any problems.
	for i := range frontier {
		if frontier[i].currentClawPosition == cm.prizeLocation {
			viableSolutions = append(viableSolutions, frontier[i])
		}
	}

	// Do actual BFS.
	for remainingDepth > 0 {
		newFrontier := make([]clawMachineMoveChain, 0)
		for _, chain := range frontier {
			aPress := clawMachineMoveChain{
				aPressCount:         chain.aPressCount + 1,
				bPressCount:         chain.bPressCount,
				currentClawPosition: chain.currentClawPosition.add(cm.buttonAMove),
				cost:                chain.cost + buttonACost,
			}
			bPress := clawMachineMoveChain{
				aPressCount:         chain.aPressCount,
				bPressCount:         chain.bPressCount + 1,
				currentClawPosition: chain.currentClawPosition.add(cm.buttonBMove),
				cost:                chain.cost + buttonBCost,
			}
			newFrontier = append(newFrontier, aPress, bPress)
		}
		// Remove equivalent moves.
		newFrontier = lo.UniqBy(newFrontier, func(chain clawMachineMoveChain) presses {
			return presses{chain.aPressCount, chain.bPressCount}
		})
		for _, chain := range newFrontier {
			if chain.currentClawPosition == cm.prizeLocation {
				viableSolutions = append(viableSolutions, chain)
			}
		}
		frontier = newFrontier
		remainingDepth -= 1
	}
	return viableSolutions
}

func canReachGoal(aMove, bMove, goal coordinate) (int, int) {
	// Apply Cramer's rule.
	determinant := aMove.x*bMove.y - aMove.y*bMove.x

	if determinant == 0 {
		log.Fatal("0 det indicates parallel vecs, we don't expect that!")
	}

	aPressesNeeded := float64(goal.x*bMove.y-goal.y*bMove.x) / float64(determinant)
	bPressesNeeded := float64(goal.y*aMove.x-goal.x*aMove.y) / float64(determinant)

	// Negative checks probably not needed here, but it won't hurt.
	if aPressesNeeded != math.Trunc(aPressesNeeded) || aPressesNeeded < 0 {
		return -1, -1
	}
	if bPressesNeeded != math.Trunc(bPressesNeeded) || bPressesNeeded < 0 {
		return -1, -1
	}
	// There is some combination of a and b that reach goal.
	return int(aPressesNeeded), int(bPressesNeeded)
}

func (cm *clawMachine) betterSearch() *clawMachineMoveChain {
	//Filter if we can even reach the solution.
	a, b := canReachGoal(cm.buttonAMove, cm.buttonBMove, cm.prizeLocation)
	if a == -1 || b == -1 {
		return nil
	}

	// Solutions are unique, why?
	solutionCandidate := clawMachineMoveChain{
		aPressCount:         a,
		bPressCount:         b,
		currentClawPosition: cm.prizeLocation,
		cost:                buttonACost*a + buttonBCost*b,
	}

	return &solutionCandidate
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	nextExpectedLine := buttonALine
	var buttonAMove coordinate
	var buttonBMove coordinate
	var prizeLocation coordinate
	clawMachines := make([]clawMachine, 0)
	for scanner.Scan() {

		if scanner.Text() == "" {
			continue
		}
		x, y, lineType := handleLine(scanner.Text())
		if lineType != nextExpectedLine {
			log.Fatalf("Line type not as expected! Expected %v, got %v", nextExpectedLine, lineType)
		}
		c := coordinate{
			x: x,
			y: y,
		}
		if nextExpectedLine == buttonALine {
			buttonAMove = c
		}
		if nextExpectedLine == buttonBLine {
			buttonBMove = c
		}
		if nextExpectedLine == prizeLine {
			prizeLocation = c
			clawMachines = append(clawMachines, clawMachine{
				buttonAMove:   buttonAMove,
				buttonBMove:   buttonBMove,
				prizeLocation: prizeLocation,
			})
		}
		nextExpectedLine += 1
		if nextExpectedLine > prizeLine {
			nextExpectedLine = buttonALine
		}
	}

	// Part 1

	minCost := 0
	for _, cm := range clawMachines {
		solutions := cm.bruteForceSolutions(200)
		if len(solutions) > 0 {
			localMin := lo.MinBy(solutions, func(a clawMachineMoveChain, b clawMachineMoveChain) bool {
				return a.cost < b.cost
			})
			minCost += localMin.cost
		}
	}
	println(minCost)

	// Part 2

	harderGames := make([]clawMachine, len(clawMachines))
	for i, cm := range clawMachines {
		harderGames[i] = clawMachine{
			cm.buttonAMove,
			cm.buttonBMove,
			coordinate{
				x: cm.prizeLocation.x + 10_000_000_000_000,
				y: cm.prizeLocation.y + 10_000_000_000_000,
			},
		}
	}
	minCost = 0
	for i, cm := range harderGames {

		solution := cm.betterSearch()

		if solution == nil {
			println(i, ": nil")
			continue
		}
		println(i, ": ", solution.cost)
		minCost += solution.cost
	}
	println(minCost)
}
