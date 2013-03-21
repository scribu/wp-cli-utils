// Looks through wp-cli/features/ and figures out which subcommands aren't used.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"regexp"
)

var WP_CLI_PATH string

type Command struct {
	Name, Synopsis, Description string
	Subcommands                 []Command
}

func getCommandsAsJSON() bytes.Buffer {
	cmd := exec.Command(WP_CLI_PATH+"/bin/wp", "--cmd-dump")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	return out
}

func decodeJSON(out bytes.Buffer) Command {
	dec := json.NewDecoder(bytes.NewReader(out.Bytes()))

	var c Command

	for {
		if err := dec.Decode(&c); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
	}

	return c
}

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
			matches[i] = match[1]
		}
	}

	found <- matches
}

func findInvocations() *map[string]int {
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
			invocations[match]++
		}
	}

	return &invocations
}

func walkCommands(cmd Command, parents []string, invocations *map[string]int, notfound *[]string) {
	if len(cmd.Subcommands) > 0 {
		for _, subcmd := range cmd.Subcommands {
			walkCommands(subcmd, append(parents, cmd.Name), invocations, notfound)
		}

		return
	}

	path := strings.Join(append(parents, cmd.Name), " ")

	for key, _ := range *invocations {
		if 0 == strings.Index("wp " + key, path) {
			return
		}
	}

	*notfound = append(*notfound, path)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: coverage.go /path/to/wp-cli")
	}

	WP_CLI_PATH = os.Args[1]

	commands := decodeJSON(getCommandsAsJSON())

	notfound := make([]string, 0)

	walkCommands(commands, make([]string, 0), findInvocations(), &notfound)

	for _, cmd := range notfound {
		fmt.Println(cmd)
	}
}
