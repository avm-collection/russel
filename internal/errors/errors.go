package errors

import (
	"fmt"
	"os"
	"strings"

	"github.com/LordOfTrident/russel/internal/token"
)

var (
	Max   = 8
	count = 0

	first = true
)

func template(where token.Where, msg, type_, attr string) {
	idx   := where.Col - 1
	line  := fmt.Sprintf("%v", where.Row)
	part1 := strings.Replace(where.Line[:idx],                "\t", "    ", -1)
	main  := strings.Replace(where.Line[idx:idx + where.Len], "\t", "    ", -1)
	part2 := strings.Replace(where.Line[idx + where.Len:],    "\t", "    ", -1)

	fmt.Fprintf(os.Stderr, "%v:\x1b[0;1m %v\x1b[0m: %v\n",   attr + type_, where, msg)
	fmt.Fprintf(os.Stderr, "    %v | %v%v\x1b[0m%v\n", line, part1, attr + main, part2)
}

func NoOutput() bool {
	return first
}

func separator() {
	if !first {
		fmt.Println()
	} else {
		first = false
	}
}

func newError() {
	count ++

	if count > Max {
		fmt.Fprintf(os.Stderr, "...\nCompilation aborted\n")
		os.Exit(1)
	}
}

func Error(where token.Where, format string, any... interface{}) {
	newError()
	separator()
	template(where, fmt.Sprintf(format, any...), "Error", "\x1b[1;91m")
}

func Simple(format string, any... interface{}) {
	newError()
	separator()
	fmt.Fprintf(os.Stderr, "\x1b[1;91mError:\x1b[0m %v\n", fmt.Sprintf(format, any...))
}

func Warning(where token.Where, format string, any... interface{}) {
	separator()
	template(where, fmt.Sprintf(format, any...), "Warning", "\x1b[1;93m")
}

func Note(where token.Where, format string, any... interface{}) {
	separator()
	template(where, fmt.Sprintf(format, any...), "Note", "\x1b[1;96m")
}

func Happend() bool {
	return count > 0
}

func Reset() {
	first = true
	count = 0
}
