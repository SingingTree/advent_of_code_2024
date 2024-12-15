package main

import (
	"bufio"
	_ "embed"
	"github.com/samber/lo"
	"log"
	"slices"
	"strings"
)

//go:embed input
var input string

type robotMove int

const (
	north robotMove = iota
	east
	south
	west
)

func robotMoveFromString(s string) robotMove {
	switch s {
	case "^":
		return north
	case ">":
		return east
	case "v":
		return south
	case "<":
		return west
	default:
		log.Fatalf("invalid robot move: %s", s)
	}
	return north
}

type gameSpace int

const (
	blank gameSpace = iota
	robot
	box
	leftSideOfBox
	rightSideOfBox
	wall
)

func gameSpaceFromString(s string) gameSpace {
	// We don't parse wide boxes here, as we generate wide maps rather than reading them.
	switch s {
	case ".":
		return blank
	case "@":
		return robot
	case "O":
		return box
	case "#":
		return wall
	default:
		log.Fatalf("invalid game space: %q", s)
	}
	return blank
}

func (g gameSpace) String() string {
	switch g {
	case blank:
		return "."
	case robot:
		return "@"
	case box:
		return "O"
	case leftSideOfBox:
		return "["
	case rightSideOfBox:
		return "]"
	case wall:
		return "#"
	default:
		log.Fatalf("invalid game space: %d", g)
	}
	return "."
}

type coordinate struct {
	row int
	col int
}

func (c *coordinate) applyTranslation(move robotMove) coordinate {
	switch move {
	case north:
		return coordinate{row: c.row - 1, col: c.col}
	case east:
		return coordinate{row: c.row, col: c.col + 1}
	case south:
		return coordinate{row: c.row + 1, col: c.col}
	case west:
		return coordinate{row: c.row, col: c.col - 1}
	default:
		log.Fatalf("invalid game move: %d", move)
	}
	return coordinate{row: c.row, col: c.col}
}

type gameMap struct {
	rawMap        [][]gameSpace
	robotLocation coordinate
	robotMoves    []robotMove
	nextMoveIndex int
}

func newGameMap(rawMap [][]gameSpace, robotMoves []robotMove) gameMap {
	for row := range rawMap {
		for col := range rawMap[row] {
			if rawMap[row][col] == robot {
				return gameMap{
					rawMap:        rawMap,
					robotLocation: coordinate{row, col},
					robotMoves:    robotMoves,
					nextMoveIndex: 0,
				}
			}
		}
	}
	log.Fatal("Couldn't find robot")
	return gameMap{}
}

// iterate was written for part 1. iterateMachTwo supersedes this method, but
// iterate is kept around so we can use it to compare outcomes.
func (gm *gameMap) iterate() {
	if gm.nextMoveIndex >= len(gm.robotMoves) {
		log.Fatal("Don't call iterate() when no moves are left!")
	}
	nextMove := gm.robotMoves[gm.nextMoveIndex]

	// No matter what we increment the move.
	gm.nextMoveIndex += 1

	candidateLocation := gm.robotLocation.applyTranslation(nextMove)
	if gm.rawMap[candidateLocation.row][candidateLocation.col] == wall {
		// Robot bumps into a wall, nothing happens.
		return
	}

	if gm.rawMap[candidateLocation.row][candidateLocation.col] == blank {
		// Robot moves into a blank space.
		gm.rawMap[candidateLocation.row][candidateLocation.col] = robot
		gm.rawMap[gm.robotLocation.row][gm.robotLocation.col] = blank
		gm.robotLocation = candidateLocation
		return
	}

	if gm.rawMap[candidateLocation.row][candidateLocation.col] != box {
		log.Fatal("Unexpected game state (logic bug?), by this point robot must be pushing box!")
	}

	// We're pushing a box.
	boxCount := 1
	wallInWay := false
	nextCoordiates := candidateLocation
	for {
		// Check if we're pushing more than 1 box, or hitting a wall.
		nextCoordiates = nextCoordiates.applyTranslation(nextMove)
		if gm.rawMap[nextCoordiates.row][nextCoordiates.col] == wall {
			wallInWay = true
			break
		}
		if gm.rawMap[nextCoordiates.row][nextCoordiates.col] == blank {
			break
		}
		if gm.rawMap[nextCoordiates.row][nextCoordiates.col] != box {
			log.Fatal("Unexpected game state (logic bug?), expected box.")
		}
		boxCount += 1
	}

	if wallInWay {
		// Can't push, because we're blocked by a wall!
		return
	}

	// Move robot.
	gm.rawMap[candidateLocation.row][candidateLocation.col] = robot
	gm.rawMap[gm.robotLocation.row][gm.robotLocation.col] = blank
	gm.robotLocation = candidateLocation
	// Move all boxes.
	boxLocation := candidateLocation.applyTranslation(nextMove)
	for range boxCount {
		gm.rawMap[boxLocation.row][boxLocation.col] = box
		boxLocation = boxLocation.applyTranslation(nextMove)
	}
}

