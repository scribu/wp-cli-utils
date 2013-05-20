// Utility for concurrently compiling all the man pages

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
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

func generateMan(part string, done chan bool) {
	defer func() { done <- true }()

	cmd := exec.Command(WP_CLI_PATH+"/bin/wp", "help", "--gen", part)

	out, _ := cmd.CombinedOutput()

	log.Println(part)
	log.Println(strings.Trim(string(out), "\n"))
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: man.go /path/to/wp-cli")
	}

	WP_CLI_PATH = os.Args[1]

	done := make(chan bool)

	wp := decodeJSON(getCommandsAsJSON())

	for _, cmd := range wp.Subcommands {
		go generateMan(cmd.Name, done)
	}

	for i := len(wp.Subcommands); i > 0; i-- {
		<-done
	}
}
