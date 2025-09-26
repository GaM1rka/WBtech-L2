package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type options struct {
	col         int    // -k N (1-based)
	numeric     bool   // -n
	reverse     bool   // -r
	unique      bool   // -u
	month       bool   // -M
	ignoreTrail bool   // -b
	checkOnly   bool   // -c
	human       bool   // -h
	sep         string // --sep (default: \t)
	wordSplit   bool   // --ws (split by any whitespace)
	inputPath   string // optional file
}

var monthOrder = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

var humanRe = regexp.MustCompile(`^\s*([+-]?\d+(?:\.\d+)?)([KkMmGgTtPpEeZzYy]?[i]?[Bb]?)?\s*$`)

// parseHuman parses sizes like 10K, 2M, 1G, 3.5MiB, etc.
// Supports SI (K, M, G, ...) approximated as 1000^n and IEC (KiB, MiB, ...) as 1024^n.
// Returns float64 for stable comparisons with -n/-h combos.
func parseHuman(s string) (float64, error) {
	m := humanRe.FindStringSubmatch(s)
	if m == nil {
		return 0, errors.New("not human number")
	}
	numStr := m[1]
	unit := strings.ToUpper(strings.TrimSpace(m[2]))
	v, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}
	if unit == "" {
		return v, nil
	}
	// Normalize common suffixes
	si := map[string]float64{
		"K": 1e3, "M": 1e6, "G": 1e9, "T": 1e12, "P": 1e15, "E": 1e18, "Z": 1e21, "Y": 1e24,
	}
	iec := map[string]float64{
		"KIB": 1024, "MIB": 1024 * 1024, "GIB": 1024 * 1024 * 1024,
		"TIB": 1024 * 1024 * 1024 * 1024,
		"PIB": 1024 * 1024 * 1024 * 1024 * 1024,
		"EIB": 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
		"ZIB": 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
		"YIB": 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
	}
	if factor, ok := si[unit]; ok {
		return v * factor, nil
	}
	if factor, ok := iec[unit]; ok {
		return v * factor, nil
	}
	// accept trailing 'B' variants: KB, MB, GB, etc. as SI
	if strings.HasSuffix(unit, "B") && len(unit) == 2 {
		if factor, ok := si[unit[:1]]; ok {
			return v * factor, nil
		}
	}
	return v, nil
}

func parseArgs() (options, error) {
	opts := options{
		col:       1,
		sep:       "\t",
		wordSplit: false,
	}
	args := os.Args[1:]
	i := 0
	for i < len(args) {
		a := args[i]
		switch a {
		case "-k":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("missing value for -k")
			}
			val := args[i+1]
			i += 2
			n, err := strconv.Atoi(val)
			if err != nil || n < 1 {
				return opts, fmt.Errorf("invalid -k value: %s", val)
			}
			opts.col = n
		case "-n":
			opts.numeric = true
			i++
		case "-r":
			opts.reverse = true
			i++
		case "-u":
			opts.unique = true
			i++
		case "-M":
			opts.month = true
			i++
		case "-b":
			opts.ignoreTrail = true
			i++
		case "-c":
			opts.checkOnly = true
			i++
		case "-h":
			opts.human = true
			i++
		case "--sep":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("missing value for --sep")
			}
			opts.sep = args[i+1]
			i += 2
		case "--ws":
			opts.wordSplit = true
			i++
		case "-":
			// explicit stdin
			i++
		default:
			// non-flag => assume it is input file path
			if strings.HasPrefix(a, "-") {
				return opts, fmt.Errorf("unknown flag: %s", a)
			}
			opts.inputPath = a
			i++
		}
	}
	// sanity: -M and -n/-h are mutually exclusive in semantics; allow but month wins unless numbers forced
	// For this implementation, precedence:
	// -h > -n > -M for key parsing modes. Documented behavior for this exercise.
	return opts, nil
}

func readAllLines(path string) ([]string, error) {
	var r io.Reader
	if path == "" || path == "-" {
		r = bufio.NewReader(os.Stdin)
	} else {
		f, err := os.Open(filepath.Clean(path))
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = bufio.NewReader(f)
	}
	sc := bufio.NewScanner(r)
	// Increase buffer for very long lines (64K default)
	const meg = 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 64*meg)
	lines := make([]string, 0, 4096)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func splitFields(s string, sep string, wordSplit bool) []string {
	if wordSplit {
		return strings.Fields(s)
	}
	if sep == "" {
		// default to whitespace when empty sep accidentally set
		return strings.Fields(s)
	}
	return strings.Split(s, sep)
}

