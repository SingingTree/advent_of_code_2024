package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"github.com/samber/lo"
	"gonum.org/v1/gonum/stat/combin"
	"slices"
	"strconv"
	"strings"
)

//go:embed input
var input string

func handleWireLine(line string) (string, int) {
	tokens := strings.Split(line, ": ")
	if len(tokens) != 2 {
		panic("Expected 2 tokens!")
	}

	value, err := strconv.Atoi(tokens[1])
	if err != nil {
		panic(err)
	}

	return tokens[0], value
}

func handleGateLine(line string) gate {
	tokens := strings.Split(line, " ")
	if len(tokens) != 5 {
		panic("Expected 5 tokens!")
	}
	lhs := tokens[0]
	rhs := tokens[2]
	outputWire := tokens[4]
	var gt gateType
	switch tokens[1] {
	case "AND":
		gt = and
	case "OR":
		gt = or
	case "XOR":
		gt = xor
	default:
		panic("Invalid gate type during parsing!")
	}

	return gate{
		lhsWire:    lhs,
		rhsWire:    rhs,
		outputWire: outputWire,
		gateType:   gt,
	}
}

type gateType int

const (
	and gateType = iota
	or
	xor
)

type gate struct {
	lhsWire    string
	rhsWire    string
	outputWire string
	gateType   gateType
}

func (g *gate) outputValue(lhs int, rhs int) int {
	switch g.gateType {
	case and:
		return lhs & rhs
	case or:
		return lhs | rhs
	case xor:
		return lhs ^ rhs
	default:
		panic("Unexpected gate type")
	}
}

type wireSolver struct {
	wireValues map[string]int
	gates      []gate
}

// Returns true if no further iteration has taken place.
func (ws *wireSolver) iterateOutputs() bool {
	hasProgressed := false

	for _, g := range ws.gates {
		if _, solved := ws.wireValues[g.outputWire]; solved {
			continue
		}

		lhsValue, lhsValueFound := ws.wireValues[g.lhsWire]
		if !lhsValueFound {
			continue
		}

		rhsValue, rhsValueFound := ws.wireValues[g.rhsWire]
		if !rhsValueFound {
			continue
		}

		output := g.outputValue(lhsValue, rhsValue)
		ws.wireValues[g.outputWire] = output
		hasProgressed = true
	}

	return !hasProgressed
}

func powerOfTwo(n int) int {
	if n == 0 {
		return 1
	}
	pow := 2
	for range n - 1 {
		pow *= 2
	}
	return pow
}

func (ws *wireSolver) outputValue(wirePrefix string) int {
	maxWireNum := 0
	for wire, _ := range ws.wireValues {
		if strings.HasPrefix(wire, wirePrefix) {
			numString := strings.TrimPrefix(wire, wirePrefix)
			num, err := strconv.Atoi(numString)
			if err != nil {
				panic(err)
			}
			if num > maxWireNum {
				maxWireNum = num
			}
		}
	}

	wireValues := make([]int, maxWireNum+1)
	for wire, value := range ws.wireValues {
		if strings.HasPrefix(wire, wirePrefix) {
			numString := strings.TrimPrefix(wire, wirePrefix)
			idx, err := strconv.Atoi(numString)
			if err != nil {
				panic(err)
			}
			wireValues[idx] = value
		}
	}

	return lo.ReduceRight(wireValues, func(agg int, item int, i int) int {
		return agg + (item * powerOfTwo(i))
	}, 0)
}

func (ws *wireSolver) checkXYZConsistentForExample() bool {
	xOutput := ws.outputValue("x")
	yOutput := ws.outputValue("y")
	zOutput := ws.outputValue("z")

	return xOutput&yOutput == zOutput
}

func (ws *wireSolver) checkXYZConsistent() bool {
	xOutput := ws.outputValue("x")
	yOutput := ws.outputValue("y")
	zOutput := ws.outputValue("z")

	return xOutput+yOutput == zOutput
}

func (ws *wireSolver) swapOutput(gateIdx1, gateIdx2 int) *wireSolver {
	clonedValues := make(map[string]int)
	for k, v := range ws.wireValues {
		clonedValues[k] = v
	}
	clonedGates := slices.Clone(ws.gates)
	tmp := clonedGates[gateIdx1].outputWire
	clonedGates[gateIdx1].outputWire = clonedGates[gateIdx2].outputWire
	clonedGates[gateIdx2].outputWire = tmp

	return &wireSolver{
		wireValues: clonedValues,
		gates:      clonedGates,
	}
}

type swapGenerator struct {
	permutations    [][]int
	permutationsIdx int
	combinGenerator *combin.CombinationGenerator
	currentCombin   []int
	swapsBuffer     []int
}

