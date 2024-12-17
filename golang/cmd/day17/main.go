package main

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"github.com/samber/lo"
	"github.com/sourcegraph/conc/pool"
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
	ctx, cancel := context.WithCancel(context.Background())
	concurrencyPool := pool.New().WithContext(ctx)

	for i := range 4_000_000_000_000 {
		concurrencyPool.Go(func(ctx context.Context) error {
			comp = newComputer(i, registerB, registerC, program)
			if comp.checkIfProgramPrintsSelf() {
				println(i)
				cancel()
			}
			return nil
		})

	}
	err := concurrencyPool.Wait()
	if err != nil {
		panic(err)
	}
}