// Iterate version for part 2 that handles push groups.
func (gm *gameMap) iterateMachTwo() {
	if gm.nextMoveIndex >= len(gm.robotMoves) {
		log.Fatal("Don't call iterate() when no moves are left!")
	}
	nextMove := gm.robotMoves[gm.nextMoveIndex]

	// No matter what we increment the move.
	gm.nextMoveIndex += 1

	pg := newPushGroup(nextMove, gm)
	if !pg.canPush() {
		return
	}
	pg.push()
}

func (gm *gameMap) doAllMoves() {
	for gm.nextMoveIndex < len(gm.robotMoves) {
		gm.iterate()
	}
}

func (gm *gameMap) doAllMovesMachTwo() {
	for gm.nextMoveIndex < len(gm.robotMoves) {
		gm.iterateMachTwo()
	}
}

func (gm *gameMap) gpsScore() int {
	score := 0
	for row := range gm.rawMap {
		for col := range gm.rawMap[row] {
			if gm.rawMap[row][col] == box {
				score += 100*row + col
			} else if gm.rawMap[row][col] == leftSideOfBox {
				score += 100*row + col
			}
		}
	}
	return score
}

func (gm *gameMap) makeWideMap() gameMap {
	wideRawMap := make([][]gameSpace, len(gm.rawMap))
	for row := range gm.rawMap {
		wideRawMap[row] = make([]gameSpace, len(gm.rawMap)*2)
		for col, gs := range gm.rawMap[row] {
			wideCol := col * 2
			switch gs {
			case blank:
				wideRawMap[row][wideCol] = blank
				wideRawMap[row][wideCol+1] = blank
			case robot:
				wideRawMap[row][wideCol] = robot
				wideRawMap[row][wideCol+1] = blank
			case box:
				wideRawMap[row][wideCol] = leftSideOfBox
				wideRawMap[row][wideCol+1] = rightSideOfBox
			case wall:
				wideRawMap[row][wideCol] = wall
				wideRawMap[row][wideCol+1] = wall
			default:
				log.Fatal("Unexpected game state (logic bug?). Are you trying to widen an already wide map?")
			}
		}
	}
	for row := range wideRawMap {
		for col := range wideRawMap[row] {
			if wideRawMap[row][col] == robot {
				return gameMap{
					rawMap:        wideRawMap,
					robotLocation: coordinate{row, col},
					robotMoves:    slices.Clone(gm.robotMoves),
					nextMoveIndex: gm.nextMoveIndex,
				}
			}
		}
	}
	log.Fatal("Failed to find robot on wide map!")
	return gameMap{}
}

func (gm *gameMap) print() {
	for row := range gm.rawMap {
		for col := range gm.rawMap[row] {
			print(gm.rawMap[row][col].String())
		}
		println()
	}
}

type pushGroup struct {
	direction        robotMove
	startCoordinates []coordinate
	gm               *gameMap
}

