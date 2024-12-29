package main

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"gonum.org/v1/gonum/stat/combin"
	"maps"
	"math/rand"
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

// Returns true if progress has been made (i.e. should be called again).
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

	return hasProgressed
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

func (ws *wireSolver) getExpectedOutputForExample() int {
	xOutput := ws.outputValue("x")
	yOutput := ws.outputValue("y")

	return xOutput & yOutput
}

func (ws *wireSolver) getExpectedOutput() int {
	xOutput := ws.outputValue("x")
	yOutput := ws.outputValue("y")

	return xOutput + yOutput
}

func (ws *wireSolver) getExpectedOutputShim() int {
	return ws.getExpectedOutput()
}

// checkBitMatch returns a bool slice indicating if the bits in expected
// match those in actual. If expected and actual are not the same length,
// the shorter of the two will be padded with 0s
func checkBitMatch(expected int, actual int) []bool {
	expectedBitString := strconv.FormatInt(int64(expected), 2)
	actualBitString := strconv.FormatInt(int64(actual), 2)
	for len(expectedBitString) < len(actualBitString) {
		expectedBitString = "0" + expectedBitString
	}
	for len(actualBitString) < len(expectedBitString) {
		actualBitString = "0" + actualBitString
	}
	matchSlice := make([]bool, len(expectedBitString))
	for i := range expectedBitString {
		matchSlice[i] = expectedBitString[i] == actualBitString[i]
	}
	return matchSlice
}

func (ws *wireSolver) clone() *wireSolver {
	clonedValues := make(map[string]int)
	for k, v := range ws.wireValues {
		clonedValues[k] = v
	}
	clonedGates := slices.Clone(ws.gates)

	return &wireSolver{
		wireValues: clonedValues,
		gates:      clonedGates,
	}
}

func (ws *wireSolver) swapOutput(gateIdx1, gateIdx2 int) {
	tmp := ws.gates[gateIdx1].outputWire
	ws.gates[gateIdx1].outputWire = ws.gates[gateIdx2].outputWire
	ws.gates[gateIdx2].outputWire = tmp
}

func (ws *wireSolver) randomizeValues() {
	for k := range ws.wireValues {
		if !strings.HasPrefix(k, "x") && !strings.HasPrefix(k, "y") {
			panic("This shouldn't be called if non-xy values have been set!")
		}
		// Set value to 0 or 1.
		ws.wireValues[k] = rand.Intn(2)
	}
}

func (ws *wireSolver) zeroWiresGreaterThanN(n int) {
	for k := range ws.wireValues {
		if strings.HasPrefix(k, "x") {
			intValue, err := strconv.Atoi(strings.TrimPrefix(k, "x"))
			if err != nil {
				panic(err)
			}
			if intValue > n {
				ws.wireValues[k] = 0
			}
		}
		if strings.HasPrefix(k, "y") {
			intValue, err := strconv.Atoi(strings.TrimPrefix(k, "y"))
			if err != nil {
				panic(err)
			}
			if intValue > n {
				ws.wireValues[k] = 0
			}
		}
	}
}

func (ws *wireSolver) findOutputDeps() (map[string][]gate, map[string][]string) {
	backwardsMap := make(map[string]gate)
	for _, g := range ws.gates {
		backwardsMap[g.outputWire] = g
	}

	gateOutputDepMap := make(map[string][]gate)
	flattenedOutputDepsMap := make(map[string][]string)
	for k, v := range backwardsMap {
		if strings.HasPrefix(k, "z") {
			wireDeps := make([]string, 0)
			gateDeps := make([]gate, 0)
			gateDeps = append(gateDeps, v)
			flatFrontier := []string{
				v.lhsWire,
				v.rhsWire,
			}
			for len(flatFrontier) > 0 {
				newFlatFrontier := make([]string, 0)
				for _, wire := range flatFrontier {
					if strings.HasPrefix(wire, "x") || strings.HasPrefix(wire, "y") {
						wireDeps = append(wireDeps, wire)
						continue
					}
					backGate := backwardsMap[wire]
					gateDeps = append(gateDeps, backGate)
					newFlatFrontier = append(newFlatFrontier, backGate.lhsWire, backGate.rhsWire)
				}
				flatFrontier = newFlatFrontier
			}
			gateOutputDepMap[k] = gateDeps
			flattenedOutputDepsMap[k] = wireDeps

		}
	}

	for k, v := range flattenedOutputDepsMap {
		flattenedOutputDepsMap[k] = lo.Uniq(v)
	}
	return gateOutputDepMap, flattenedOutputDepsMap
}

