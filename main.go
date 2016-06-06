package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const workers = 32

var wg = sync.WaitGroup{}

var root string
var extensions []string
var plain string
var pattern *regexp.Regexp
var insensitive bool
var regexSearch = true

var filenames = make(chan string)

// output takes []string instead of string so all lines from
// the same file are printed together, instead of interleaved with
// the output of other files
var output = make(chan []string)

// search is a filepath.WalkFunc suitable for passing to
// filepath.Walk which passes the filenames found into a channel.
func search(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Printf("error checking %q: %q\n", path, err)
		return nil
	}
	// ignore directories, etc.
	if !info.Mode().IsRegular() {
		return nil
	}

	// if extensions is set, only put this filename in the channel if its
	// extension is one of the ones we care about
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
	// get the extensions from args and put them into the global extensions
	// slice
	args = getExts(args)
	args = getCaseStr(args)
	args = getRoot(args)
	args = getRegexFlag(args)
	if len(args) != 1 {
		log.Fatal("Unable to find pattern.")
	}
	plain = args[0]
	if insensitive {
		plain = strings.ToLower(plain)
	}
	p, err := regexp.Compile(plain)
	if err != nil {
		log.Fatalf("Unable to compile pattern %q: %q\n", plain, err)
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

// getRegexFlag accepts command-line flags and strips out
// the -n flag (if it exists). It returns a boolean for whether
// a regex or plain-text search should be done.
func getRegexFlag(args []string) []string {
	var unused []string
	for _, val := range args {
		if val == "-n" {
			regexSearch = false
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
		wg.Add(1)
		go checkFile()
	}

	go func() {
		filepath.Walk(root, search)
		close(filenames)
	}()

	go func() {
		wg.Wait()
		close(output)
	}()

	for lines := range output {
		for _, line := range lines {
			fmt.Println(line)
		}
	}
	os.Stdout.Sync()
}

// checkFile takes a filename and reads the file to determine
// whether the file contains the regex in the global pattern.
func checkFile() {
	defer wg.Done()
	pat := pattern.Copy()
	for filename := range filenames {
		file, err := os.Open(filename)
		defer file.Close()
		if err != nil {
			log.Printf("error attempting to read %q: %q\n", filename, err)
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
				fileType = DetectContentType(scanner.Bytes())
				if fileType[:4] != "text" {
					break
				}
			}
			if insensitive {
				txt = strings.ToLower(txt)
			}
			if regexSearch {
				found := pat.FindIndex([]byte(txt))
				if found != nil {
					lines = append(lines, fmt.Sprintf("%s:%d:%s", filename, line, orig))
				}
			} else {
				if strings.Contains(txt, plain) {
					lines = append(lines, fmt.Sprintf("%s:%d:%s", filename, line, orig))
				}
			}
		}
		output <- lines
		file.Close()
	}
}

func isDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}
