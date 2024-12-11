package main

import (
	"bufio"
	_ "embed"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/stream"
	"log"
	"slices"
	"strings"
)

//go:embed input
var input string

type coordinate struct {
	row, col int
}

type coordinateWithFacing struct {
	row, col, facing int
}

func (coord coordinateWithFacing) getCoordinate() coordinate {
	return coordinate{coord.row, coord.col}
}

const (
	north = iota
	east
	south
	west
)

func isGuardChar(char string) bool {
	if slices.Contains([]string{"^", ">", "v", "<"}, char) {
		return true
	}
	return false
}

func facingFromChar(facingChar string) int {
	if facingChar == "^" {
		return north
	}
	if facingChar == ">" {
		return east
	}
	if facingChar == "v" {
		return south
	}
	if facingChar == "<" {
		return west
	}
	log.Fatal("invalid facing")
	return -1
}

type gameMap struct {
	floorPlan     [][]string
	guardPosition coordinate
	guardFacing   int

	seenGuardPositions               map[coordinateWithFacing]struct{}
	seenGuardPositionsIgnoringFacing map[coordinate]struct{}
}

func (gm *gameMap) isObstacle(coord coordinate) bool {
	if gm.floorPlan[coord.row][coord.col] == "#" || gm.floorPlan[coord.row][coord.col] == "O" {
		return true
	}
	return false
}

func (gm *gameMap) isOffMap(coord coordinate) bool {
	if coord.row < 0 || coord.col < 0 {
		return true
	}
	if coord.row >= len(gm.floorPlan) || coord.col >= len(gm.floorPlan[0]) {
		return true
	}
	return false
}

func (gm *gameMap) changeGuardFacing() {
	gm.guardFacing += 1
	if gm.guardFacing > west {
		gm.guardFacing = north
	}
}

func (gm *gameMap) walkGuard() (coordinateWithFacing, bool, bool) {
	var nextMoveCandidate coordinateWithFacing
	hasValidNextMove := false
	rotationCount := 0
	for !hasValidNextMove {
		if rotationCount >= 4 {
			log.Fatal("rotation count is too high")
		}

		switch gm.guardFacing {
		case north:
			nextMoveCandidate = coordinateWithFacing{gm.guardPosition.row - 1, gm.guardPosition.col, gm.guardFacing}
		case east:
			nextMoveCandidate = coordinateWithFacing{gm.guardPosition.row, gm.guardPosition.col + 1, gm.guardFacing}
		case south:
			nextMoveCandidate = coordinateWithFacing{gm.guardPosition.row + 1, gm.guardPosition.col, gm.guardFacing}
		case west:
			nextMoveCandidate = coordinateWithFacing{gm.guardPosition.row, gm.guardPosition.col - 1, gm.guardFacing}
		default:
			log.Fatal("invalid guard facing")
		}

		if gm.isOffMap(nextMoveCandidate.getCoordinate()) {
			return nextMoveCandidate, false, true
		}

		if gm.isObstacle(nextMoveCandidate.getCoordinate()) {
			gm.changeGuardFacing()
			rotationCount += 1
		} else {
			hasValidNextMove = true
		}
	}

	gm.guardPosition = nextMoveCandidate.getCoordinate()
	_, seenMoveBefore := gm.seenGuardPositions[nextMoveCandidate]
	if !seenMoveBefore {
		gm.seenGuardPositions[nextMoveCandidate] = struct{}{}
		gm.seenGuardPositionsIgnoringFacing[nextMoveCandidate.getCoordinate()] = struct{}{}
	}
	return nextMoveCandidate, seenMoveBefore, false

}

func (gm *gameMap) printMap() {
	printableMap := make([][]string, len(gm.floorPlan))
	for row := range gm.floorPlan {
		printableMap[row] = make([]string, len(gm.floorPlan[row]))
		for col := range gm.floorPlan[row] {
			printableMap[row][col] = gm.floorPlan[row][col]
		}
	}
	for k := range gm.seenGuardPositions {
		printableMap[k.row][k.col] = "X"
	}

	for row := range printableMap {
		for col := range printableMap[row] {
			print(printableMap[row][col])
		}
		println()
	}
}

func figureOutLoopingObstructions(gm gameMap) []coordinate {
	obstructionsThatCauseLoops := make([]coordinate, 0)

	resultStream := stream.New()
	for row := range gm.floorPlan {
		for col := range gm.floorPlan[row] {
			if gm.floorPlan[row][col] == "#" {
				// Already obstructed.
				continue
			}

			if row == gm.guardPosition.row && col == gm.guardPosition.col {
				// Not allowed to obstruct guard start
				continue
			}

			resultStream.Go(func() stream.Callback {
				copiedFloorPlan := make([][]string, len(gm.floorPlan))
				for i := range gm.floorPlan {
					copiedFloorPlan[i] = make([]string, len(gm.floorPlan[i]))
					copy(copiedFloorPlan[i], gm.floorPlan[i])
				}
				copiedGame := createGameMap(copiedFloorPlan)

				copiedGame.floorPlan[row][col] = "O"

				_, guardLooped, offMap := copiedGame.walkGuard()
				for !guardLooped && !offMap {
					_, guardLooped, offMap = copiedGame.walkGuard()
				}
				if guardLooped {
					return func() {
						obstructionsThatCauseLoops = append(obstructionsThatCauseLoops, coordinate{row, col})
					}
				}
				return func() {}
			})

		}
	}
	resultStream.Wait()

	return obstructionsThatCauseLoops
}

func createGameMap(floorPlan [][]string) gameMap {
	for row := range floorPlan {
		for col := range floorPlan[row] {
			if isGuardChar(floorPlan[row][col]) {
				guardFacing := facingFromChar(floorPlan[row][col])
				seenGuardPositions := make(map[coordinateWithFacing]struct{})
				seenGuardPositions[coordinateWithFacing{row, col, guardFacing}] = struct{}{}
				seenGuardPositionsIgnoringFacing := make(map[coordinate]struct{})
				seenGuardPositionsIgnoringFacing[coordinate{row, col}] = struct{}{}
				return gameMap{
					floorPlan:     floorPlan,
					guardPosition: coordinate{row, col},
					guardFacing:   guardFacing,

					seenGuardPositions:               seenGuardPositions,
					seenGuardPositionsIgnoringFacing: seenGuardPositionsIgnoringFacing,
				}
			}
		}
	}

	log.Fatal("unreachable")
	return gameMap{}
}

func handleLine(line string) []string {
	return lo.ChunkString(line, 1)
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))
	floorPlan := make([][]string, 0)

	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		floorPlan = append(floorPlan, handleLine(scanner.Text()))
	}

	game := createGameMap(floorPlan)

	loopingObstructions := figureOutLoopingObstructions(game)

	for _, _, offMap := game.walkGuard(); offMap == false; _, _, offMap = game.walkGuard() {
		//println(newCoordinate.row, ", ", newCoordinate.col)
	}
	println(len(game.seenGuardPositionsIgnoringFacing))

	print(len(loopingObstructions))

	//game.printMap()

}