func newPushGroup(
	direction robotMove,
	gm *gameMap,
) pushGroup {
	startCoordinates := make([]coordinate, 0)

	frontier := []coordinate{gm.robotLocation}
	seenCoordinates := make(map[coordinate]struct{})
	for len(frontier) > 0 {
		nextFrontier := make([]coordinate, 0)
		for _, c := range frontier {
			if _, seen := seenCoordinates[c]; seen {
				continue
			}
			startCoordinates = append(startCoordinates, c)
			seenCoordinates[c] = struct{}{}
			nextCoordinate := c.applyTranslation(direction)
			if gm.rawMap[nextCoordinate.row][nextCoordinate.col] == box {
				nextFrontier = append(nextFrontier, nextCoordinate)
			} else if gm.rawMap[nextCoordinate.row][nextCoordinate.col] == leftSideOfBox {
				nextFrontier = append(nextFrontier, nextCoordinate)
				nextFrontier = append(nextFrontier, nextCoordinate.applyTranslation(east))
			} else if gm.rawMap[nextCoordinate.row][nextCoordinate.col] == rightSideOfBox {
				nextFrontier = append(nextFrontier, nextCoordinate)
				nextFrontier = append(nextFrontier, nextCoordinate.applyTranslation(west))
			}
		}
		frontier = nextFrontier
	}

	return pushGroup{
		direction:        direction,
		startCoordinates: startCoordinates,
		gm:               gm,
	}
}

func (pg *pushGroup) canPush() bool {
	pushedCoordinates := make([]coordinate, len(pg.startCoordinates))
	for i, c := range pg.startCoordinates {
		nextCoordinate := c.applyTranslation(pg.direction)
		pushedCoordinates[i] = nextCoordinate
	}
	for _, c := range pushedCoordinates {
		if pg.gm.rawMap[c.row][c.col] == wall {
			return false
		}
	}
	return true
}

func (pg *pushGroup) push() {
	// Figure out what coordinates will look like following the push.
	coordinateToNewValue := make(map[coordinate]gameSpace)
	for _, c := range pg.startCoordinates {
		nextCoordinate := c.applyTranslation(pg.direction)
		coordinateToNewValue[nextCoordinate] = pg.gm.rawMap[c.row][c.col]
	}
	// Do the push.
	var robotLocation coordinate
	for c, gs := range coordinateToNewValue {
		pg.gm.rawMap[c.row][c.col] = gs
		if gs == robot {
			robotLocation = c
		}
	}
	pg.gm.robotLocation = robotLocation
	// Tidy up coordinates following the push.
	for _, c := range pg.startCoordinates {
		_, wasPushedInto := coordinateToNewValue[c]
		if !wasPushedInto {
			// If the coordinate wasn't pushed into, it becomes blank.
			pg.gm.rawMap[c.row][c.col] = blank
		}
	}
}

func handleMapLine(line string) []gameSpace {
	chars := lo.ChunkString(line, 1)
	mapLine := make([]gameSpace, len(chars))
	for i, char := range chars {
		mapLine[i] = gameSpaceFromString(char)
	}
	return mapLine
}

func handleRobotMoveLine(line string) []robotMove {
	chars := lo.ChunkString(line, 1)
	moves := make([]robotMove, len(chars))
	for i, char := range chars {
		moves[i] = robotMoveFromString(char)
	}
	return moves
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	handlingMap := true
	rawMap := make([][]gameSpace, 0)
	robotMoves := make([]robotMove, 0)
	for scanner.Scan() {
		if scanner.Text() == "" {
			handlingMap = false
			continue
		}
		if handlingMap {
			rawMap = append(rawMap, handleMapLine(scanner.Text()))
		} else {
			robotMoves = append(robotMoves, handleRobotMoveLine(scanner.Text())...)
		}
	}

	gm := newGameMap(rawMap, robotMoves)

	// Get the wide map ready before we start messing with the map.
	wideGm := gm.makeWideMap()

	// Part1
	gm.doAllMovesMachTwo()
	println(gm.gpsScore())

	// Part2
	wideGm.doAllMovesMachTwo()
	println(wideGm.gpsScore())
}
