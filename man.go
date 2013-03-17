// Utility for concurently compiling all the man pages

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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

func handleCommand(cmd Command, path string) {
	full_path := path + " " + cmd.Name

	if len(cmd.Subcommands) == 0 {
		fmt.Println(full_path + " " + cmd.Synopsis)
	} else {
		for _, subcmd := range cmd.Subcommands {
			handleCommand(subcmd, full_path)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("need to pass the path to the WP-CLI directory")
	}

	WP_CLI_PATH = os.Args[1]

	handleCommand(decodeJSON(getCommandsAsJSON()), "")
}
