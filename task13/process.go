package main

import (
	"fmt"
	"strconv"
	"strings"
)

func processLine(line string, fields []int, delimiter string, separated bool) (string, bool) {
	if separated && !strings.Contains(line, delimiter) {
		return "", false
	}

	parts := strings.Split(line, delimiter)
	result := make([]string, 0, len(fields))

	for _, field := range fields {
		index := field - 1
		if index < 0 || index > len(parts)-1 {
			continue
		}
		result = append(result, parts[index])
	}

	return strings.Join(result, delimiter), true
}

func parseFields(spec string) ([]int, error) {
	result := make([]int, 0)
	seen := make(map[int]bool)

	parts := strings.Split(spec, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("invalid fields format")
		}

		if strings.Contains(part, "-") {
			bounds := strings.Split(part, "-")
			if len(bounds) != 2 {
				return nil, fmt.Errorf("invalid range: %s", part)
			}

			from, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid range start: %s", part)
			}

			to, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid range end: %s", part)
			}

			if from <= 0 || to <= 0 {
				return nil, fmt.Errorf("fields must be positive: %s", part)
			}

			if from > to {
				return nil, fmt.Errorf("invalid range %s: start is greater than end", part)
			}

			for i := from; i <= to; i++ {
				if !seen[i] {
					result = append(result, i)
					seen[i] = true
				}
			}
		} else {
			field, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid field: %s", part)
			}

			if field <= 0 {
				return nil, fmt.Errorf("field numbers must be positive: %d", field)
			}

			if !seen[field] {
				result = append(result, field)
				seen[field] = true
			}
		}
	}

	return result, nil
}
