package main

import (
	"bufio"
	_ "embed"
	"log"
	"regexp"
	"strconv"
	"strings"
)

//go:embed input
var input string

// p=4,11 v=-61,-65
var lineRegex = regexp.MustCompile(`^p=(-*\d+),(-*\d+) v=(-*\d+),(-*\d+)$`)

type coordinate struct {
	x int
	y int
}

type robot struct {
	position coordinate
	velocity coordinate
}

// maxX and maxY are non inclusive.
func (r *robot) move(maxX int, maxY int) robot {
	newX := r.position.x + r.velocity.x
	newY := r.position.y + r.velocity.y
	if newX < 0 {
		newX = maxX + newX
	}
	if newY < 0 {
		newY = maxY + newY
	}
	if newX >= maxX {
		newX = newX - maxX
	}
	if newY >= maxY {
		newY = newY - maxY
	}
	if newX < 0 || newY < 0 {
		log.Fatal("New coordinate less than 0")
	}
	if newX > maxX || newY > maxY {
		log.Fatal("New coordinate greater than max coordinate")
	}
	return robot{
		position: coordinate{newX, newY},
		velocity: coordinate{r.velocity.x, r.velocity.y},
	}
}

type gameMap struct {
	robotMap    [][][]*robot
	boardWidth  int
	boardHeight int
}

func (gm *gameMap) iterate() gameMap {
	newRobotMap := make([][][]*robot, len(gm.robotMap))
	for row := range newRobotMap {
		newRobotMap[row] = make([][]*robot, len(gm.robotMap[row]))
	}
	for row := range gm.robotMap {
		for col := range gm.robotMap[row] {
			for _, r := range gm.robotMap[row][col] {
				newR := r.move(gm.boardWidth, gm.boardHeight)
				newRobotMap[newR.position.y][newR.position.x] = append(
					newRobotMap[newR.position.y][newR.position.x],
					&newR,
				)
			}
		}
	}
	return gameMap{
		robotMap:    newRobotMap,
		boardWidth:  gm.boardWidth,
		boardHeight: gm.boardHeight,
	}
}

func newGameMap(robots []robot, maxWidth int, maxHeight int) gameMap {
	robotMap := make([][][]*robot, maxHeight)
	for row := range robotMap {
		robotMap[row] = make([][]*robot, maxWidth)
	}
	for _, r := range robots {
		robotMap[r.position.y][r.position.x] = append(robotMap[r.position.y][r.position.x], &r)
	}
	gm := gameMap{
		robotMap:    robotMap,
		boardWidth:  maxWidth,
		boardHeight: maxHeight,
	}
	return gm
}

func (gm *gameMap) safetyFactor() int {
	halfWidth := (gm.boardWidth - 1) / 2
	halfHeight := (gm.boardHeight - 1) / 2

	// Top left.
	topLeftSum := 0
	for i := 0; i < halfHeight; i++ {
		for j := 0; j < halfWidth; j++ {
			topLeftSum += len(gm.robotMap[i][j])
		}
	}

	// Top right.
	topRightSum := 0
	for i := 0; i < halfHeight; i++ {
		for j := halfWidth + 1; j < gm.boardWidth; j++ {
			topRightSum += len(gm.robotMap[i][j])
		}
	}

	// Bottom left.
	bottomLeftSum := 0
	for i := halfHeight + 1; i < gm.boardHeight; i++ {
		for j := 0; j < halfWidth; j++ {
			bottomLeftSum += len(gm.robotMap[i][j])
		}
	}

	// Bottom right.
	bottomRightSum := 0
	for i := halfHeight + 1; i < gm.boardHeight; i++ {
		for j := halfWidth + 1; j < gm.boardWidth; j++ {
			bottomRightSum += len(gm.robotMap[i][j])
		}
	}

	return topLeftSum * topRightSum * bottomLeftSum * bottomRightSum
}

