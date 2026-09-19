package main

import (
	"fmt"
	"os"
	"strings"
)

// Usage: go run . <farm.txt>
// Prints: original file, blank line, then one "L<ant>-<room> ..." line per turn.
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "ERROR: invalid data format, no input file specified")
		os.Exit(1)
	}
	colony, err := ParseFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	turns, err := Solve(colony)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var sb strings.Builder
	for _, line := range colony.RawInput {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	for _, mv := range turns {
		parts := make([]string, len(mv))
		for i, m := range mv {
			parts[i] = fmt.Sprintf("L%d-%s", m.Ant, m.Room)
		}
		sb.WriteString(strings.Join(parts, " "))
		sb.WriteString("\n")
	}
	fmt.Print(sb.String())
}
