package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const buffer = 500
const workers = 100

var ignore []string

// ping is an empty struct to send through channels
// as a notification that something finished.
type ping struct{}

var root string
var extensions []string
var pattern *regexp.Regexp
var insensitive = false

var filenames = make(chan string, buffer)
var output = make(chan []string)
var done = make(chan ping)

func init() {
	// Read rc file if available.
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	rc, err := ioutil.ReadFile(path.Join(cwd, ".suchenrc"))
	// Default values for ignore.
	ignore = []string{"__pycache__"}
	if err != nil {
		return
	}
	for _, bad := range strings.Split(string(rc), "\n") {
		if bad != "" {
			ignore = append(ignore, bad)
		}
	}
}

// search is a filepath.WalkFunc suitable for passing to
// filepath.Walk which passes the filenames found into a channel.
func search(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return nil
	}
	if len(extensions) > 0 {
		for _, ext := range extensions {
			if filepath.Ext(path) == ext {
				filenames <- path
				return nil
			}
		}
		return nil
	}
	filenames <- path
	return nil
}

// init parses the command-line arguments into the values
// used to execute. There should be a pattern at the very
// least. Optionally, a path (defaulting to "."), and
// file extensions to search may be provided.
func init() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatal("No arguments passed.")
	}
	args = getExts(args)
	args = getRoot(args)
	args = getCaseStr(args)
	if len(args) != 1 {
		log.Fatal("Unable to find pattern.")
	}
	pat := args[0]
	if insensitive {
		pat = strings.ToLower(pat)
	}
	p, err := regexp.Compile(pat)
	if err != nil {
		log.Fatalf("Unable to compile pattern %q: %q\n", pat, err)
	}
	pattern = p
}

// getCaseStr accepts command-line flags and strips out
// the -i flag (if it exists). It returns a boolean for whether
// the regex should be case-insensitive and the args.
func getCaseStr(args []string) []string {
	var unused []string
	for _, val := range args {
		if val == "-i" {
			insensitive = true
		} else {
			unused = append(unused, val)
		}
	}
	return unused
}

// getExts sets the extensions global variable,
// removes any extension arguments from args,
// and returns args for further processing.
func getExts(args []string) []string {
	var unused []string
	for _, val := range args {
		if strings.HasPrefix(val, "--") {
			if len(val) < 3 {
				log.Fatalf("Invalid extension: '%s'\n", val)
			}
			extensions = append(extensions, "."+val[2:])
		} else {
			unused = append(unused, val)
		}
	}
	return unused
}

// getRoot finds a valid directory in the command-line
// args, sets it to the global "root" variable, and
// returns the remaining arguments.
func getRoot(args []string) []string {
	var unused []string
	for _, val := range args {
		if isDir(val) {
			if root != "" {
				log.Fatalf("Too many directory arguments\n")
			} else {
				root = val
			}
		} else {
			unused = append(unused, val)
		}
	}
	if root == "" {
		root = "."
	}
	return unused
}

func main() {
	for i := 0; i < workers; i++ {
		go checkFile()
	}

	go func() {
		filepath.Walk(root, search)
		close(filenames)
	}()

	go func() {
		for i := 0; i < workers; i++ {
			<-done
		}
		close(output)
	}()

	for lines := range output {
	RESULTS:
		for _, line := range lines {
			for _, ugly := range ignore {
				if strings.Contains(line, ugly) {
					break RESULTS
				}
			}
			fmt.Println(line)
		}
	}
	os.Stdout.Sync()
}

// checkFile takes a filename and reads the file to determine
// whether the file contains the regex in the global pattern.
func checkFile() {
	for filename := range filenames {
		file, err := os.Open(filename)
		defer file.Close()
		if err != nil {
			log.Fatal(err)
			return
		}
		scanner := bufio.NewScanner(file)
		var fileType string
		line := 0
		var lines []string
		for scanner.Scan() {
			line++
			orig := scanner.Text()
			txt := orig
			if line == 1 {
				fileType = http.DetectContentType(scanner.Bytes())
			}
			if insensitive {
				txt = strings.ToLower(txt)
			}
			found := pattern.FindIndex([]byte(txt))
			if found != nil {
				if fileType[:4] != "text" {
					break
				}
				lines = append(lines, fmt.Sprintf("%s:%d:%s", filename, line, orig))
			}
		}
		output <- lines
		file.Close()
	}
	done <- ping{}
}

func isDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}
