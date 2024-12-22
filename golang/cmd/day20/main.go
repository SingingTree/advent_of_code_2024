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
				break
			}
			absColDelta := colDelta
			if absColDelta < 0 {
				absColDelta = -absColDelta
			}
			if absRowDelta+absColDelta > clipBudget {
				break
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

	n := cheatStart.applyTranslation(north)
	e := cheatStart.applyTranslation(east)
	s := cheatStart.applyTranslation(south)
	w := cheatStart.applyTranslation(west)
	frontier := []coordinate{n, e, s, w}
	// Filter initial frontier to contain wall tiles.
	frontier = lo.Filter(frontier, func(c coordinate, _ int) bool {
		return r.inBounds(c) && r.rawMap[c.row][c.col] == wall
	})

	for _, wallCoordinate := range frontier {
		// We don't need to reduce the clip budget for find end coords. We've used 1 moving into the wall, but
		// the end coordinate code users 1 cost to exit the wall, which 'gives' us that 1 back as it were.
		ends := r.findAllViableEndCoordinates(wallCoordinate, clipBudget)
		for _, end := range ends {
			if end == cheatStart {
				// Skip ends that move us out to the start of the cheat.
				continue
			}
			if r.cheatlessPathIndexLookup[end] < r.cheatlessPathIndexLookup[cheatStart] {
				// Skip cheats that move us earlier up the path.
				continue
			}

			// Add all the expected coordinates to middle path (including our initial wall and skipping the end).
			middlePath := make([]coordinate, 0)
			currentRow := wallCoordinate.row
			// Walk up or down all the rows we need to get to the end.
			for currentRow != end.row {
				middlePath = append(middlePath, coordinate{row: currentRow, col: wallCoordinate.col})
				if currentRow < end.row {
					currentRow += 1
				} else {
					currentRow -= 1
				}
			}
			if currentRow != end.row {
				panic("This should be unreachable!")
			}
			currentCol := wallCoordinate.col
			for currentCol != end.col {
				middlePath = append(middlePath, coordinate{row: currentRow, col: currentCol})
				if currentCol < end.col {
					currentCol += 1
				} else {
					currentCol -= 1
				}
			}

			cheatlessCost := r.cheatlessPathIndexLookup[end] - r.cheatlessPathIndexLookup[cheatStart]
			if len(middlePath)+1 >= cheatlessCost {
				// Skip cheats that cost more or the same as the cheatless path.
				continue
			}

			clips = append(clips, clip{
				clipStartAndEnd: clipStartAndEnd{
					start: cheatStart,
					end:   end,
				},
				middle: middlePath,
			})
		}
	}

	// Todo, this is likely not needed anymore.
	trimmedClips := lo.UniqBy(clips, func(c clip) clipStartAndEnd {
		return c.clipStartAndEnd
	})
	return trimmedClips
}

func (r *race) findSolutions(numClips int, perClipWallBudget int) []racePath {
	if numClips == 0 {
		return []racePath{
			r.pathWithoutCheats,
		}
	}
	if numClips != 1 {
		panic("Not yet setup to handle more than 1 cheat!")
	}

	cheatClips := make([]clip, 0)

	for _, c := range r.pathWithoutCheats.path {
		cheatClips = append(cheatClips, r.findClips(c, perClipWallBudget)...)
	}

	cheatPaths := make([]racePath, len(cheatClips))
	for i, cc := range cheatClips {
		// Clone the path without cheats up to where the cheat starts (non including start).
		path := slices.Clone(r.pathWithoutCheats.path[:r.cheatlessPathIndexLookup[cc.start]])
		// Add the cheat start and middle to the path.
		path = append(path, cc.start)
		path = append(path, cc.middle...)
		// Add the rest of the path after the cheat has finished.
		path = append(path, r.pathWithoutCheats.path[r.cheatlessPathIndexLookup[cc.end]:]...)

		cheatPaths[i] = racePath{
			race:  r,
			path:  path,
			clips: []clip{cc},
		}
	}

	return cheatPaths
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
	clipStartAndEnd
	middle []coordinate
}

type racePath struct {
	race  *race
	path  []coordinate
	clips []clip
}

func newRacePath(r *race, numNoClipsAllowed int) racePath {
	return racePath{
		race:  r,
		path:  []coordinate{r.start},
		clips: make([]clip, 0),
	}
}

func (rp *racePath) extend(c coordinate) racePath {
	return racePath{
		race:  rp.race,
		path:  append(slices.Clone(rp.path), c),
		clips: slices.Clone(rp.clips),
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
	for _, cl := range rp.clips {
		for _, c := range cl.middle {
			mapStringBuffer[c.row][c.col] = "C"
		}
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
	solutionsZeroClips := r.findSolutions(0, 0)
	slices.SortFunc(solutionsZeroClips, func(a, b racePath) int {
		return a.cost() - b.cost()
	})
	costWithNoCheating := solutionsZeroClips[0].cost()

	solutionsOneClip := r.findSolutions(1, 1)
	slices.SortFunc(solutionsOneClip, func(a, b racePath) int {
		return a.cost() - b.cost()
	})
	count := 0
	for _, cheatSolution := range solutionsOneClip {
		if cheatSolution.cost() == costWithNoCheating-2 {
			count += 1
		}
	}
	println(count)

	solutions20Clip := r.findSolutions(1, 20)
	slices.SortFunc(solutionsOneClip, func(a, b racePath) int {
		return a.cost() - b.cost()
	})
	count = 0
	for _, cheatSolution := range solutions20Clip {
		if cheatSolution.cost() == costWithNoCheating-72 {
			//fmt.Printf("%v\n", cheatSolution.path)
			//cheatSolution.print()
			cheatSolution.validate()
			count += 1
		}
	}
	println(count) // 236481 too low
}

func (rp *racePath) validate() {
	for i := 0; i < len(rp.path)-1; i++ {
		rowDistance := rp.path[i].row - rp.path[i+1].row
		if rowDistance < 0 {
			rowDistance *= 1
		}
		colDistance := rp.path[i].col - rp.path[i+1].col
		if colDistance < 0 {
			colDistance *= 1
		}
		if rowDistance+colDistance > 1 {
			panic("Discontinuity")
		}
	}
	for _, c := range rp.path {
		if !rp.race.inBounds(c) {
			panic("Out of bounds")
		}
	}
	if rp.path[0] != rp.race.start {
		panic("Incorrect start")
	}
	if rp.path[len(rp.path)-1] != rp.race.finish {
		panic("Incorrect finish")
	}
	if len(rp.clips) >= 1 {
		if len(rp.clips[0].middle) > 20 {
			panic("Clipped too long")
		}
	}
}