func newSwapGenerator(numGates int, numSwapPairs int) swapGenerator {
	permutations := combin.Permutations(numSwapPairs*2, numSwapPairs*2)
	permutationsFirstFiltering := make([][]int, 0)
	// Some permutations will produce redundant pairs, so filter them out.
firstPermutationCheckLoop:
	for _, permutation := range permutations {
		for i := 0; i < len(permutation); i += 2 {
			// Filter out permutations where the pair (i, j) has (j < i). E.g.
			// because we'll get (0, 1) as a pair from our permutation, we don't
			// need also to get (1, 0), as these are equivalent swaps.
			if permutation[i] > permutation[i+1] {
				continue firstPermutationCheckLoop
			}
		}
		permutationsFirstFiltering = append(permutationsFirstFiltering, permutation)
	}
	permutationsSecondFiltering := make([][]int, 0)
secondPermutationCheckLoop:
	for _, permutation := range permutationsFirstFiltering {
		for i := 0; i < len(permutation)-2; i += 2 {
			// Skip permutations where the first element in each pair is not
			// less than the first element in the following pair. This is a bit
			// weird, but I *think* this works and avoid redundant combinations.
			// E.g. consider the pairs (0, 2), (1, 3), (4, 5). An equivalent
			// group  of pairs exists in (1, 3), (4, 5), (0, 2), and we can cull
			// that via the following check.
			if permutation[i] > permutation[i+2] {
				continue secondPermutationCheckLoop
			}
		}
		permutationsSecondFiltering = append(permutationsSecondFiltering, permutation)
	}

	gen := combin.NewCombinationGenerator(numGates, numSwapPairs*2)
	combinationBuffer := make([]int, numSwapPairs*2)
	generated := gen.Next()
	if !generated {
		panic("Should get first combination!")
	}
	gen.Combination(combinationBuffer)
	return swapGenerator{
		permutations:    permutationsSecondFiltering,
		permutationsIdx: 0,
		combinGenerator: gen,
		currentCombin:   combinationBuffer,
		swapsBuffer:     make([]int, numSwapPairs*2),
	}
}

func (sg *swapGenerator) gen() ([]int, bool) {
	for i, idx := range sg.permutations[sg.permutationsIdx] {
		sg.swapsBuffer[i] = sg.currentCombin[idx]
	}

	sg.permutationsIdx += 1
	hasNext := true
	if sg.permutationsIdx == len(sg.permutations) {
		sg.permutationsIdx = 0
		hasNext = sg.combinGenerator.Next()
		if hasNext {
			sg.combinGenerator.Combination(sg.currentCombin)
		}
	}

	return sg.swapsBuffer, hasNext
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	ws := wireSolver{
		wireValues: make(map[string]int),
		gates:      make([]gate, 0),
	}

	scanningWires := true

	for scanner.Scan() {
		if scanner.Text() == "" {
			scanningWires = false
			continue
		}

		if scanningWires {
			wire, value := handleWireLine(scanner.Text())
			ws.wireValues[wire] = value
		} else {
			g := handleGateLine(scanner.Text())
			ws.gates = append(ws.gates, g)
		}
	}

	// Part 1.
	// Create a clone of ws by doing a redundant swap, so we can re-use ws for p2.
	p1Ws := ws.swapOutput(0, 0)
	for !p1Ws.iterateOutputs() {
	}
	println(p1Ws.outputValue("z"))

	// Part 2.
	//swapPairs := 2 // Uncomment for example.
	swapPairs := 4
	sg := newSwapGenerator(len(ws.gates), swapPairs)
	for swaps, hasNext := sg.gen(); hasNext; swaps, hasNext = sg.gen() {
		//fmt.Println(swaps)
		swappedWireSolver := &ws
		for i := 0; i < len(swaps); i += 2 {
			swappedWireSolver = swappedWireSolver.swapOutput(swaps[i], swaps[i+1])
			//fmt.Printf(
			//	"swapped [%s] <-> [%s]\n",
			//	swappedWireSolver.gates[swaps[i]].outputWire,
			//	swappedWireSolver.gates[swaps[i+1]].outputWire,
			//)
		}
		for swappedWireSolver.iterateOutputs() {
		}
		if swappedWireSolver.checkXYZConsistent() {
			swappedOutputs := make([]string, 0)
			for _, swap := range swaps {
				swappedOutputs = append(swappedOutputs, ws.gates[swap].outputWire)
			}
			slices.Sort(swappedOutputs)
			fmt.Println(strings.Join(swappedOutputs, ","))
			break
		}
	}

}
