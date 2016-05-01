package main

import "strings"

func joinCommandLine(runArgs []string) string {
	formatted := []string{}
	for _, ra := range runArgs {
		trimmed := strings.Trim(ra, " '\"")
		if strings.Contains(trimmed, " ") {
			formatted = append(formatted, "\""+trimmed+"\"")
		} else {
			formatted = append(formatted, trimmed)
		}
	}
	return strings.Join(formatted, " ")
}
