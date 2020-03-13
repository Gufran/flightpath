// +build docs

package main

import (
	"flag"
	"fmt"
)

const (
	tplFlagName    = "==`-%s`==\n"
	tplFlagDefault = ":    Default `%q`\n"
	tplFlagUsage   = "     %s\n"
	header         = `# Usage

Following command line flags can be used to configure flightpath

`
)

func main() {
	fmt.Print(header)
	flag.CommandLine.VisitAll(printUsage)
}

func printUsage(f *flag.Flag) {
	fmt.Printf(tplFlagName, f.Name)
	fmt.Println()
	fmt.Printf(tplFlagDefault, f.DefValue)
	fmt.Println()
	fmt.Printf(tplFlagUsage, f.Usage)
	fmt.Println()
}
