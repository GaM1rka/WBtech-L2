package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	AFlag := flag.Int("A", 0, "Print additional lines after the found line")
	BFlag := flag.Int("B", 0, "Print additional lines before the found line")
	CFlag := flag.Int("C", 0, "Print additional lines before and after the found line")
	cFlag := flag.Bool("c", false, "Print just number of matched liens")
	iFlag := flag.Bool("i", false, "Ignore the letter case")
	vFlag := flag.Bool("v", false, "Invert filter: print lines that do not match the expression")
	FFlag := flag.Bool("F", false, "Print exactly matched lines")
	nFlag := flag.Bool("n", false, "Print the line number before the line")

	flag.Parse()

	if len(flag.Args()) != 2 {
		fmt.Fprintf(os.Stderr, "Pattern and file name required!")
		os.Exit(1)
	}

	pattern := flag.Args()[0]
	file := flag.Args()[1]

	Run(pattern, file, *AFlag, *BFlag, *CFlag, *cFlag, *iFlag, *vFlag, *FFlag, *nFlag)
}
