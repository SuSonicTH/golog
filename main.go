package main

import (
	"bufio"
	"flag"
	"fmt"
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
			if !(class == "" || le.class == class) {
				continue
			}
			if !(level == "" || le.level == level) {
				continue
			}
			if !(searchRe == nil || searchRe.MatchString(le.line)) {
				continue
			}
			fmt.Print(le.line)
			if lines {
				fmt.Println("--------------------------------------------------------------------------------")
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
	return func(yield func(LogEntry) bool) {
		var nextEntry LogEntry
		var sb strings.Builder

		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		logEntry := findFirstEntry(scanner, re)
		for scanner.Scan() {
			line := scanner.Text()
			if match := re.FindStringSubmatch(line); match != nil {
				nextEntry = getEntry(match)
				logEntry.line = sb.String()
				if !yield(logEntry) {
					return
				}
				logEntry = nextEntry
				sb.Reset()
			}
			sb.WriteString(line)
			sb.WriteRune('\n')
		}
		logEntry.line = sb.String()
		yield(logEntry)
	}
}

func findFirstEntry(scanner *bufio.Scanner, re *regexp.Regexp) LogEntry {
	for scanner.Scan() {
		line := scanner.Text()
		if match := re.FindStringSubmatch(line); match != nil {
			entry := getEntry(match)
			return entry
		}
	}
	return LogEntry{}
}

func getEntry(match []string) LogEntry {
	return LogEntry{
		timeStamp: match[1],
		level:     match[2],
		class:     match[3],
	}
}
