package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	lr := newLogReader("BlackMed.log", "^(\\d\\d\\d\\d-\\d\\d-\\d\\d \\d\\d:\\d\\d:\\d\\d,\\d\\d\\d) \\[([A-Za-z]+)\\] ([^ ]+)")
	for true {
		if le := lr.next(); le == nil {
			return
		} else if le.class == "ERROR" {
			fmt.Print(le.line)
		}
	}
}

type LogEntry struct {
	line      string
	timeStamp string
	class     string
	level     string
}

type LogReader struct {
	path      string
	re        *regexp.Regexp
	lines     []string
	current   int
	nextEntry LogEntry
}

func newLogReader(path string, pattern string) LogReader {
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}
	lr := LogReader{
		path:    path,
		re:      re,
		lines:   readFile(path),
		current: 0,
	}
	for i, line := range lr.lines {
		if match := re.FindStringSubmatch(line); match != nil {
			lr.current = i
			lr.nextEntry = LogEntry{
				timeStamp: match[1],
				class:     match[2],
				level:     match[3],
			}
			break
		}
	}
	return lr
}

func (lr *LogReader) next() *LogEntry {
	if lr.current == len(lr.lines) {
		return nil
	}

	le := lr.nextEntry

	var sb strings.Builder
	sb.WriteString(lr.lines[lr.current])
	sb.WriteRune('\n')

	for i := lr.current + 1; i < len(lr.lines); i++ {
		if match := lr.re.FindStringSubmatch(lr.lines[i]); match != nil {
			lr.current = i
			lr.nextEntry = LogEntry{
				timeStamp: match[1],
				class:     match[2],
				level:     match[3],
			}
			break
		} else {
			sb.WriteString(lr.lines[i])
			sb.WriteRune('\n')
			if i == len(lr.lines)-1 {
				lr.current = len(lr.lines)
			}
		}
	}
	le.line = sb.String()
	return &le
}

func readFile(path string) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return strings.Split(string(b), "\n")
}
