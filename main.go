package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
)

func main() {
	var search string
	var class string
	var level string
	var lines bool

	flag.StringVar(&search, "search", "", "regular expression to match")
	flag.StringVar(&class, "class", "", "a class to match (literal match)")
	flag.StringVar(&level, "level", "", "a level to match (literal match)")
	flag.BoolVar(&lines, "lines", false, "print a line to seperate messages")

	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Fprint(os.Stderr, "Error: no file argument given\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.CommandLine.PrintDefaults()
		return
	}

	var searchRe *regexp.Regexp
	if search != "" {
		var searchReErr error
		searchRe, searchReErr = regexp.Compile(search)
		if searchReErr != nil {
			panic(searchReErr)
		}
	}

	re, err := regexp.Compile(`^(\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d,\d\d\d) \[([A-Za-z]+)\] ([^ ]+)`)
	if err != nil {
		panic(err)
	}

	for _, file := range flag.Args() {
		for le := range LogLines(file, re) {
			if class == "" || le.class == class {
				if level == "" || le.level == level {
					if searchRe == nil || searchRe.MatchString(le.line) {
						fmt.Print(le.line)
						if lines {
							fmt.Println("--------------------------------------------------------------------------------")
						}
					}
				}
			}
		}
	}
}

type LogEntry struct {
	line      string
	timeStamp string
	class     string
	level     string
}

func LogLines(path string, re *regexp.Regexp) func(func(LogEntry) bool) {
	lines := readFile(path)
	var nextEntry LogEntry
	var sb strings.Builder

	start, logEntry := findFirstEntry(lines, re)

	return func(yield func(LogEntry) bool) {
		for i := start + 1; i < len(lines); i++ {
			if match := re.FindStringSubmatch(lines[i]); match != nil {
				nextEntry = getEntry(match)

				logEntry.line = sb.String()
				if !yield(logEntry) {
					return
				}
				logEntry = nextEntry
				sb.Reset()
			}
			sb.WriteString(lines[i])
			sb.WriteRune('\n')
		}
		logEntry.line = sb.String()
		yield(logEntry)
	}
}

func readFile(path string) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return strings.Split(string(b), "\n")
}

func findFirstEntry(lines []string, re *regexp.Regexp) (start int, logEntry LogEntry) {
	for i, line := range lines {
		if match := re.FindStringSubmatch(line); match != nil {
			entry := getEntry(match)
			return i, entry
		}
	}
	return math.MaxInt - 1, LogEntry{}
}

func getEntry(match []string) LogEntry {
	return LogEntry{
		timeStamp: match[1],
		level:     match[2],
		class:     match[3],
	}
}
