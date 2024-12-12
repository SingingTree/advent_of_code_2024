package main

import (
	"bufio"
	_ "embed"
	"github.com/samber/lo"
	"slices"

	"strings"
)

//go:embed input
var input string

func handleLine(line string) []string {
	return lo.ChunkString(line, 1)
}

type gameMap struct {
	rawMap  [][]string
	regions []region
}

func newGameMap(rawMap [][]string) gameMap {
	gm := gameMap{rawMap: rawMap}
	gm.findRegions()
	return gm
}

type coordinate struct {
	row int
	col int
}

type fenceCoordinate struct {
	coordinate
	northFence bool
	eastFence  bool
	southFence bool
	westFence  bool
}

type region struct {
	label       string
	coordinates []coordinate
}

func (r *region) inBounds(coord coordinate) bool {
	if slices.Contains(r.coordinates, coord) {
		return true
	}
	return false
}

func (r *region) fenceNeeded() int {
	fenceCount := 0
	for _, c := range r.coordinates {
		north := coordinate{row: c.row - 1, col: c.col}
		east := coordinate{row: c.row, col: c.col + 1}
		south := coordinate{row: c.row + 1, col: c.col}
		west := coordinate{row: c.row, col: c.col - 1}

		// Need a fence on all sides that are out of the region.
		if !r.inBounds(north) {
			fenceCount += 1
		}
		if !r.inBounds(east) {
			fenceCount += 1
		}
		if !r.inBounds(south) {
			fenceCount += 1
		}
		if !r.inBounds(west) {
			fenceCount += 1
		}
	}
	return fenceCount
}

func (r *region) fenceCost() int {
	//println(r.label, ": ", r.fenceNeeded(), " * ", len(r.coordinates))
	return r.fenceNeeded() * len(r.coordinates)
}

func (r *region) fenceSides() int {
	fenceLookup := make(map[coordinate]fenceCoordinate)
	// Build a map of fences.
	for _, c := range r.coordinates {
		north := coordinate{row: c.row - 1, col: c.col}
		east := coordinate{row: c.row, col: c.col + 1}
		south := coordinate{row: c.row + 1, col: c.col}
		west := coordinate{row: c.row, col: c.col - 1}

		var northFence, eastFence, southFence, westFence bool
		if !r.inBounds(north) {
			northFence = true
		}
		if !r.inBounds(east) {
			eastFence = true
		}
		if !r.inBounds(south) {
			southFence = true
		}
		if !r.inBounds(west) {
			westFence = true
		}

		fenceLookup[c] = fenceCoordinate{
			coordinate: c,
			northFence: northFence,
			eastFence:  eastFence,
			southFence: southFence,
			westFence:  westFence,
		}
	}

	smallestRowCoord := lo.MinBy(r.coordinates, func(coord coordinate, minRow coordinate) bool {
		return coord.row < minRow.row
	})
	smallestColCoord := lo.MinBy(r.coordinates, func(coord coordinate, minRow coordinate) bool {
		return coord.col < minRow.col
	})
	smallestRow := smallestRowCoord.row
	smallestCol := smallestColCoord.col

	largestRowCoord := lo.MaxBy(r.coordinates, func(coord coordinate, maxRow coordinate) bool {
		return coord.row > maxRow.row
	})
	largestColCoord := lo.MaxBy(r.coordinates, func(coord coordinate, maxCol coordinate) bool {
		return coord.col > maxCol.col
	})
	largestRow := largestRowCoord.row
	largestCol := largestColCoord.col

	sides := 0
	var inNorthFence, inEastFence, inSouthFence, inWestFence bool
	// Sweep along every row.
	for i := smallestRow; i <= largestRow; i++ {
		for j := smallestCol; j <= largestCol; j++ {
			if !r.inBounds(coordinate{row: i, col: j}) {
				// We only check north and south fences during row slides.
				if inNorthFence {
					sides += 1
					inNorthFence = false
				}
				if inSouthFence {
					sides += 1
					inSouthFence = false
				}
				// Check next coordinate
				continue
			}
			// This coordinate is in the region.
			fenceC := fenceLookup[coordinate{row: i, col: j}]
			if fenceC.northFence {
				inNorthFence = true
			} else if inNorthFence {
				// Ending a north fence.
				sides += 1
				inNorthFence = false
			}

			if fenceC.southFence {
				inSouthFence = true
			} else if inSouthFence {
				// Ending a south fence.
				sides += 1
				inSouthFence = false
			}
		}
		// Fences end at the end of the row if not already ended.
		if inNorthFence {
			sides += 1
			inNorthFence = false
		}
		if inSouthFence {
			sides += 1
			inSouthFence = false
		}
	}

	// Sweep along every col.
	for i := smallestCol; i <= largestCol; i++ {
		for j := smallestRow; j <= largestRow; j++ {
			if !r.inBounds(coordinate{row: j, col: i}) {
				// We only check north and south fences during row slides.
				if inEastFence {
					sides += 1
					inEastFence = false
				}
				if inWestFence {
					sides += 1
					inWestFence = false
				}
				// Check next coordinate
				continue
			}
			// This coordinate is in the region.
			fenceC := fenceLookup[coordinate{row: j, col: i}]
			if fenceC.eastFence {
				inEastFence = true
			} else if inEastFence {
				// Ending an east fence.
				sides += 1
				inEastFence = false
			}

			if fenceC.westFence {
				inWestFence = true
			} else if inWestFence {
				// Ending a south fence.
				sides += 1
				inWestFence = false
			}
		}
		// Fences end at the end of the col if not already ended.
		if inEastFence {
			sides += 1
			inEastFence = false
		}
		if inWestFence {
			sides += 1
			inWestFence = false
		}
	}
	return sides
}

