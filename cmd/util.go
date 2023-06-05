package main

import (
	"sort"
	"strconv"
	"strings"
)

func atoi(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
}

func atoi64(s string) int64 {
	i, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 32)
	return i
}

func atof(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func Unique(col []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range col {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	sort.Strings(list)
	return list
}
