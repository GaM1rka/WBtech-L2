package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	fieldsStr := flag.String("f", "", "the number of colums to be printed")
	delimiter := flag.String("d", "\t", "delimiter to use")
	separated := flag.Bool("s", false, "to use the strings that only contain delimiter")

	flag.Parse()

	if *fieldsStr == "" {
		log.Fatal("flag -f is required")
	}

	fields, err := parseFields(*fieldsStr)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		result, ok := processLine(line, fields, *delimiter, *separated)
		if !ok {
			continue
		}

		fmt.Println(result)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
