package main

import (
	"bytes"
	"path"
	"path/filepath"
	"log"
	"regexp"
	"os"
)

var WP_CLI_PATH string

func readFileIntoBuffer(fname string) (buf bytes.Buffer, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return
	}

	buf.ReadFrom(f)
	f.Close()

	return
}

func findInvocationsInFile(fname string, re *regexp.Regexp, found chan []string) {
	buf, err := readFileIntoBuffer(fname)
	if err != nil { panic(err) }

	matchesRaw := re.FindAllStringSubmatch(buf.String(), -1)

	var matches []string

	if matchesRaw == nil {
		matches = make([]string, 0)
	} else {
		matches = make([]string, len(matchesRaw))
		for i, match := range matchesRaw {
			log.Println(match[1])
			matches[i] = match[1]
		}
	}

	found <- matches
}

func findInvocations() map[string]int {
	found := make(chan []string)

	files, err := filepath.Glob(path.Join(WP_CLI_PATH, "features/*.feature"))
	if err != nil { panic(err) }

	re := regexp.MustCompile("I run `wp (.+)`")

	for _, file := range files {
		go findInvocationsInFile(file, re, found)
	}

	invocations := make(map[string]int)

	for i := len(files); i > 0; i-- {
		matches := <-found

		for _, match := range matches {
			log.Println(match)
			invocations[match]++
		}
	}

	return invocations
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: coverage.go /path/to/wp-cli")
	}

	WP_CLI_PATH = os.Args[1]

	invocations := findInvocations()

	log.Println(invocations)
}
