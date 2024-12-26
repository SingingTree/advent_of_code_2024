package main

import (
	"bufio"
	_ "embed"
	"errors"
	"github.com/dominikbraun/graph"
	"gonum.org/v1/gonum/stat/combin"
	"maps"
	"slices"
	"strings"
)

//go:embed input
var input string

func handleLine(line string) (string, string) {
	nodes := strings.Split(line, "-")
	if len(nodes) != 2 {
		panic("Expected 2 nodes!")
	}

	return nodes[0], nodes[1]
}

func findNConnectedSubNetworks(network graph.Graph[string, string], n int) [][]string {
	adjMap, err := network.AdjacencyMap()
	if err != nil {
		panic(err)
	}

	subNetworks := make([][]string, 0)
	subNetworksSeen := make(map[string]struct{})

	for outerNode, adjNodes := range adjMap {
		if len(adjNodes) < n-1 {
			continue
		}

		adjNodesSlice := make([]string, 0)
		for adjNode := range maps.Keys(adjNodes) {
			adjNodesSlice = append(adjNodesSlice, adjNode)
		}

		combinations := combin.Combinations(len(adjNodesSlice), n-1)
	combination_loop:
		for _, combination := range combinations {
			// Check that all values in the combination are able to reach all others.
			for _, idx1 := range combination {
				node1 := adjNodesSlice[idx1]
				for _, idx2 := range combination {
					node2 := adjNodesSlice[idx2]
					if idx1 == idx2 {
						continue
					}
					if _, ok := adjMap[node1][node2]; !ok {
						// This combination is invalid, try the next.
						continue combination_loop
					}
				}
			}
			// All nodes are reachable from one another.
			subNetwork := make([]string, 0)
			subNetwork = append(subNetwork, outerNode)
			for _, idx := range combination {
				subNetwork = append(subNetwork, adjNodesSlice[idx])
			}
			slices.Sort(subNetwork)
			subnetworkString := strings.Join(subNetwork, ",")
			if _, ok := subNetworksSeen[subnetworkString]; !ok {
				subNetworksSeen[subnetworkString] = struct{}{}
				subNetworks = append(subNetworks, subNetwork)
			}
		}
	}

	return subNetworks
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	network := graph.New(graph.StringHash)

	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}

		node1, node2 := handleLine(scanner.Text())
		_, err := network.Vertex(node1)
		if err != nil {
			if errors.Is(err, graph.ErrVertexNotFound) {
				err = network.AddVertex(node1)
				if err != nil {
					panic(err)
				}
			} else {
				panic(err)
			}
		}

		_, err = network.Vertex(node2)
		if err != nil {
			if errors.Is(err, graph.ErrVertexNotFound) {
				err = network.AddVertex(node2)
				if err != nil {
					panic(err)
				}
			} else {
				panic(err)
			}
		}

		err = network.AddEdge(node1, node2)
		if err != nil {
			panic(err)
		}
	}

	// Part 1.
	subNetworks := findNConnectedSubNetworks(network, 3)
	count := 0
	for _, subNetwork := range subNetworks {
		for _, computer := range subNetwork {
			if strings.HasPrefix(computer, "t") {
				count += 1
				break
			}
		}
	}
	println(count)

	// Part 2.
	order, err := network.Order()
	if err != nil {
		panic(err)
	}
	for n := order; n > 0; n-- {
		subNetworks := findNConnectedSubNetworks(network, n)
		if len(subNetworks) > 0 {
			println(strings.Join(subNetworks[0], ","))
			break
		}
	}
}
