package main

import (
	"bufio"
	_ "embed"
	"github.com/samber/lo"
	"strings"
)

//go:embed input
var input string

func handleLine(line string) []string {
	return lo.ChunkString(line, 1)
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		if scanner.Text() == "" {
			// Skip blank lines.
			continue
		}
	}
}