type adderGate struct {
	// For zNN, where NN is some integer, there should be a gate that has
	// inputs xNN, yNN, XORs them, and outputs to a wire feeding outputXor.
	xyXor gate
	// For zNN, where NN is some integer, there should be a gate that has
	// inputs xNN-1, yNN-1, ANDs them, and outputs to a wire feeding carryInOr.
	prevXyAnd gate
	// prevCarryOutAnd handles part of the logic of carrying out from zNN-1.
	// It should have input wires from zNN-1's xyXor and zNN-1's carryInOr.
	// It outputs to a wire feeding carryInOr.
	prevCarryOutAnd gate
	// carryInOr carries in from the adding from zNN-1. It should have an input
	// wire from prevXyAnd, and another from prevCarryOutAnd. It should output
	// to a wire feeding outputXor.
	carryInOr gate
	// outputXor sets zNN. It should have an input wire from xyXor, and another
	// from carryInOr.
	outputXor gate
}

func tryConstructAdderGate(gates []gate) (adderGate, error) {
	var carryInOr gate
	count := 0
	for _, g := range gates {
		if g.gateType == or {
			carryInOr = g
			count += 1
			if count > 1 {
				return adderGate{}, errors.New("too many OR gates")
			}
		}
	}
	if count == 0 {
		return adderGate{}, errors.New("no OR gates")
	}
	var xyXor gate
	count = 0
	for _, g := range gates {
		xyLhs := strings.HasPrefix(g.lhsWire, "x") || strings.HasPrefix(g.lhsWire, "y")
		xyRhs := strings.HasPrefix(g.rhsWire, "x") || strings.HasPrefix(g.rhsWire, "y")
		if g.gateType == xor && xyLhs && xyRhs {
			xyXor = g
			count += 1
			if count > 1 {
				return adderGate{}, errors.New("too many xyXor gates")
			}
		}
	}
	if count == 0 {
		return adderGate{}, errors.New("no xyXor gates")
	}
	var outputXor gate
	count = 0
	for _, g := range gates {
		if g.lhsWire == xyXor.outputWire || g.rhsWire == xyXor.outputWire {
			outputXor = g
			count += 1
			if count > 1 {
				return adderGate{}, errors.New("too many outputXor gates")
			}
		}
	}
	if count == 0 {
		return adderGate{}, errors.New("no outputXor gates")
	}
	var carryInOrOutputWire string
	if outputXor.lhsWire == xyXor.outputWire {
		carryInOrOutputWire = outputXor.rhsWire
	} else {
		carryInOrOutputWire = outputXor.lhsWire
	}
	if carryInOr.outputWire != carryInOrOutputWire {
		return adderGate{}, errors.New("carryInOr outputWire doesn't match outputXor input")
	}

	var prevXyAnd gate
	count = 0
	for _, g := range gates {
		xyLhs := strings.HasPrefix(g.lhsWire, "x") || strings.HasPrefix(g.lhsWire, "y")
		xyRhs := strings.HasPrefix(g.rhsWire, "x") || strings.HasPrefix(g.rhsWire, "y")
		if g.gateType == and && xyLhs && xyRhs {
			prevXyAnd = g
			count += 1
			if count > 1 {
				return adderGate{}, errors.New("too many prevXyAnd gates")
			}
		}
	}
	if count == 0 {
		return adderGate{}, errors.New("no prevXyAnd gates")
	}
	if prevXyAnd.outputWire != carryInOr.lhsWire && prevXyAnd.outputWire != carryInOr.rhsWire {
		return adderGate{}, errors.New("prevXyAnd outputWire doesn't match carryInOr input")
	}

	var prevCarryOutAnd gate
	count = 0
	for _, g := range gates {
		if g != xyXor && g != prevXyAnd && g != carryInOr && g != outputXor {
			prevCarryOutAnd = g
			count += 1
			if count > 1 {
				return adderGate{}, errors.New("too many prevCarryOutAnd gates")
			}
		}
	}
	if count == 0 {
		return adderGate{}, errors.New("no prevCarryOutAnd gates")
	}

	return adderGate{
		xyXor,
		prevXyAnd,
		prevCarryOutAnd,
		carryInOr,
		outputXor,
	}, nil
}