func keyAt(line string, col int, opts options) (raw string) {
	if opts.ignoreTrail {
		line = strings.TrimRight(line, " \t")
	}
	fields := splitFields(line, opts.sep, opts.wordSplit)
	idx := col - 1
	if idx >= 0 && idx < len(fields) {
		return fields[idx]
	}
	return "" // missing field evaluates as empty
}

func monthKey(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	k := strings.ToLower(s[:min(3, len(s))])
	m, ok := monthOrder[k]
	return m, ok
}

func numericKey(s string) (float64, error) {
	// try float
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

func compare(a, b string, opts options) int {
	ka := keyAt(a, opts.col, opts)
	kb := keyAt(b, opts.col, opts)

	// choose mode: -h > -n > -M > lexicographic
	if opts.human {
		va, ea := parseHuman(ka)
		vb, eb := parseHuman(kb)
		if ea == nil && eb == nil {
			switch {
			case va < vb:
				return -1
			case va > vb:
				return 1
			}
		} else if ea == nil && eb != nil {
			return -1 // parsed numbers come before non-numbers
		} else if ea != nil && eb == nil {
			return 1
		}
		// fallback lexicographic if both fail
	}
	if opts.numeric && !opts.human {
		va, ea := numericKey(ka)
		vb, eb := numericKey(kb)
		if ea == nil && eb == nil {
			switch {
			case va < vb:
				return -1
			case va > vb:
				return 1
			}
		} else if ea == nil && eb != nil {
			return -1
		} else if ea != nil && eb == nil {
			return 1
		}
		// fallback lexicographic
	}
	if opts.month && !opts.numeric && !opts.human {
		ma, oka := monthKey(ka)
		mb, okb := monthKey(kb)
		if oka && okb {
			switch {
			case ma < mb:
				return -1
			case ma > mb:
				return 1
			}
		} else if oka && !okb {
			return -1
		} else if !oka && okb {
			return 1
		}
		// fallback lexicographic
	}
	// lexicographic fallback
	if ka < kb {
		return -1
	} else if ka > kb {
		return 1
	}
	// equal keys => tie-breaker: full line lexicographic to make sort stable-ish
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}

func isSorted(lines []string, opts options) bool {
	for i := 1; i < len(lines); i++ {
		cmp := compare(lines[i-1], lines[i], opts)
		if opts.reverse {
			cmp = -cmp
		}
		if cmp > 0 {
			return false
		}
	}
	return true
}

func uniqInPlace(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}
	out := lines[:0]
	prev := ""
	first := true
	for _, s := range lines {
		if first || s != prev {
			out = append(out, s)
			prev = s
			first = false
		}
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	opts, err := parseArgs()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(2)
	}

	lines, err := readAllLines(opts.inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
		os.Exit(1)
	}

	less := func(i, j int) bool {
		cmp := compare(lines[i], lines[j], opts)
		if opts.reverse {
			return cmp > 0
		}
		return cmp < 0
	}

	// Sort unless only checking
	if !opts.checkOnly {
		sort.SliceStable(lines, less)
		if opts.unique {
			lines = uniqInPlace(lines)
		}
		// Output
		w := bufio.NewWriter(os.Stdout)
		for _, s := range lines {
			_, _ = w.WriteString(s)
			_, _ = w.WriteString("\n")
		}
		w.Flush()
		return
	}

	// -c: check sortedness
	if isSorted(lines, opts) {
		// GNU sort prints nothing if sorted; here — silence is OK
		return
	}
	// Print first disorder like GNU sort's "disorder at line N"
	for i := 1; i < len(lines); i++ {
		cmp := compare(lines[i-1], lines[i], opts)
		if opts.reverse {
			cmp = -cmp
		}
		if cmp > 0 {
			fmt.Fprintf(os.Stderr, "sort: disorder at line %d\n", i+1)
			os.Exit(1)
		}
	}
}
