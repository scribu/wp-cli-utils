// Utility for concurently compiling all the man pages

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

var WP_CLI_PATH string

type Command struct {
	Name, Synopsis, Description string
	Subcommands                 []Command
}

func callRonn(input string) string {
	cmd := exec.Command("ronn", "--date=2012-01-01 --roff --manual='WP-CLI'")

	cmd.Stdin = strings.NewReader(input)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	return out.String()
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

func convertPath(cmd_path string) string {
	cmd_path = strings.Replace(cmd_path, " ", "-", -1)
	cmd_path = strings.Replace(cmd_path, "-wp-", "", 1)

	return cmd_path
}

func getManSrc(cmd_path string) string {

	f, err := os.Open(path.Join(WP_CLI_PATH, "man-src", convertPath(cmd_path)+".txt"))
	if err != nil {
		return ""
	}

	var b bytes.Buffer
	b.ReadFrom(f)
	f.Close()

	return b.String()
}

func writeMan(contents, cmd_path string) {
	final_path := path.Join(WP_CLI_PATH, "man", convertPath(cmd_path)+".1")
	f, err := os.Create(final_path)
	if err != nil {
		log.Fatal("can't create ", final_path)
	}

	_, err = f.Write([]byte(contents))
	if err != nil {
		panic(err)
	}
}

func handleCommand(cmd Command, path string) {
	full_path := path + " " + cmd.Name

	manSrc := getManSrc(full_path)

	if "" != manSrc {
		roff := strings.Replace(callRonn(manSrc), " \"January 2012\"", "", -1)
		writeMan(roff, full_path)
		log.Println("generated" + full_path)
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