func (gm *gameMap) inBounds(c coordinate) bool {
	if c.y < 0 || c.y >= gm.boardHeight {
		return false
	}
	if c.x < 0 || c.x >= gm.boardWidth {
		return false
	}
	return true

}

func (gm *gameMap) neighbourCoordinates(c coordinate) []coordinate {
	neighbours := make([]coordinate, 0)
	up := coordinate{c.x, c.y - 1}
	if gm.inBounds(up) {
		neighbours = append(neighbours, up)
	}
	right := coordinate{c.x + 1, c.y}
	if gm.inBounds(right) {
		neighbours = append(neighbours, right)
	}
	down := coordinate{c.x, c.y + 1}
	if gm.inBounds(down) {
		neighbours = append(neighbours, down)
	}
	left := coordinate{c.x - 1, c.y}
	if gm.inBounds(left) {
		neighbours = append(neighbours, left)
	}
	return neighbours
}

func (gm *gameMap) biggestClump() int {
	exploredTiles := make(map[coordinate]struct{})
	biggestClump := 0
	for i := range gm.robotMap {
		for j := range gm.robotMap[i] {
			currentClump := 0
			frontier := []coordinate{{j, i}}
			for len(frontier) > 0 {
				newFrontier := make([]coordinate, 0)
				for _, c := range frontier {
					_, seen := exploredTiles[c]
					if seen {
						continue
					}
					exploredTiles[c] = struct{}{}
					botsAtCurrent := len(gm.robotMap[c.y][c.x])
					if botsAtCurrent == 0 {
						continue
					}
					currentClump += botsAtCurrent

					neighbourCandidates := gm.neighbourCoordinates(c)
					for _, neighbourCandidate := range neighbourCandidates {
						_, seen := exploredTiles[neighbourCandidate]
						if seen {
							continue
						}
						newFrontier = append(newFrontier, neighbourCandidate)
					}
				}
				frontier = newFrontier
			}
			if currentClump > biggestClump {
				biggestClump = currentClump
			}
		}
	}
	return biggestClump
}

func (gm *gameMap) printMap() {
	for _, row := range gm.robotMap {
		for _, col := range row {
			print(len(col), " ")
		}
		println()
	}
	println()
}

func (gm *gameMap) printMapSparse() {
	for _, row := range gm.robotMap {
		for _, col := range row {
			if len(col) > 0 {
				print("*")
			} else {
				print(" ")
			}
		}
		println()
	}
	println()
}

func handleLine(line string) robot {
	matches := lineRegex.FindStringSubmatch(line)
	if len(matches) != 5 {
		log.Fatalf("Expected 5 matches, got %d", len(matches))
	}
	positionX, err := strconv.Atoi(matches[1])
	if err != nil {
		log.Fatalf("Failed to parse position X: %v", err)
	}
	positionY, err := strconv.Atoi(matches[2])
	if err != nil {
		log.Fatalf("Failed to parse position Y: %v", err)
	}
	velocityX, err := strconv.Atoi(matches[3])
	if err != nil {
		log.Fatalf("Failed to parse velocity X: %v", err)
	}
	velocityY, err := strconv.Atoi(matches[4])
	if err != nil {
		log.Fatalf("Failed to parse velocity Y: %v", err)
	}

	return robot{
		position: coordinate{positionX, positionY},
		velocity: coordinate{velocityX, velocityY},
	}
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	robots := make([]robot, 0)
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		robots = append(robots, handleLine(scanner.Text()))
	}

	gm := newGameMap(robots, 101, 103)
	for range 100 {
		gm = gm.iterate()
	}

	// part 1
	println(gm.safetyFactor())

	// Part 2

	gm = newGameMap(robots, 101, 103)
	for i := range 10_000 {
		// Search for clumped robots on the assumption the tree will involve
		// the robots being grouped to draw.
		if gm.biggestClump() > 200 {
			println(gm.biggestClump())
			println(i)
			gm.printMapSparse()
		}
		gm = gm.iterate()
	}
}