func (r *region) discountFenceCost() int {
	//println(r.label, ": ", r.fenceSides(), " * ", len(r.coordinates), " = ", r.fenceSides()*len(r.coordinates))
	return r.fenceSides() * len(r.coordinates)
}

func (gm *gameMap) getRegion(coord coordinate) region {
	regionLabel := gm.rawMap[coord.row][coord.col]

	regionCoordinates := make([]coordinate, 0)
	frontier := []coordinate{coord}
	exploredCoordinates := make(map[coordinate]struct{})
	for len(frontier) > 0 {
		newFrontier := make([]coordinate, 0)
		for _, c := range frontier {
			if _, alreadySeen := exploredCoordinates[c]; alreadySeen {
				continue
			}
			exploredCoordinates[c] = struct{}{}
			regionCoordinates = append(regionCoordinates, c)

			north := coordinate{c.row - 1, c.col}
			east := coordinate{c.row, c.col + 1}
			south := coordinate{c.row + 1, c.col}
			west := coordinate{c.row, c.col - 1}

			_, northSeen := exploredCoordinates[north]
			if !northSeen && gm.isBounds(north) && gm.rawMap[north.row][north.col] == regionLabel {
				newFrontier = append(newFrontier, north)
			}
			_, eastSeen := exploredCoordinates[east]
			if !eastSeen && gm.isBounds(east) && gm.rawMap[east.row][east.col] == regionLabel {
				newFrontier = append(newFrontier, east)
			}
			_, southSeen := exploredCoordinates[south]
			if !southSeen && gm.isBounds(south) && gm.rawMap[south.row][south.col] == regionLabel {
				newFrontier = append(newFrontier, south)
			}
			_, westSeen := exploredCoordinates[west]
			if !westSeen && gm.isBounds(west) && gm.rawMap[west.row][west.col] == regionLabel {
				newFrontier = append(newFrontier, west)
			}
		}
		frontier = newFrontier
	}
	return region{
		label:       regionLabel,
		coordinates: regionCoordinates,
	}
}

func (gm *gameMap) findRegions() {
	exploredCoordinates := make(map[coordinate]struct{})
	gm.regions = make([]region, 0)

	for row := range gm.rawMap {
		for col := range gm.rawMap[row] {
			_, ok := exploredCoordinates[coordinate{row, col}]
			if ok {
				//println("Skipping ", row, ", ", col)
				// Don't explore regions we already know about.
				continue
			}
			reg := gm.getRegion(coordinate{row, col})
			for _, c := range reg.coordinates {
				exploredCoordinates[c] = struct{}{}
			}
			gm.regions = append(gm.regions, reg)
		}
	}
}

func (gm *gameMap) isBounds(coord coordinate) bool {
	if coord.row < 0 || coord.col < 0 {
		return false
	}
	if coord.row >= len(gm.rawMap) || coord.col >= len(gm.rawMap[coord.row]) {
		return false
	}
	return true
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	rawMap := make([][]string, 0)
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		row := handleLine(scanner.Text())
		rawMap = append(rawMap, row)
	}

	gm := newGameMap(rawMap)

	fenceCost := 0
	for _, reg := range gm.regions {
		fenceCost += reg.fenceCost()
	}
	println(fenceCost)

	discountFenceCost := 0
	for _, reg := range gm.regions {
		discountFenceCost += reg.discountFenceCost()
	}
	println(discountFenceCost)
}
