package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"github.com/samber/lo"
	"regexp"
	"strconv"
	"strings"
)

//go:embed input
var input string

var registerLineRegex = regexp.MustCompile(`^Register ([ABC]): (\d+)$`)

func handleRegisterLine(line string) (string, int) {
	matches := registerLineRegex.FindStringSubmatch(line)
	if len(matches) != 3 {
		panic("Invalid register line")
	}

	integer, err := strconv.Atoi(matches[2])
	if err != nil {
		panic(err)
	}

	return matches[1], integer
}

func handleProgramLine(line string) []int {
	trimmedLine, found := strings.CutPrefix(line, "Program: ")
	if !found {
		panic("Invalid program line")
	}
	numStrings := strings.Split(trimmedLine, ",")

	nums := make([]int, len(numStrings))
	for i, numString := range numStrings {
		num, err := strconv.Atoi(numString)
		if err != nil {
			panic(fmt.Sprintf("Failed to convert program arg %v", numString))
		}
		nums[i] = num
	}

	return nums
}

type computer struct {
	registerA          int
	registerB          int
	registerC          int
	program            []int
	instructionPointer int
	outputBuffer       []int
}

func (c *computer) print() {
	println("A: ", c.registerA)
	println("B: ", c.registerB)
	println("C: ", c.registerC)
	print("Program: ")
	for _, i := range c.program {
		print(i, " ")
	}
	println()
	println("InstructionPointer: ", c.instructionPointer)
}

func (c *computer) decodeComboOperand(operand int) int {
	switch operand {
	case 1:
		fallthrough
	case 2:
		fallthrough
	case 3:
		return operand
	case 4:
		return c.registerA
	case 5:
		return c.registerB
	case 6:
		return c.registerC
	default:
		panic(fmt.Sprintf("Invalid operand: %d", operand))
	}
}

// MFW golang doesn't have integer power in its stdlib. Why would it?
// Just convert your ints to floats and back? No thanks.
func intPow(base, exponent int) int {
	if exponent == 0 {
		return 1
	}

	if exponent == 1 {
		return base
	}

	result := base
	for range exponent - 1 {
		result *= base
	}
	return result
}

// opcode 0.
func (c *computer) adv(operand int) {
	operandValue := c.decodeComboOperand(operand)
	divisor := intPow(2, operandValue)
	c.registerA /= divisor
}

// opcode 1.
func (c *computer) bxl(operand int) {
	c.registerB ^= operand
}

// opcode 2.
func (c *computer) bst(operand int) {
	operandValue := c.decodeComboOperand(operand)
	c.registerB = operandValue % 8
}

// opcode 3.
func (c *computer) jnz(operand int) bool {
	if c.registerA == 0 {
		return true
	}
	c.instructionPointer = operand
	return false
}

// opcode 4.
func (c *computer) bxc(_ int) {
	c.registerB ^= c.registerC
}

// opcode 5.
func (c *computer) out(operand int) {
	operandValue := c.decodeComboOperand(operand)
	c.outputBuffer = append(c.outputBuffer, operandValue%8)
}

// opcode 6.
func (c *computer) bdv(operand int) {
	operandValue := c.decodeComboOperand(operand)
	divisor := intPow(2, operandValue)
	c.registerB = c.registerA / divisor
}

// opcode 7.
func (c *computer) cdv(operand int) {
	operandValue := c.decodeComboOperand(operand)
	divisor := intPow(2, operandValue)
	c.registerC = c.registerA / divisor
}

func (c *computer) runInstruction() {
	opcode := c.program[c.instructionPointer]
	operand := c.program[c.instructionPointer+1]
	increaseInstructionPointer := true
	switch opcode {
	case 0:
		c.adv(operand)
	case 1:
		c.bxl(operand)
	case 2:
		c.bst(operand)
	case 3:
		increaseInstructionPointer = c.jnz(operand)
	case 4:
		c.bxc(operand)
	case 5:
		c.out(operand)
	case 6:
		c.bdv(operand)
	case 7:
		c.cdv(operand)
	default:
		panic(fmt.Sprintf("Invalid instruction opcode: %d", opcode))
	}

	if increaseInstructionPointer {
		c.instructionPointer += 2
	}
}

func (c *computer) runProgram() {
	for c.instructionPointer < len(c.program) {
		c.runInstruction()
	}
	outputItems := lo.Map(c.outputBuffer, func(item int, _ int) string {
		return fmt.Sprintf("%d", item)
	})
	println(strings.Join(outputItems, ","))
}

func (c *computer) checkIfProgramPrintsSelf() bool {
	checkedOutputTo := 0
	for c.instructionPointer < len(c.program) {
		c.runInstruction()
		for i := checkedOutputTo; i < len(c.outputBuffer); i++ {
			if c.outputBuffer[i] != c.program[i] {
				return false
			}
			checkedOutputTo = i
		}
	}
	if len(c.outputBuffer) != len(c.program) {
		return false
	}
	return true
}

type opcodeAndOperand struct {
	opcode  int
	operand int
}

type registerContentType int

const (
	literal registerContentType = iota
	literalModEight
	unknown
)

