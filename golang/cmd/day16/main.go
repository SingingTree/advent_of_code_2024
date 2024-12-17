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

type direction int

const (
	north direction = iota
	east
	south
	west
)

func (d direction) rotateClockwise() direction {
	switch d {
	case north:
		return east
	case east:
		return south
	case south:
		return west
	case west:
		return north
	}
	panic("invalid direction")
}

func (d direction) rotateCounterclockwise() direction {
	switch d {
	case north:
		return west
	case west:
		return south
	case south:
		return east
	case east:
		return north
	}
	panic("invalid direction")
}

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
	}
	panic("invalid translation and/or coordinate")
}

type gameSpace int

const (
	blank gameSpace = iota
	wall
	start
	end
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
		return end
	}
	panic(fmt.Sprintf("invalid game space %q", s))
}

func (gs *gameSpace) toString() string {
	switch *gs {
	case blank:
		return "."
	case wall:
		return "#"
	case start:
		return "S"
	case end:
		return "E"
	}
	panic(fmt.Sprintf("invalid game space %q", *gs))
}

type move int

const (
	forward move = iota
	clockwiseTurn
	counterclockwiseTurn
	initialPosition
)

type coordinateAndFacing struct {
	coordinate coordinate
	facing     direction
}

type moveWithCoordinateAndFacing struct {
	move move
	coordinateAndFacing
}

type moveChain struct {
	cost     int
	reindeer coordinateAndFacing
	moves    []moveWithCoordinateAndFacing
	//locationsToCost map[coordinateAndFacing]int
}

type game struct {
	rawMap        [][]gameSpace
	startPosition coordinate
	// Start direction is east per instructions.
	endPosition coordinate
	moveChains  []moveChain
}

func newGame(rawMap [][]gameSpace) game {
	var startPosition coordinate
	var endPosition coordinate
	var foundStartPosition bool
	var foundEndPosition bool
	for row := range rawMap {
		for col := range rawMap[row] {
			if rawMap[row][col] == start {
				startPosition = coordinate{row: row, col: col}
				foundStartPosition = true
			}
			if rawMap[row][col] == end {
				endPosition = coordinate{row: row, col: col}
				foundEndPosition = true
			}
			if foundStartPosition && foundEndPosition {
				reindeerCoordinateAndFacing := coordinateAndFacing{
					coordinate: startPosition,
					facing:     east,
				}
				initialMove := moveChain{
					cost:     0,
					reindeer: reindeerCoordinateAndFacing,
					moves: []moveWithCoordinateAndFacing{
						{
							move:                initialPosition,
							coordinateAndFacing: reindeerCoordinateAndFacing,
						},
					},
				}
				return game{
					rawMap:        rawMap,
					startPosition: startPosition,
					endPosition:   endPosition,
					moveChains:    []moveChain{initialMove},
				}
			}
		}
	}
	panic("didn't find reindeer position and/or end position")
}

func (g *game) canReindeerMoveForward(cf coordinateAndFacing) bool {
	moveCandidate := cf.coordinate.applyTranslation(cf.facing)
	if g.rawMap[moveCandidate.row][moveCandidate.col] == wall {
		return false
	}
	return true
}

