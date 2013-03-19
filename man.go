// Utility for concurently compiling all the man pages

package main

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var WP_CLI_PATH string

func convertPath(cmd_path string) []string {
	cmd_path = strings.Replace(path.Base(cmd_path), ".txt", "", 1)

	return strings.Split(cmd_path, "-")
}

func runCmd(parts []string) (bytes.Buffer, bytes.Buffer, error) {
	cmd := exec.Command(WP_CLI_PATH+"/bin/wp", append(parts, "--man")...)
	cmd.Dir = "/home/scribu/wp" // TODO

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return stdout, stderr, err
}

func unifyLastPart(parts []string) ([]string, error) {
	head := len(parts)-2
	if head < 0 {
		return parts, errors.New("Can't combine a single part")
	}

	return append(parts[:head], strings.Join(parts[head:], "-")), nil
}

func generateMan(src_path string, done chan bool) {
	defer func() { done <- true }()

	parts := convertPath(src_path)

	for {
		stdout, stderr, err := runCmd(parts)

		if err == nil {
			if "" == stdout.String() {
				log.Println(parts)
			} else {
				log.Println(strings.Trim(stdout.String(), "\n"))
			}
			break
		}

		parts, err = unifyLastPart(parts)
		if err != nil {
			log.Fatal(stderr.String())
			break
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("need to pass the path to the WP-CLI directory")
	}

	WP_CLI_PATH = os.Args[1]

	files, err := filepath.Glob(path.Join(WP_CLI_PATH, "man-src/*.txt"))
	if err != nil { panic(err) }

	done := make(chan bool)

	for _, file := range files {
		go generateMan(file, done)
	}

	for i := len(files); i>0; i-- {
		<-done
	}
}
