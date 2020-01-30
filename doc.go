// +build docs

package main

import (
	"flag"
	"fmt"
)

const (
	tableHeader = "| Option |  Description |"
	tableSep    = "|:--------|:------------|"
	tableRow    = "| `-%s` |  %s |\n"
)

func main() {
	fmt.Println(tableHeader)
	fmt.Println(tableSep)
	flag.CommandLine.VisitAll(printFlagTable)
}

func printFlagTable(f *flag.Flag) {
	fmt.Printf(tableRow, f.Name, f.Usage)
}