// diagnoseAdderGates tries to diagnose issues with the adderGate responsible
// for the bit at outputBit index. It returns a list of swaps it thinks should
// be made, where each 2 items in the list are indices for gates that should
// have outputs swapped.
func (ws *wireSolver) diagnoseAdderGates(outputBit int) []int {
	gateOutputDepMap, _ := ws.findOutputDeps()
	keys := slices.Collect(maps.Keys(gateOutputDepMap))
	slices.Sort(keys)

	gatesForOutput := make([][]gate, outputBit)
	for i := range outputBit {
		gates := make([]gate, 0)
		for _, g := range gateOutputDepMap[keys[i]] {
			gates = append(gates, g)
		}
		gatesForOutput[i] = gates
	}

	newGatesForOutput := make([][]gate, outputBit)
	for i := 0; i < len(gatesForOutput)-1; i += 1 {
		newGates := make([]gate, 0)
		for _, g1 := range gatesForOutput[i+1] {
			newGate := true
			for _, g2 := range gatesForOutput[i] {
				if g1 == g2 {
					newGate = false
					break
				}
			}
			if newGate {
				newGates = append(newGates, g1)
			}
		}
		newGatesForOutput[i+1] = newGates
	}

	for _, k := range keys {
		trimmedNum, ok := strings.CutPrefix(k, "z")
		if !ok {
			panic("Should always have z prefix")
		}
		keyNum, err := strconv.Atoi(trimmedNum)
		if err != nil {
			panic(err)
		}
		expectedAndGates := (keyNum-1)*2 + 1
		actualAndGates := 0
		if keyNum == 0 {
			expectedAndGates = 0
		}
		expectedXorGates := keyNum + 1
		actualXorGates := 0
		for _, g := range gateOutputDepMap[k] {
			if g.gateType == and {
				actualAndGates += 1
			} else if g.gateType == xor {
				actualXorGates += 1
			}
		}

		if keyNum < outputBit {
			if expectedAndGates != actualAndGates {
				panic("And gates wrong")
			}
			if expectedXorGates != actualXorGates {
				panic("Xor gates wrong")
			}
		}
	}

	adderGates := make([]adderGate, outputBit)
	for i := 2; i < outputBit; i += 1 {
		ag, err := tryConstructAdderGate(newGatesForOutput[i])
		if err != nil {
			panic(err)
		}
		adderGates[i] = ag
	}
	// Check gates are sensible.
	for i := 2; i < len(adderGates)-1; i += 1 {
		ag1 := adderGates[i]
		ag2 := adderGates[i+1]
		hasCarryWiring1 := ag2.prevCarryOutAnd.lhsWire == ag1.xyXor.outputWire ||
			ag2.prevCarryOutAnd.rhsWire == ag1.xyXor.outputWire
		hasCarryWiring2 := ag2.prevCarryOutAnd.lhsWire == ag1.carryInOr.outputWire ||
			ag2.prevCarryOutAnd.rhsWire == ag1.carryInOr.outputWire
		if !hasCarryWiring1 || !hasCarryWiring2 {
			fmt.Println("Adder Gate at ", i, " has incorrect wiring for carrying")
		}
	}

	lastAddrGate := adderGates[len(adderGates)-1]
	lastOutputXor := lastAddrGate.outputXor
	// to find prevCarryOutAnd for our current gate we can search for a gate
	// with the same inputs as lastOutputXor.
	matchCandidates := make([]gate, 0)
	for _, g := range ws.gates {
		lhsMatch := false
		if g.lhsWire == lastOutputXor.lhsWire || g.rhsWire == lastOutputXor.lhsWire {
			lhsMatch = true
		}
		rhsMatch := false
		if g.lhsWire == lastOutputXor.rhsWire || g.rhsWire == lastOutputXor.rhsWire {
			rhsMatch = true
		}
		if !lhsMatch || !rhsMatch {
			continue
		}
		if g == lastOutputXor {
			continue
		}
		if g.gateType != and {
			continue
		}
		matchCandidates = append(matchCandidates, g)
	}
	if len(matchCandidates) != 1 {
		panic("Expected to find 1 candidate")
	}
	prevCarryOutAnd := matchCandidates[0]
	// xyXor should be unique.
	matchCandidates = make([]gate, 0)
	for _, g := range ws.gates {
		if (g.lhsWire == fmt.Sprintf("x%02d", outputBit) ||
			g.lhsWire == fmt.Sprintf("y%02d", outputBit)) &&
			(g.rhsWire == fmt.Sprintf("x%02d", outputBit) ||
				g.rhsWire == fmt.Sprintf("y%02d", outputBit)) &&
			g.gateType == xor {
			matchCandidates = append(matchCandidates, g)
		}
	}
	if len(matchCandidates) != 1 {
		panic("Expected to find 1 candidate")
	}
	xyXor := matchCandidates[0]
	// Find prevXyAnd, it should be unique.
	matchCandidates = make([]gate, 0)
	for _, g := range ws.gates {
		if (g.lhsWire == fmt.Sprintf("x%02d", outputBit-1) ||
			g.lhsWire == fmt.Sprintf("y%02d", outputBit-1)) &&
			(g.rhsWire == fmt.Sprintf("x%02d", outputBit-1) ||
				g.rhsWire == fmt.Sprintf("y%02d", outputBit-1)) &&
			g.gateType == and {
			matchCandidates = append(matchCandidates, g)
		}
	}
	if len(matchCandidates) != 1 {
		panic("Expected to find 1 candidate")
	}
	prevXyAnd := matchCandidates[0]
	// We can use the outputs from prevXyAnd and prevCarryOutAnd to find
	// carryInOr.
	matchCandidates = make([]gate, 0)
	for _, g := range ws.gates {
		if (g.lhsWire == prevXyAnd.outputWire ||
			g.lhsWire == prevCarryOutAnd.outputWire) &&
			(g.rhsWire == prevXyAnd.outputWire ||
				g.rhsWire == prevCarryOutAnd.outputWire) &&
			g.gateType == or {
			matchCandidates = append(matchCandidates, g)
		}
	}
	if len(matchCandidates) != 1 {
		panic("Expected to find 1 candidate")
	}
	carryInOr := matchCandidates[0]
	// We can use the outputs from carryInOr and xyXor to find outputXor.
	matchCandidates = make([]gate, 0)
	for _, g := range ws.gates {
		if (g.lhsWire == carryInOr.outputWire ||
			g.lhsWire == xyXor.outputWire) &&
			(g.rhsWire == carryInOr.outputWire ||
				g.rhsWire == xyXor.outputWire) &&
			g.gateType == xor {
			matchCandidates = append(matchCandidates, g)
		}
	}
	if len(matchCandidates) != 1 {
		panic("Expected to find 1 candidate")
	}
	outputXor := matchCandidates[0]

	ag := adderGate{
		xyXor:           xyXor,
		prevXyAnd:       prevXyAnd,
		prevCarryOutAnd: prevCarryOutAnd,
		carryInOr:       carryInOr,
		outputXor:       outputXor,
	}

	swaps := make([]int, 0)

	// See if the outputXor gate needs rewiring.
	matchCandidates = make([]gate, 0)
	for _, g := range ws.gates {
		if g.outputWire == fmt.Sprintf("z%02d", outputBit) {
			matchCandidates = append(matchCandidates, g)
		}
	}
	if len(matchCandidates) != 1 {
		panic("Expected to find 1 candidate")
	}
	// outputGate is the current gate writing the zNN output.
	outputGate := matchCandidates[0]
	if outputGate != ag.outputXor {
		// The output gate needs to be swapped so we're actually outputting
		// as expected.
		outputGateIndex := slices.Index(ws.gates, outputGate)
		outputXorIndex := slices.Index(ws.gates, ag.outputXor)
		swaps = append(swaps, outputGateIndex, outputXorIndex)
	}

	return swaps
}

