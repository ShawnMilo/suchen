package main

import (
    "bufio"
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "regexp"
    "runtime"
    "strings"
)

const buffer = 500

var root string
var extensions []string
var pattern *regexp.Regexp
var cpus int

// getNames creates a filepath.WalkFunc suitable for passing to
// filepath.Walk which passes the filenames found into a channel.
func getNames(c chan string) filepath.WalkFunc {
    return func(path string, info os.FileInfo, err error) error {
        if !info.Mode().IsRegular() {
            return nil
        }
        if len(extensions) > 0 {
            for _, ext := range extensions {
                if filepath.Ext(path) == ext {
                    c <- path
                    return nil
                }
            }
            return nil
        }
        c <- path
        return nil
    }
}

// init parses the command-line arguments into the values
// used to execute. There should be a pattern at the very
// least. Optionally, a path (defaulting to "."), and
// file extensions to search may be provided.
func init() {
    flag.Parse()
    args := flag.Args()
    if len(args) == 0 {
        log.Fatalf("No arguments passed.")
    }
    args = getExts(args)
    args = getRoot(args)
    if len(args) != 1 {
        log.Fatalf("Unable to find pattern.\n")
    }
    p, err := regexp.Compile(args[0])
    if err != nil {
        log.Fatal(err)
    }
    pattern = p
    cpus = runtime.NumCPU()
    runtime.GOMAXPROCS(cpus)
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

// isDir returns true if name points to a directory
func isDir(name string) bool {
    file, err := os.Open(name)
    if err != nil {
        return false
    }
    defer file.Close()

    info, err := file.Stat()
    if err != nil {
        return false
    }

    return info.Mode().IsDir()
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
    filenames := make(chan string, buffer)
    toProcess := make(chan string, buffer)
    done := make(chan bool, buffer)
    walkFunc := getNames(filenames)
    go func() {
        filepath.Walk(root, walkFunc)
        close(filenames)
    }()
    for i := 0; i < cpus; i++ {
        go feedCheckFile(toProcess, done)
    }
    go func() {
        for filename := range filenames {
            toProcess <- filename
        }
        close(toProcess)
    }()
    for i := 0; i < cpus; i++ {
        <-done
    }
}

// feedCheckFile receives a channel of strings and a done
// channel (of bool) and calls checkFile in a loop. It would be
// easier to rewrite checkFile to just add a simple loop, but this
// allows the checkFile code to remain simpler in case I decide
// to use it differently later.
func feedCheckFile(filenames chan string, done chan bool) {
    for val := range filenames {
        checkFile(val)
    }
    done <- true
}

// checkFile takes a filename and reads the file to determine
// whether the file contains the regex in the global pattern.
func checkFile(filename string) {
    file, err := os.Open(filename)
    if err != nil {
        log.Println(err)
        return
    }
    defer file.Close()
    scanner := bufio.NewScanner(file)
    line := 0
    for scanner.Scan() {
        line += 1
        found := pattern.FindIndex(scanner.Bytes())
        if found != nil {
            fmt.Printf("%s:%d: %s\n", filename, line, scanner.Text())
        }
    }
}
