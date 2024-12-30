package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"github.com/samber/lo"
	"log"
	"net/http"
	_ "net/http/pprof"
	"slices"
	"strings"
)

//go:embed input
var input string

type gameSpace int

const (
	blank gameSpace = iota
	wall
	start
	finish
)

func gameSpaceFromString(s string) gameSpace {
	switch s {
	case ".":
		return blank
	case "#":
		return wall
	case "S":
		return start
	case "E":
		return finish
	default:
		panic(fmt.Sprintf("invalid game space string: %s", s))
	}
}

func (s gameSpace) toString() string {
	switch s {
	case blank:
		return "."
	case wall:
		return "#"
	case start:
		return "S"
	case finish:
		return "E"
	default:
		panic(fmt.Sprintf("invalid game space: %d", s))
	}
}

func handleLine(line string) []gameSpace {
	chars := lo.ChunkString(line, 1)
	gameSpaces := make([]gameSpace, len(chars))
	for i, char := range chars {
		gameSpaces[i] = gameSpaceFromString(char)
	}
	return gameSpaces
}

type direction int

const (
	north direction = iota
	east
	south
	west
)

type coordinate struct {
	row int
	col int
}

func (c *coordinate) applyTranslation(d direction) coordinate {
	switch d {
	case north:
		return coordinate{row: c.row - 1, col: c.col}
	case east:
		return coordinate{row: c.row, col: c.col + 1}
	case south:
		return coordinate{row: c.row + 1, col: c.col}
	case west:
		return coordinate{row: c.row, col: c.col - 1}
	default:
		panic(fmt.Sprintf("Invalid Direction: %d", d))
	}
}

type race struct {
	rawMap                   [][]gameSpace
	start                    coordinate
	finish                   coordinate
	pathWithoutCheats        racePath
	cheatlessPathIndexLookup map[coordinate]int
}

func newRace(rawMap [][]gameSpace) race {
	var startCoord coordinate
	startFound := false
	var finishCoord coordinate
	finishFound := false

findStartAndEnd:
	for row := range rawMap {
		for col := range rawMap[row] {
			if rawMap[row][col] == start {
				startCoord = coordinate{row: row, col: col}
				startFound = true
			} else if rawMap[row][col] == finish {
				finishCoord = coordinate{row: row, col: col}
				finishFound = true
			}
			if startFound && finishFound {
				break findStartAndEnd
			}
		}
	}
	if !startFound || !finishFound {
		panic("Didn't find start and/or finish for map!")
	}
	r := race{
		rawMap:                   rawMap,
		start:                    startCoord,
		finish:                   finishCoord,
		cheatlessPathIndexLookup: make(map[coordinate]int),
	}
	r.populatePathWithoutCheats()
	return r
}

func (r *race) inBounds(c coordinate) bool {
	if c.row < 0 || c.col < 0 {
		return false
	}
	if c.row >= len(r.rawMap) || c.col >= len(r.rawMap[c.row]) {
		return false
	}
	return true
}

// populatePathWithoutCheats walks the 'fair' path. This is deterministic for
// the maps given in the problem. This should be done first, as the code to
// figure out cheat paths uses the fair path to branch from.
func (r *race) populatePathWithoutCheats() {
	solution := r.findNoCheatsSolution()
	r.pathWithoutCheats = solution
	for i, c := range solution.path {
		r.cheatlessPathIndexLookup[c] = i
	}
}

func (r *race) findNoCheatsSolution() racePath {
	solutions := make([]racePath, 0)

	frontier := []racePath{
		newRacePath(r, 0),
	}

	for len(frontier) > 0 {
		newFrontier := make([]racePath, 0)
		for _, path := range frontier {
			lc := path.lastCoord()

			if lc == r.finish {
				solutions = append(solutions, path)
				continue
			}

			newFrontier = append(newFrontier, path.nextMoves()...)
		}
		frontier = newFrontier
	}

	if len(solutions) != 1 {
		panic("Expect problem to have 1 unique solution when not cheating!")
	}
	return solutions[0]
}