func checkLastNBitsOfBitMatch(bitMatchSlice []bool, lastNBits int) bool {
	for i := len(bitMatchSlice) - 1; i >= 0 && i > len(bitMatchSlice)-lastNBits; i-- {
		if !bitMatchSlice[i] {
			return false
		}
	}
	return true
}

// jiggleCheck giggles the inputs and checks the expected output value to
// verify the bit is stable under different inputs.
func (ws *wireSolver) jiggleCheck(lastNBits int) bool {
	for range 100 {
		clone := ws.clone()
		clone.randomizeValues()
		clone.zeroWiresGreaterThanN(lastNBits - 1)
		expectedOutput := clone.getExpectedOutputShim()
		for clone.iterateOutputs() {
		}
		output := clone.outputValue("z")
		bitMatchSlice := checkBitMatch(expectedOutput, output)
		if !checkLastNBitsOfBitMatch(bitMatchSlice, lastNBits) {
			return false
		}
	}
	return true
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
	// Create a clone of ws so we can re-use ws for p2.
	p1Ws := ws.clone()
	for p1Ws.iterateOutputs() {
	}
	println(p1Ws.outputValue("z"))

	// Part 2.
	clone := ws.clone()
	allSwaps := make([]int, 0)
	// Testing shows gate 9 is busted.
	swaps := clone.diagnoseAdderGates(9)
	for i := 0; i < len(swaps); i += 2 {
		clone.swapOutput(swaps[i], swaps[i+1])
	}
	allSwaps = append(allSwaps, swaps...)
	// Gate 20 is busted too.
	swaps = clone.diagnoseAdderGates(20)
	for i := 0; i < len(swaps); i += 2 {
		clone.swapOutput(swaps[i], swaps[i+1])
	}
	allSwaps = append(allSwaps, swaps...)

	// Brute force the rest. Can actually brute force everything, but
	// was interesting to do the wiring approach above.

	// We've already done a swap above.
	expectedNumSwaps := 2
	combinations := combin.Combinations(len(ws.gates), 2)
	expectedOutput := ws.getExpectedOutputShim()
	expectedOutputBinStr := strconv.FormatInt(int64(expectedOutput), 2)
	swaps = make([]int, 0)
	for i := len(expectedOutputBinStr) - 1; i >= 0; i -= 1 {
		combinationIndex := 0
		numSwaps := 1
		swapIndices := combin.Combinations(len(combinations), numSwaps)
		var swapCandidates []int
		lastNBits := len(expectedOutputBinStr) - i
		for {
			// Clone so we can preserve the original.
			baseClone := clone.clone()
			for i := 0; i < len(swapCandidates); i += 2 {
				baseClone.swapOutput(swapCandidates[i], swapCandidates[i+1])
			}
			for i := 0; i < len(swaps); i += 2 {
				baseClone.swapOutput(swaps[i], swaps[i+1])
			}

			clone := baseClone.clone()

			expectedCloneOutput := clone.getExpectedOutputShim()

			for clone.iterateOutputs() {
			}
			bitMatchSlice := checkBitMatch(expectedCloneOutput, clone.outputValue("z"))
			if checkLastNBitsOfBitMatch(bitMatchSlice, lastNBits) {
				clone := baseClone.clone()
				if clone.jiggleCheck(lastNBits) {
					break
				}
			}
			swapCandidates = make([]int, numSwaps*2)
			for j, idx := range swapIndices[combinationIndex] {
				swap := combinations[idx]
				swapCandidates[j*2] = swap[0]
				swapCandidates[j*2+1] = swap[1]
			}
			combinationIndex += 1
			if combinationIndex >= len(swapIndices) {
				numSwaps += 1

				if numSwaps+len(swaps)/2 > expectedNumSwaps {
					panic("Too many swaps!")
				}

				combinationIndex = 0
				swapIndices = combin.Combinations(len(combinations), numSwaps)
			}
		}
		if len(swapCandidates) > 0 {
			swaps = append(swaps, swapCandidates...)
			if len(swaps) == expectedNumSwaps*2 {
				break
			}
			combinations = lo.Filter(combinations, func(item []int, _ int) bool {
				for _, swapCandidate := range swapCandidates {
					if lo.Contains(item, swapCandidate) {
						return false
					}
				}
				return true
			})
		}
	}
	allSwaps = append(allSwaps, swaps...)
	swapStrings := make([]string, len(allSwaps))
	for i := range allSwaps {
		swapStrings[i] = ws.gates[allSwaps[i]].outputWire
	}
	slices.Sort(swapStrings)
	fmt.Println(strings.Join(swapStrings, ","))
	// not "nbq,snj,srq,vkt,wjf,z03,z29,z31"
	// not "ddn,kqh,nhs,nnf,z09,z20"
	// not "cdk,cdm,jdk,nhs,nnf,nnf,z09,z09"
}
