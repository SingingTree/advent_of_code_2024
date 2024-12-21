package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"github.com/samber/lo"
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

func (r *race) findClips(cheatStart coordinate, clipBudget int) []clip {
	if clipBudget == 0 {
		return make([]clip, 0)
	}
	tileCost := make(map[coordinate]int)
	clips := make([]clip, 0)

	n := []coordinate{cheatStart.applyTranslation(north)}
	e := []coordinate{cheatStart.applyTranslation(east)}
	s := []coordinate{cheatStart.applyTranslation(south)}
	w := []coordinate{cheatStart.applyTranslation(west)}
	frontier := [][]coordinate{n, e, s, w}
	// Filter initial frontier to contain wall tiles.
	frontier = lo.Filter(frontier, func(c []coordinate, _ int) bool {
		return r.inBounds(c[0]) && r.rawMap[c[0].row][c[0].col] == wall
	})
	costSoFar := 0
	for len(frontier) > 0 && clipBudget-costSoFar > 0 {
		costSoFar += 1
		newFrontier := make([][]coordinate, 0)
		for _, cc := range frontier {
			lastCoord, found := lo.Last(cc)
			if !found {
				panic("Should always have a last coord!")
			}
			n2 := lastCoord.applyTranslation(north)
			e2 := lastCoord.applyTranslation(east)
			s2 := lastCoord.applyTranslation(south)
			w2 := lastCoord.applyTranslation(west)
			nextCoords := []coordinate{n2, e2, s2, w2}

			for _, c := range nextCoords {
				if !r.inBounds(c) {
					continue
				}
				if c == cheatStart {
					// Skip cheats that walk back to the cheatStart, they're trivially
					// non-optimal.
					continue
				}
				// Check if we've already seen this tile.
				cachedCost, found := tileCost[c]
				if found {
					if cachedCost > costSoFar {
						panic("This should be unreachable, because we're doing BFS")
					}
					// We've already found cheaper cheats.
					continue
				}
				tileCost[c] = costSoFar
				// Seen tile check done.

				if r.rawMap[c.row][c.col] != wall {
					if r.cheatlessPathIndexLookup[cheatStart] > r.cheatlessPathIndexLookup[c] {
						// Skip cheats that increase our cost.
						continue
					}
					clips = append(clips, clip{
						clipStartAndEnd: clipStartAndEnd{cheatStart, c},
						middle:          slices.Clone(cc),
					})
					continue
				}
				// Handle non-end cases.
				newFrontier = append(newFrontier, append(slices.Clone(cc), c))
			}
		}
		frontier = newFrontier
	}

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

func (rp *racePath) isClipping() bool {
	lc := rp.lastCoord()
	if rp.race.rawMap[lc.row][lc.col] == wall {
		return true
	}
	return false
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

func main() {
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
		if cheatSolution.cost() <= costWithNoCheating-100 {
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
		if cheatSolution.cost() == costWithNoCheating-52 {
			//fmt.Printf("%v, %v\n", cheatSolution.clips[0].start, cheatSolution.clips[0].end)
			count += 1
		}
	}
	println(count)
}