func (r *race) findAllViableEndCoordinates(s coordinate, clipBudget int) []coordinate {
	endCoordinates := make([]coordinate, 0)
	for rowDelta := -clipBudget; rowDelta <= clipBudget; rowDelta++ {
		if s.row+rowDelta < 0 {
			continue
		}
		if s.row+rowDelta >= len(r.rawMap) {
			// rowDelta only grows, so further colDelta values in this loop will continue to fail this check. Just
			// break to avoid redundant checks.
			break
		}
		absRowDelta := rowDelta
		if absRowDelta < 0 {
			absRowDelta = -absRowDelta
		}
		for colDelta := -clipBudget; colDelta <= clipBudget; colDelta++ {
			if s.col+colDelta < 0 {
				continue
			}
			if s.col+colDelta >= len(r.rawMap[0]) {
				// colDelta only grows, so further colDelta values in this loop will continue to fail this check. Just
				// break to avoid redundant checks.
				break
			}
			absColDelta := colDelta
			if absColDelta < 0 {
				absColDelta = -absColDelta
			}
			if absRowDelta+absColDelta > clipBudget {
				// We're using too much clip budget, try the next values. Note, if we're at a negative value like -9,
				// the next loop iteration will reduce the absColDelta by going to -8.
				continue
			}
			row := s.row + rowDelta
			col := s.col + colDelta
			endCoordinate := coordinate{row: row, col: col}
			if r.rawMap[row][col] != wall {
				endCoordinates = append(endCoordinates, endCoordinate)
			}
		}
	}

	return endCoordinates
}

func (r *race) findClips(cheatStart coordinate, clipBudget int) []clip {
	if clipBudget == 0 {
		return make([]clip, 0)
	}
	clips := make([]clip, 0)

	ends := r.findAllViableEndCoordinates(cheatStart, clipBudget)
	for _, end := range ends {
		if end == cheatStart {
			// Skip ends that move us out to the start of the cheat.
			continue
		}
		if r.cheatlessPathIndexLookup[end] < r.cheatlessPathIndexLookup[cheatStart] {
			// Skip cheats that move us earlier up the path.
			continue
		}

		// Walk up or down all the rows we need to get to the end.
		rowDistance := cheatStart.row - end.row
		if rowDistance < 0 {
			rowDistance *= -1
		}
		colDistance := cheatStart.col - end.col
		if colDistance < 0 {
			colDistance *= -1
		}
		clipDistance := rowDistance + colDistance

		// cheatlessCost is the cost of getting from cheatStart to end on the normal path.
		cheatlessCost := r.cheatlessPathIndexLookup[end] - r.cheatlessPathIndexLookup[cheatStart]
		if clipDistance >= cheatlessCost {
			// Skip cheats that cost more or the same as the cheatless path.
			continue
		}

		clips = append(clips, clip{
			clipStartAndEnd: clipStartAndEnd{
				start: cheatStart,
				end:   end,
			},
			clippedDistance: clipDistance,
			race:            r,
		})
	}

	// Todo, this is likely not needed anymore.
	trimmedClips := lo.UniqBy(clips, func(c clip) clipStartAndEnd {
		return c.clipStartAndEnd
	})
	return trimmedClips
}

func (r *race) findSolutions(clipBudget int) []clip {
	cheatClips := make([]clip, 0)

	for _, c := range r.pathWithoutCheats.path {
		cheatClips = append(cheatClips, r.findClips(c, clipBudget)...)
	}

	return cheatClips
}

func (r *race) print() {
	for row := range r.rawMap {
		for col := range r.rawMap[row] {
			print(r.rawMap[row][col].toString())
		}
		println()
	}
}

type clipStartAndEnd struct {
	start, end coordinate
}

type clip struct {
	race *race
	clipStartAndEnd
	clippedDistance int
}

