package main

import (
    "flag"
    "fmt"
    "log"
    "milo/utils"
    "os"
    "path/filepath"
    "strings"
)

var root string
var extensions []string
var pattern string

// getNames creates a filepath.WalkFunc suitable for passing to
// filepath.Walk which passes the filenames found into a channel.
func getNames(c chan string) filepath.WalkFunc {
    return func(path string, info os.FileInfo, err error) error {
        if info.Mode().IsRegular() {
            c <- path
        }
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
}

// getExts sets the extensions global variable,
// removes any extension arguments from args,
// and returns args for further processing.
func getExts(args []string) []string {
    var unused []string
    for _, val := range args {
        if strings.HasPrefix(val, "--") {
            extensions = append(extensions, val)
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
        if utils.IsDir(val) {
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

    filenames := make(chan string, 3333)

    // Make a function containing this channel.
    f := getNames(filenames)

    go func() {
        filepath.Walk(root, f)
        close(filenames)
    }()

    count := 0

    for _ = range filenames {
        count += 1
    }

    fmt.Printf("%d files found.\n", count)

}
