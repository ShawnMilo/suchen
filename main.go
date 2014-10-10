package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

const buffer = 500

var root string
var extensions []string
var pattern *regexp.Regexp
var insensitive = false

// search creates a filepath.WalkFunc suitable for passing to
// filepath.Walk which passes the filenames found into a channel.
func search(path string, info os.FileInfo, err error) error {
    if !info.Mode().IsRegular() {
        return nil
    }
    if len(extensions) > 0 {
        for _, ext := range extensions {
            if filepath.Ext(path) == ext {
                checkFile(path)
                return nil
            }
        }
        return nil
    }
    checkFile(path)
    return nil
}

// init parses the command-line arguments into the values
// used to execute. There should be a pattern at the very
// least. Optionally, a path (defaulting to "."), and
// file extensions to search may be provided.
func init() {
    args := os.Args[1:]
    if len(args) == 0 {
        log.Fatalf("No arguments passed.")
    }
    args = getExts(args)
    args = getRoot(args)
    args = getCaseStr(args)
    if len(args) != 1 {
        log.Fatalf("Unable to find pattern.\n")
    }
    pat := args[0]
    if insensitive {
        pat = strings.ToLower(pat)
    }
    p, err := regexp.Compile(pat)
    if err != nil {
        log.Fatal(err)
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
        if IsDir(val) {
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
    filepath.Walk(root, search)
}

// checkFile takes a filename and reads the file to determine
// whether the file contains the regex in the global pattern.
func checkFile(filename string) {
    file, err := os.Open(filename)
    if err != nil {
        log.Fatal(err)
        return
    }
    defer file.Close()
    scanner := bufio.NewScanner(file)
    line := 0
    for scanner.Scan() {
        line += 1
        txt := scanner.Text()
        if insensitive {
            txt = strings.ToLower(txt)
        }
        found := pattern.FindIndex([]byte(txt))
        if found != nil {
            fmt.Printf("%s:%d:%s\n", filename, line, scanner.Text())
        }
    }
}

func IsDir(path string) bool {
    stat, err := os.Stat(path)
    if err != nil {
        return false
    }
    return stat.IsDir()
}
