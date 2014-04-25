package main

import (
    "fmt"
    "path/filepath"
    "flags"
)

func dumper(c chan string) filepath.WalkFunc {

    return func (path string, info os.FileInfo, err error) error {
        c <- string
        return nil
    }


}

func main() {
    var filenames channel string
    var root string

    f := dumper(filenames)

    
    flags.parse()
    if len(flags.Argv) == 0 {
        root = "."
    } else {
        root = flags.Argv[0]
    }

    go func() {
        filepath.Walk(root, f)
        close(filenames)
    }()

    count := 0

    for filename := range(filenames) {
        count += 1
    }

    fmt.Printf("%d files found.\n", count)

}
