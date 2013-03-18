// Utility for concurently compiling all the man pages

package main

import (
	"bytes"
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

func generateMan(src_path string, f os.FileInfo, err error) error {
	if !strings.HasSuffix(src_path, ".txt") {
		return nil
	}

	parts := convertPath(src_path)

	cmd := exec.Command(WP_CLI_PATH+"/bin/wp", append(parts, "--man")...)
	cmd.Dir = "/home/scribu/wp" // TODO

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		log.Fatal(stderr.String())
	} else {
		log.Println(stdout.String())
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("need to pass the path to the WP-CLI directory")
	}

	WP_CLI_PATH = os.Args[1]

	filepath.Walk(path.Join(WP_CLI_PATH, "man-src"), generateMan)
}
