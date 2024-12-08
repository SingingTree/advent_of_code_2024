package main

import (
	"bufio"
	_ "embed"
	"github.com/samber/lo"
	"log"
	"strings"
)

//go:embed input
var input string

type coordinate struct {
	row, col int
}

func (c *coordinate) sum(other coordinate) coordinate {
	return coordinate{c.row + other.row, c.col + other.col}
}

func (c *coordinate) eq(other coordinate) bool {
	if c.row == other.row && c.col == other.col {
		return true
	}
	return false
}

type gameMap struct {
	rawMap [][]string

	freqToAntennas map[string][]coordinate
}

func (gm *gameMap) isOffMap(coord coordinate) bool {
	if coord.row < 0 || coord.col < 0 {
		return true
	}
	if coord.row >= len(gm.rawMap) || coord.col >= len(gm.rawMap[0]) {
		return true
	}
	return false
}

func (gm *gameMap) calculateAntinodesPartOne() map[coordinate]struct{} {
	antinodes := make(map[coordinate]struct{})
	for _, antennas := range gm.freqToAntennas {
		for i, lhsAntennaCoords := range antennas {
			for j, rhsAntennaCoords := range antennas {
				if i == j {
					// Don't calculate antinodes with self.
					continue
				}
				differenceOne := coordinate{
					lhsAntennaCoords.row - rhsAntennaCoords.row,
					lhsAntennaCoords.col - rhsAntennaCoords.col,
				}
				differenceTwo := coordinate{
					rhsAntennaCoords.row - lhsAntennaCoords.row,
					rhsAntennaCoords.col - lhsAntennaCoords.col,
				}
				antinodeCandates := []coordinate{
					lhsAntennaCoords.sum(differenceOne),
					lhsAntennaCoords.sum(differenceTwo),
					rhsAntennaCoords.sum(differenceOne),
					rhsAntennaCoords.sum(differenceTwo),
				}
				for _, candate := range antinodeCandates {
					if candate.eq(lhsAntennaCoords) || candate.eq(rhsAntennaCoords) {
						// Skip antinodes on stations, which 2 candidates will be.
						continue
					}
					if gm.isOffMap(candate) {
						// Skip candidates off the map.
						continue
					}
					antinodes[candate] = struct{}{}
				}
			}
		}
	}
	return antinodes
}

func (gm *gameMap) calculateAntinodesPartTwo() map[coordinate]struct{} {
	antinodes := make(map[coordinate]struct{})
	for _, antennas := range gm.freqToAntennas {
		for i, lhsAntennaCoords := range antennas {
			for j, rhsAntennaCoords := range antennas {
				if i == j {
					// Don't calculate antinodes with self.
					continue
				}
				differenceOne := coordinate{
					lhsAntennaCoords.row - rhsAntennaCoords.row,
					lhsAntennaCoords.col - rhsAntennaCoords.col,
				}
				differenceTwo := coordinate{
					rhsAntennaCoords.row - lhsAntennaCoords.row,
					rhsAntennaCoords.col - lhsAntennaCoords.col,
				}

				// This is a bit yuck, but the problem is small enough we can brute force it.
				antinodeCandates := make([]coordinate, 0)
				candidate := lhsAntennaCoords.sum(differenceOne)
				for !gm.isOffMap(candidate) {
					antinodeCandates = append(antinodeCandates, candidate)
					candidate = candidate.sum(differenceOne)
				}
				candidate = rhsAntennaCoords.sum(differenceOne)
				for !gm.isOffMap(candidate) {
					antinodeCandates = append(antinodeCandates, candidate)
					candidate = candidate.sum(differenceOne)
				}
				candidate = lhsAntennaCoords.sum(differenceTwo)
				for !gm.isOffMap(candidate) {
					antinodeCandates = append(antinodeCandates, candidate)
					candidate = candidate.sum(differenceOne)
				}
				candidate = rhsAntennaCoords.sum(differenceTwo)
				for !gm.isOffMap(candidate) {
					antinodeCandates = append(antinodeCandates, candidate)
					candidate = candidate.sum(differenceOne)
				}

				for _, candate := range antinodeCandates {
					if gm.isOffMap(candate) {
						// Skip candidates off the map.
						log.Fatal("Should be unreachable!")
					}
					antinodes[candate] = struct{}{}
				}
			}
		}
	}
	return antinodes
}

func createGameMap(rawMap [][]string) gameMap {
	freqToAntennas := make(map[string][]coordinate)
	for row := range rawMap {
		for col := range rawMap[row] {
			if rawMap[row][col] != "." {
				freq := rawMap[row][col]
				if _, ok := freqToAntennas[freq]; !ok {
					freqToAntennas[freq] = make([]coordinate, 0)
				}
				freqToAntennas[freq] = append(freqToAntennas[freq], coordinate{row, col})
			}
		}
	}

	return gameMap{
		rawMap:         rawMap,
		freqToAntennas: freqToAntennas,
	}
}

func handleLine(line string) []string {
	return lo.ChunkString(line, 1)
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))
	rawMap := make([][]string, 0)

	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		rawMap = append(rawMap, handleLine(scanner.Text()))
	}

	gm := createGameMap(rawMap)
	antinodes := gm.calculateAntinodesPartOne()
	println(len(antinodes))

	antinodes = gm.calculateAntinodesPartTwo()
	println(len(antinodes))
}