func (c clip) cost() int {
	cheatlessTotalCost := c.race.pathWithoutCheats.cost()
	costToClipStart := c.race.cheatlessPathIndexLookup[c.start]
	cheatlessCostToClipEnd := c.race.cheatlessPathIndexLookup[c.end]
	// cheatlessClipPathCost is the cost the cheatless path takes to go from c.start -> c.end.
	cheatlessClipPathCost := cheatlessCostToClipEnd - costToClipStart
	// The savings are the difference between the cheatless distance and the cheated.
	savings := cheatlessClipPathCost - c.clippedDistance
	if savings <= 0 {
		panic("This should be unreachable! We don't create cheats that don't save moves!")
	}
	return cheatlessTotalCost - savings
}

type racePath struct {
	race *race
	path []coordinate
}

func newRacePath(r *race, numNoClipsAllowed int) racePath {
	return racePath{
		race: r,
		path: []coordinate{r.start},
	}
}

func (rp *racePath) extend(c coordinate) racePath {
	return racePath{
		race: rp.race,
		path: append(slices.Clone(rp.path), c),
	}
}

func (rp *racePath) lastCoord() coordinate {
	lastCoord, found := lo.Last(rp.path)
	if !found {
		panic("Should always have at least 1 coordinate in a race path!")
	}
	return lastCoord
}

func (rp *racePath) nextMoves() []racePath {
	lc := rp.lastCoord()

	n := lc.applyTranslation(north)
	e := lc.applyTranslation(east)
	s := lc.applyTranslation(south)
	w := lc.applyTranslation(west)
	potentialMoves := []coordinate{n, e, s, w}
	nextMoves := make([]racePath, 0)
	for _, m := range potentialMoves {
		if !rp.race.inBounds(m) {
			continue
		}
		if slices.Contains(rp.path, m) {
			continue
		}
		if rp.race.rawMap[m.row][m.col] == wall {
			// Skip walls.
			continue
		}
		nextMoves = append(nextMoves, rp.extend(m))
	}

	return nextMoves
}

func (rp *racePath) cost() int {
	return len(rp.path) - 1
}

func (rp *racePath) print() {
	mapStringBuffer := make([][]string, len(rp.race.rawMap))
	for row := range rp.race.rawMap {
		mapStringBuffer[row] = make([]string, len(rp.path))
		for col := range rp.race.rawMap[row] {
			mapStringBuffer[row][col] = rp.race.rawMap[row][col].toString()
		}
	}

	for _, c := range rp.path {
		mapStringBuffer[c.row][c.col] = "P"
	}
	for row := range mapStringBuffer {
		for col := range mapStringBuffer[row] {
			fmt.Print(mapStringBuffer[row][col])
		}
		fmt.Println()
	}
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	scanner := bufio.NewScanner(strings.NewReader(input))

	rawMap := make([][]gameSpace, 0)
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}

		rawMap = append(rawMap, handleLine(scanner.Text()))
	}

	r := newRace(rawMap)
	solutionsZeroClips := []racePath{r.pathWithoutCheats}
	slices.SortFunc(solutionsZeroClips, func(a, b racePath) int {
		return a.cost() - b.cost()
	})
	costWithNoCheating := solutionsZeroClips[0].cost()

	solutions2Clip := r.findSolutions(2)
	slices.SortFunc(solutions2Clip, func(a, b clip) int {
		return a.cost() - b.cost()
	})
	count := 0
	for _, cheatSolution := range solutions2Clip {
		//if cheatSolution.cost() == costWithNoCheating-2 {
		if cheatSolution.cost() <= costWithNoCheating-100 {
			count += 1
		}
	}
	fmt.Println(count)

	// Part 2
	solutions20Clip := r.findSolutions(20)
	slices.SortFunc(solutions20Clip, func(a, b clip) int {
		return a.cost() - b.cost()
	})
	count = 0
	for _, cheatSolution := range solutions20Clip {
		//if cheatSolution.cost() == costWithNoCheating-76 {
		if cheatSolution.cost() <= costWithNoCheating-100 {
			count += 1
		}
	}
	fmt.Println(count)
	// 236481 too low
	// 984009 too low
	// 1111431 too high
}
