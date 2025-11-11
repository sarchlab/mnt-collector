package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseTraceFile parses a .traceg file and returns a map opcode->count.
// It expects lines for instructions in the format used in provided sample files.
func ParseTraceFile(path string) (map[string]int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	opcounts := make(map[string]int64)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// skip header lines (start with '-' or '#', or words like "thread block", "warp", "insts")
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "thread block") || strings.HasPrefix(line, "warp") || strings.HasPrefix(line, "insts") {
			continue
		}

		// Instruction lines have fields, first is PC hex, second is mask, third is dest_num (int)
		elems := strings.Fields(line)
		// quick sanity check
		if len(elems) < 4 {
			continue
		}
		// parse dest_num at elems[2]
		destNum, err := strconv.Atoi(elems[2])
		if err != nil {
			// not an instruction line
			continue
		}
		opIndex := 3 + destNum
		if opIndex >= len(elems) {
			continue
		}
		opcode := elems[opIndex]
		// strip possible trailing punctuation
		opcode = strings.TrimSpace(opcode)
		// In sample opcodes can be like "LDG.E.CONSTANT" or "LDC.64". We keep them as-is.
		opcounts[opcode]++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}
	return opcounts, nil
}
