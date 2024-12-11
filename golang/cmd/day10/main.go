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
	chars := lo.ChunkString(line, 1)
	ints := make([]int, len(chars))
	var err error
	for i := range chars {
		if chars[i] == "." {
			// Allow reading of non-complete maps.
			ints[i] = -1
			continue
		}
		ints[i], err = strconv.Atoi(chars[i])
		if err != nil {
			log.Fatal(err)
		}
	}
	return ints
}

type coordinate struct {
	row, col int
}

type trailWalk struct {
	start coordinate
	steps []coordinate
}

func (tw *trailWalk) eq(other trailWalk) bool {
	if tw.start != other.start {
		return false
	}
	if len(tw.steps) != len(other.steps) {
		return false
	}
	for i := range tw.steps {
		if tw.steps[i] != other.steps[i] {
			return false
		}
	}
	return true
}

func (tw *trailWalk) isComplete(gm gameMap) bool {
	lastStep := tw.steps[len(tw.steps)-1]
	if gm.rawMap[lastStep.row][lastStep.col] == 9 {
		return true
	}

	return false
}

func (tw *trailWalk) end() coordinate {
	if len(tw.steps) == 0 {
		log.Fatal("Shouldn't be called on empty walk")
	}
	return tw.steps[len(tw.steps)-1]
}

func (tw *trailWalk) findNextSteps(gm gameMap) []trailWalk {
	lastStep := tw.start
	if len(tw.steps) > 0 {
		lastStep = tw.steps[len(tw.steps)-1]
	}
	lastStepHeight := gm.rawMap[lastStep.row][lastStep.col]
	trailWalks := make([]trailWalk, 0)

	north := coordinate{row: lastStep.row - 1, col: lastStep.col}
	east := coordinate{row: lastStep.row, col: lastStep.col + 1}
	south := coordinate{row: lastStep.row + 1, col: lastStep.col}
	west := coordinate{row: lastStep.row, col: lastStep.col - 1}

	if gm.inBounds(north) && gm.rawMap[north.row][north.col] == lastStepHeight+1 {
		steps := make([]coordinate, len(tw.steps))
		copy(steps, tw.steps)
		steps = append(steps, north)
		trailWalks = append(trailWalks, trailWalk{
			start: tw.start,
			steps: steps,
		})
	}

	if gm.inBounds(east) && gm.rawMap[east.row][east.col] == lastStepHeight+1 {
		steps := make([]coordinate, len(tw.steps))
		copy(steps, tw.steps)
		steps = append(steps, east)
		trailWalks = append(trailWalks, trailWalk{
			start: tw.start,
			steps: steps,
		})
	}

	if gm.inBounds(south) && gm.rawMap[south.row][south.col] == lastStepHeight+1 {
		steps := make([]coordinate, len(tw.steps))
		copy(steps, tw.steps)
		steps = append(steps, south)
		trailWalks = append(trailWalks, trailWalk{
			start: tw.start,
			steps: steps,
		})
	}

	if gm.inBounds(west) && gm.rawMap[west.row][west.col] == lastStepHeight+1 {
		steps := make([]coordinate, len(tw.steps))
		copy(steps, tw.steps)
		steps = append(steps, west)
		trailWalks = append(trailWalks, trailWalk{
			start: tw.start,
			steps: steps,
		})
	}

	return trailWalks
}

type gameMap struct {
	rawMap           [][]int
	startCoordinates []coordinate
}

func (gm *gameMap) inBounds(c coordinate) bool {
	if c.row < 0 || c.col < 0 {
		return false
	}
	if c.row >= len(gm.rawMap) || c.col >= len(gm.rawMap[0]) {
		return false
	}
	return true
}

func (gm *gameMap) findTrails() []trailWalk {
	completeWalks := make([]trailWalk, 0)
	inProgressWalks := make([]trailWalk, len(gm.startCoordinates))
	for i := range gm.startCoordinates {
		inProgressWalks[i] = trailWalk{
			start: gm.startCoordinates[i],
			steps: []coordinate{},
		}
	}

	for len(inProgressWalks) > 0 {
		nextWalksToCheck := make([]trailWalk, 0)
		for i := range inProgressWalks {
			nextSteps := inProgressWalks[i].findNextSteps(*gm)
			for j := range nextSteps {
				if nextSteps[j].isComplete(*gm) {
					completeWalks = append(completeWalks, nextSteps[j])
				} else {
					nextWalksToCheck = append(nextWalksToCheck, nextSteps[j])
				}
			}
		}
		inProgressWalks = nextWalksToCheck
	}
	return completeWalks
}

func createMap(rawMap [][]int) gameMap {
	startCoordinates := make([]coordinate, 0)
	for row := range rawMap {
		for col := range rawMap[row] {
			if rawMap[row][col] == 0 {
				startCoordinates = append(startCoordinates, coordinate{row, col})
			}
		}
	}
	return gameMap{
		rawMap:           rawMap,
		startCoordinates: startCoordinates,
	}
}

type startAndEnd struct {
	start coordinate
	end   coordinate
}

func filterWalks(walks []trailWalk) []trailWalk {
	filtered := make([]trailWalk, 0)
	startAndEndMap := make(map[startAndEnd]struct{})
	for i := range walks {
		sae := startAndEnd{
			start: walks[i].start,
			end:   walks[i].end(),
		}
		_, ok := startAndEndMap[sae]
		if !ok {
			filtered = append(filtered, walks[i])
			startAndEndMap[sae] = struct{}{}
		}
	}
	return filtered
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	rawMap := make([][]int, 0)
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		mapLine := handleLine(scanner.Text())
		rawMap = append(rawMap, mapLine)
	}

	gm := createMap(rawMap)
	trails := gm.findTrails()
	filteredTrails := filterWalks(trails)
	trailScores := make(map[coordinate]int)
	for i := range filteredTrails {
		trailScores[filteredTrails[i].start] += 1
	}
	var sum int
	for k, v := range trailScores {
		println(k.row, ",", k.col, ": ", v)
		sum += v
	}
	println(sum)

	trailScores = make(map[coordinate]int)
	for i := range trails {
		trailScores[trails[i].start] += 1
	}
	sum = 0
	for k, v := range trailScores {
		println(k.row, ",", k.col, ": ", v)
		sum += v
	}
	println(sum)
}