func (mc *moveChain) move(m move) moveChain {
	switch m {
	case forward:
		newPos := mc.reindeer.coordinate.applyTranslation(mc.reindeer.facing)
		return moveChain{
			cost: mc.cost + 1,
			reindeer: coordinateAndFacing{
				coordinate: newPos,
				facing:     mc.reindeer.facing,
			},
			moves: append(slices.Clone(mc.moves), moveWithCoordinateAndFacing{
				move: forward,
				coordinateAndFacing: coordinateAndFacing{
					newPos,
					mc.reindeer.facing,
				},
			}),
		}
	case clockwiseTurn:
		newFacing := mc.reindeer.facing.rotateClockwise()
		return moveChain{
			cost: mc.cost + 1_000,
			reindeer: coordinateAndFacing{
				coordinate: mc.reindeer.coordinate,
				facing:     newFacing,
			},
			moves: append(slices.Clone(mc.moves), moveWithCoordinateAndFacing{
				move: clockwiseTurn,
				coordinateAndFacing: coordinateAndFacing{
					mc.reindeer.coordinate,
					newFacing,
				},
			}),
		}
	case counterclockwiseTurn:
		newFacing := mc.reindeer.facing.rotateCounterclockwise()
		return moveChain{
			cost: mc.cost + 1_000,
			reindeer: coordinateAndFacing{
				coordinate: mc.reindeer.coordinate,
				facing:     newFacing,
			},
			moves: append(slices.Clone(mc.moves), moveWithCoordinateAndFacing{
				move: counterclockwiseTurn,
				coordinateAndFacing: coordinateAndFacing{
					mc.reindeer.coordinate,
					newFacing,
				},
			}),
		}
	case initialPosition:
		panic("Should never try and execute an initial position move!")
	}
	panic(fmt.Sprintf("invalid move: %d", m))
}

func (g *game) findSolutions() []moveChain {
	coordinateToLowestCost := make(map[coordinateAndFacing]int)

	winningChains := make([]moveChain, 0)

	frontier := g.moveChains
	for len(frontier) > 0 {
		newFrontier := make([]moveChain, 0)
		for _, mc := range frontier {
			lastMove, ok := lo.Last(mc.moves)
			if !ok {
				panic("A movechain should always have moves")
			}
			cost, ok := coordinateToLowestCost[lastMove.coordinateAndFacing]
			if ok && cost < mc.cost {
				// We're already reached this position with a lower cost. Cull this chain.
				continue
			}

			coordinateToLowestCost[lastMove.coordinateAndFacing] = mc.cost

			if lastMove.coordinate == g.endPosition {
				winningChains = append(winningChains, mc)
				continue
			}

			if g.canReindeerMoveForward(mc.reindeer) {
				newFrontier = append(newFrontier, mc.move(forward))
			}
			newFrontier = append(newFrontier, mc.move(clockwiseTurn), mc.move(counterclockwiseTurn))
		}
		frontier = newFrontier
	}

	slices.SortFunc(winningChains, func(a moveChain, b moveChain) int {
		return a.cost - b.cost
	})
	winningCost := winningChains[0].cost
	winningChainsWithMinCost := make([]moveChain, 0)
	for _, mc := range winningChains {
		if mc.cost > winningCost {
			break
		}
		winningChainsWithMinCost = append(winningChainsWithMinCost, mc)
	}

	return winningChainsWithMinCost
}

func (g *game) printMap() {
	for row := range g.rawMap {
		for col := range g.rawMap[row] {
			print(g.rawMap[row][col].toString())
		}
		println()
	}
}

func (g *game) printMapWithMoveChain(mc *moveChain) {
	stringMap := make([][]string, len(g.rawMap))
	for row := range g.rawMap {
		stringMap[row] = make([]string, len(g.rawMap[row]))
		for col := range g.rawMap[row] {
			stringMap[row][col] = g.rawMap[row][col].toString()
		}
	}

	for _, m := range mc.moves {
		stringMap[m.coordinate.row][m.coordinate.col] = "M"
	}

	for row := range stringMap {
		for col := range stringMap[row] {
			print(stringMap[row][col])
		}
		println()
	}
}

func handleLine(line string) []gameSpace {
	chars := lo.ChunkString(line, 1)
	gameLine := make([]gameSpace, len(chars))
	for i, char := range chars {
		gameLine[i] = gameSpaceFromString(char)
	}
	return gameLine
}

func findWinningTileCount(winningChains []moveChain) int {
	tiles := make(map[coordinate]struct{})
	for _, mc := range winningChains {
		for _, m := range mc.moves {
			tiles[m.coordinate] = struct{}{}
		}
	}

	return len(tiles)
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

	// Part 1
	g := newGame(rawMap)
	solutions := g.findSolutions()
	println(solutions[0].cost)

	// Part 2
	println(findWinningTileCount(solutions))
}
