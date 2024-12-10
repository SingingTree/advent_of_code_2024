package main

import (
	"bufio"
	_ "embed"
	"github.com/samber/lo"
	"log"
	"slices"
	"strconv"
	"strings"
)

//go:embed input
var input string

type file struct {
	size int
	id   int
}

func handleLine(line string) ([]file, []int) {
	nums := lo.ChunkString(line, 1)
	files := make([]file, 0)
	freeSpace := make([]int, 0)
	isFile := true
	fileId := 0
	for _, numStr := range nums {
		num, err := strconv.Atoi(numStr)
		if err != nil {
			log.Fatal(err)
		}
		if isFile {
			files = append(files, file{size: num, id: fileId})
			fileId += 1
		} else {
			freeSpace = append(freeSpace, num)
		}
		isFile = !isFile
	}

	return files, freeSpace
}

func getDiskLayout(files []file, space []int) []int {
	disk := make([]int, 0)
	if len(files)-1 != len(space) {
		log.Fatal("Number of files is different than spaces!")
	}
	for i := range files {
		for range files[i].size {
			disk = append(disk, files[i].id)
		}
		if i == len(files)-1 {
			// Avoid going out of bounds since we have 1 more file than space.
			break
		}
		for range space[i] {
			disk = append(disk, -1)
		}
	}

	return disk
}

func compactDiskPart1(disk []int) []int {
	compactedDisk := make([]int, len(disk))
	copy(compactedDisk, disk)
	for slices.Contains(compactedDisk, -1) {
		emptyIndex := slices.Index(compactedDisk, -1)
		for j := len(compactedDisk) - 1; j > 0; j-- {
			if compactedDisk[j] != -1 {
				compactedDisk[emptyIndex] = compactedDisk[j]
				compactedDisk = compactedDisk[:j]
				break
			}
		}
	}
	return compactedDisk
}

type diskAddress struct {
	index int
	size  int
}

func findFreeSpaces(disk []int) []diskAddress {
	freeSpaces := make([]diskAddress, 0)
	inFreeSpace := false
	var freeSpaceCount, index int
	for i := range disk {
		if disk[i] == -1 {
			if !inFreeSpace {
				inFreeSpace = true
				index = i
			}
			freeSpaceCount += 1
		} else if inFreeSpace {
			// Close off the free space we were counting
			freeSpaces = append(freeSpaces, diskAddress{index: index, size: freeSpaceCount})
			freeSpaceCount = 0
			inFreeSpace = false
		}
	}

	return freeSpaces
}

func findFileNum(disk []int, fileNum int) diskAddress {
	index := slices.Index(disk, fileNum)
	if index < 0 {
		log.Fatal("Should find file!")
	}
	count := 0
	for i := index; i < len(disk); i++ {
		if disk[i] == fileNum {
			count++
		} else {
			break
		}
	}

	return diskAddress{index: index, size: count}
}

func compactDiskPart2(disk []int) []int {
	compactedDisk := make([]int, len(disk))
	copy(compactedDisk, disk)

	var fileNum int
	for j := len(compactedDisk) - 1; j > 0; j-- {
		if compactedDisk[j] != -1 {
			fileNum = compactedDisk[j]
			break
		}
	}

	for fileNum > 0 {
		freeSpaceLocations := findFreeSpaces(compactedDisk)
		fileLocation := findFileNum(compactedDisk, fileNum)
		for i := range freeSpaceLocations {
			if freeSpaceLocations[i].index > fileLocation.index {
				// Don't move files into later free space.
				break
			}
			if freeSpaceLocations[i].size >= fileLocation.size {
				for j := 0; j < fileLocation.size; j++ {
					index := freeSpaceLocations[i].index + j
					if compactedDisk[index] != -1 {
						log.Fatal("logic error")
					}
					compactedDisk[index] = fileNum
					index = fileLocation.index + j
					if compactedDisk[index] != fileNum {
						log.Fatal("logic error")
					}
					compactedDisk[index] = -1
				}
				break
			}
		}
		fileNum -= 1
	}

	return compactedDisk
}

func printDisk(disk []int) {
	for i := range disk {
		if disk[i] > -1 {
			print(disk[i])
		} else {
			print(".")
		}
	}
	println()
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	var files []file
	var freeSpace []int

	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		files, freeSpace = handleLine(scanner.Text())
	}

	disk := getDiskLayout(files, freeSpace)

	compactedDisk := compactDiskPart1(disk)
	checksum := 0
	for i := range compactedDisk {
		checksum += compactedDisk[i] * i
	}
	println(checksum)

	compactedDisk = compactDiskPart2(disk)
	checksum = 0
	for i := range compactedDisk {
		if compactedDisk[i] > -1 {
			checksum += compactedDisk[i] * i
		}
	}
	println(checksum)
}
