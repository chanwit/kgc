// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package walker /* texas ranger */

import (
    "os"
    "path/filepath"
)

// This package does something along the lines of: find PATH -type f
// Filters can be added on both directory and filenames in order to filter
// the resulting slice of pathnames.


// reassign to filter pathwalk
var IncludeDir = func(p string) bool { return true }
var IncludeFile = func(p string) bool { return true }

type collect struct {
    Files []string
}

func newCollect() *collect {
    c := new(collect)
    c.Files = make([]string, 0)
    return c
}

func (c *collect) VisitDir(path string, d *os.FileInfo) bool {
    return IncludeDir(path)
}

func (c *collect) VisitFile(path string, d *os.FileInfo) {
    if IncludeFile(path) {
        c.Files = append(c.Files, path)
    }
}

func PathWalk(root string) []string {
    c := newCollect()
    errs := make(chan os.Error)
    filepath.Walk(root, c, errs)
    return c.Files
}
