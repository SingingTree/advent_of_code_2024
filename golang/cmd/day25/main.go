package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"github.com/samber/lo"
	"strings"
)

//go:embed input
var input string

func handleLine(line string) []bool {
	chunks := lo.ChunkString(line, 1)
	boolMap := make([]bool, len(chunks))
	for i := range chunks {
		boolMap[i] = chunks[i] == "#"
	}
	return boolMap
}

type lock struct {
	maxHeight int
	heights   []int
}

type key struct {
	heights []int
}

func canUnlock(l lock, k key) bool {
	for i := range l.heights {
		if l.heights[i]+k.heights[i] > l.maxHeight {
			return false
		}
	}
	return true
}

func assembleLockOrKey(pieces [][]bool) (*lock, *key) {
	isLock := pieces[0][0]

	trimmedPieces := pieces[1:]
	heights := make([]int, len(trimmedPieces[0]))

	for col := range trimmedPieces[0] {
		height := 0
		for row := range heights {
			if trimmedPieces[row][col] {
				height += 1
			}
		}
		heights[col] = height
	}

	if isLock {
		return &lock{
			maxHeight: len(trimmedPieces[0]),
			heights:   heights,
		}, nil
	} else {
		return nil, &key{heights: heights}
	}
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	pieces := make([][]bool, 0)
	locks := make([]lock, 0)
	keys := make([]key, 0)

	for scanner.Scan() {
		if scanner.Text() == "" {
			lock, key := assembleLockOrKey(pieces)
			if lock != nil {
				locks = append(locks, *lock)
			} else {
				keys = append(keys, *key)
			}
			pieces = make([][]bool, 0)
			continue
		}

		pieces = append(pieces, handleLine(scanner.Text()))
	}
	// Make the last lock/key.
	if len(pieces) > 0 {
		lock, key := assembleLockOrKey(pieces)
		if lock != nil {
			locks = append(locks, *lock)
		} else {
			keys = append(keys, *key)
		}
		pieces = make([][]bool, 0)
	}

	numUnlocks := 0
	for _, lock := range locks {
		for _, key := range keys {
			if canUnlock(lock, key) {
				numUnlocks++
			}
		}
	}
	// Not 98 or 4326
	fmt.Println(numUnlocks)
}
