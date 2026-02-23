package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func Run(pattern, fileName string, a, b, c int, count, ignoreCase, invert, fixed, lineNum bool) {
	a = max(a, c)
	b = max(b, c)

	r, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("ERROR opening %s: %v\n", fileName, err)
		return
	}
	defer r.Close()

	scanner := bufio.NewScanner(r)
	lines := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	matchString := 0
	lineMap := make(map[int]bool)
	if ignoreCase {
		pattern = strings.ToLower(pattern)
	}

	for idx := 0; idx < len(lines); idx++ {
		line := lines[idx]
		if ignoreCase {
			line = strings.ToLower(lines[idx])
		}
		if matches(line, pattern, fixed, invert) {
			if count {
				matchString++
				continue
			}
			if a >= 0 && b >= 0 {
				start := max(idx-b, 0)
				end := min(idx+a, len(lines)-1)
				for i := start; i <= end; i++ {
					if !lineMap[i] {
						if lineNum {
							fmt.Printf("%d:%s\n", i+1, lines[i])
						} else {
							fmt.Println(lines[i])
						}
					}
					lineMap[i] = true
				}
			}
		}
	}

	if count {
		fmt.Println(matchString)
	}
}

func matches(line, pattern string, fixed, invert bool) bool {
	if fixed && !invert {
		if strings.Contains(line, pattern) {
			return true
		}
		return false
	}

	re := regexp.MustCompile(pattern)
	if invert {
		return !re.MatchString(line)
	}
	return re.MatchString(line)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