func (c *computer) findRegisterAThatPrintsProgram() int {
	opcodesAndOperands := make([]opcodeAndOperand, len(c.program)/2)
	for i := range opcodesAndOperands {
		opcodesAndOperands[i] = opcodeAndOperand{
			opcode:  c.program[i*2],
			operand: c.program[i*2+1],
		}
	}

	count := lo.CountBy(opcodesAndOperands, func(oao opcodeAndOperand) bool {
		return oao.opcode == 0
	})
	if count != 1 {
		panic("Expected 1 adiv")
	}
	adivOpAndOperand, _, found := lo.FindIndexOf(opcodesAndOperands, func(oao opcodeAndOperand) bool {
		return oao.opcode == 0
	})
	if !found {
		panic("Must find adiv op")
	}
	if adivOpAndOperand.operand > 3 {
		panic("Don't expect register operands")
	}
	adDivDenominator := intPow(2, adivOpAndOperand.operand)

	reversedOpcodesAndOperands := lo.Reverse(opcodesAndOperands)

	if reversedOpcodesAndOperands[0].opcode != 3 {
		panic(fmt.Sprintf("Invalid opcode, expected jnz as last op, got: %d", reversedOpcodesAndOperands[0].opcode))
	}
	if reversedOpcodesAndOperands[0].operand != 0 {
		panic(fmt.Sprintf(
			"Expect jnz to jump to instruction 0, but jumpes to %d, our logic doesn't handle that",
			reversedOpcodesAndOperands[0].operand,
		))
	}
	for _, oao := range reversedOpcodesAndOperands[1:] {
		if oao.opcode == 3 {
			panic("Expected only jnz in program to be at end.")
		}
	}

	// Index 0 -> solutions for sub-problem with just last elem of program.
	// Index 1 -> solutions for sub-problem with last 2 elems of program.
	// ...
	// Index len(c.program) - 1 -> solution for whole problem.
	solutionsForSubproblems := make([][]int, len(c.program))
	var lastSolutions []int
	for i := range solutionsForSubproblems {
		var candidateSolutions []int
		if lastSolutions == nil {
			candidateSolutions = make([]int, 0)
			for i := 1; i < adDivDenominator; i++ {
				candidateSolutions = append(candidateSolutions, i)
			}
		} else {
			for _, lastSolution := range lastSolutions {
				// For every solution in our last set, we add a range of solutions:
				// - prevSolution * aDivDenominator at the bottom end.
				// - prevSolution * aDivDenominator + (aDivDenominator - 1) at the top end.
				// - and all numbers in between.
				//
				// An example of why we do so: if our previous solutions were 2 and 4, with a div of 5 then all the
				// combinations that could lead to 2 or 4 are.
				// 10, 11, 12, 13, 14, 20, 21, 22, 23, 24. Because
				// 10 / 5 = 2, 11 / 5 = 2, 12 / 5 = 2, 13 / 5 = 2, 14 / 5 = 2 and
				// 20 / 5 = 4, 21 / 5 = 4, 22 / 5 = 4, 23 / 5 = 4, 24 / 5 = 4.
				//
				// The calculations here are the same as the long form example, and derive all potential candidates
				// that would lead to the solutions the next previous solutions.
				base := lastSolution * adDivDenominator
				for i := range adDivDenominator {
					candidateSolutions = append(candidateSolutions, base+i)
				}
			}
		}
		actualSolutions := make([]int, 0)
		subProblemOutputIndex := i + 1
		subProblemOutput := c.program[len(c.program)-subProblemOutputIndex]
		for i := range candidateSolutions {
			c.outputBuffer = make([]int, 0)
			c.registerA = candidateSolutions[i]
			c.registerB = 0
			c.registerC = 0
			c.instructionPointer = 0
			for len(c.outputBuffer) == 0 && c.instructionPointer < len(c.program)-2 /* stop before last jump */ {
				c.runInstruction()
			}
			if len(c.outputBuffer) > 0 && c.outputBuffer[0] == subProblemOutput {
				actualSolutions = append(actualSolutions, candidateSolutions[i])
			}
		}
		solutionsForSubproblems[i] = actualSolutions
		lastSolutions = actualSolutions
	}
	return lo.Min(solutionsForSubproblems[len(solutionsForSubproblems)-1])
}

func newComputer(registerA int, registerB int, registerC int, program []int) computer {
	return computer{
		registerA:          registerA,
		registerB:          registerB,
		registerC:          registerC,
		program:            program,
		instructionPointer: 0,
		outputBuffer:       make([]int, 0),
	}
}

func main() {
	scanner := bufio.NewScanner(strings.NewReader(input))

	handlingRegisters := true
	registerCount := 0
	var registerA int
	var registerB int
	var registerC int
	var program []int
	for scanner.Scan() {
		if scanner.Text() == "" {
			handlingRegisters = false
			continue
		}
		if handlingRegisters {
			_, registerValue := handleRegisterLine(scanner.Text())
			if registerCount == 0 {
				registerA = registerValue
			} else if registerCount == 1 {
				registerB = registerValue
			} else if registerCount == 2 {
				registerC = registerValue
			} else {
				panic("Invalid register count")
			}
			registerCount += 1
			continue
		}
		program = handleProgramLine(scanner.Text())
	}

	// Part 1.
	comp := newComputer(registerA, registerB, registerC, program)
	comp.runProgram()

	// Part 2.
	comp = newComputer(registerA, registerB, registerC, program)
	println(comp.findRegisterAThatPrintsProgram())
}
