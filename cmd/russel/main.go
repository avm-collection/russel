package main

import (
	"os"
	"fmt"
	"flag"
	"path/filepath"
	"strings"

	"github.com/LordOfTrident/russel/internal/errors"
	"github.com/LordOfTrident/russel/internal/config"
	"github.com/LordOfTrident/russel/internal/token"
	"github.com/LordOfTrident/russel/internal/compiler"
)

var (
	out  = flag.String("o",     "",    "Path of the output binary")
	v    = flag.Bool("version", false, "Show the version")
	maxE = flag.Int("maxE",     8,     "Max amount of compiler errors")

	args []string
)

func shiftArgs() (string, bool) {
	if len(args) == 0 {
		return "", false
	}

	arg := args[0]
	args = args[1:]

	return arg, true
}

func parseArgs() {
	flag.Parse()

	args = flag.Args()
	for i := 0; i < len(flag.Args()); i ++ {
		if len(flag.Args()[i]) == 0 {
			continue
		}

		if flag.Args()[i][0] != '-' {
			continue
		}

		args = flag.Args()[:i]
		flag.CommandLine.Parse(flag.Args()[i:])

		break
	}
}

func printError(format string, args... interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", fmt.Sprintf(format, args...))
}

func printTry(arg string) {
	fmt.Fprintf(os.Stderr, "Try '%v %v'\n", os.Args[0], arg)
}

func usage() {
	fmt.Printf("%v v%v.%v.%v\n\n", config.AsciiLogo,
	           config.VersionMajor, config.VersionMinor, config.VersionPatch)
	fmt.Printf("Github: %v\n", config.GithubLink)
	fmt.Printf("Usage: %v [build [FILE] | run FILE] [OPTIONS]\n", os.Args[0])
	fmt.Println("Options:")
	fmt.Println("  -help\n        Show this message")
	fmt.Println("  -h    Alias for -help")
	flag.PrintDefaults()
}

func version() {
	fmt.Printf("%v %v.%v.%v\n", config.AppName,
	           config.VersionMajor, config.VersionMinor, config.VersionPatch)
}

func run() {
	panic("'run' mode not implemented yet")
}

func build() {
	path, ok := shiftArgs()
	if !ok {
		panic("Build system mode not implemented yet")
	}

	if len(*out) == 0 {
		if len(filepath.Ext(path)) == 0 {
			*out = path + ".out"
		} else {
			*out = strings.TrimSuffix(path, filepath.Ext(path)) + ".anasm"
		}

		*out = filepath.Base(*out)
	} else {
		*out += ".anasm"
	}

	if len(args) > 0 {
		printError("Unexpected argument '%v'", args[0])
		printTry("-h")

		os.Exit(1)
	}

	compile(path, *out)
}

func compile(path, out string) {
	data, err := os.ReadFile(path)
	if err != nil {
		printError("Could not open file '%v'", path)
		printTry("-h")

		os.Exit(1)
	}

	c := compiler.New(string(data), path)

	err = c.CompileInto(out)
	if err != nil {
		printError(err.Error())

		os.Exit(1)
	}
}

func init() {
	token.AllTokensCoveredTest()

	flag.Usage = usage

	// Aliases
	flag.BoolVar(v, "v", *v, "Alias for -version")

	parseArgs()

	errors.Max = *maxE
}

func main() {
	if *v {
		version()

		return
	}

	mode, ok := shiftArgs()
	if !ok {
		printError("No mode specified")
		printTry("-h")

		os.Exit(1)
	}

	switch mode {
	case "run":   run()
	case "build": build()

	default:
		printError("Unknown mode '%v'", mode)
		printTry("-h")

		os.Exit(1)
	}
}
