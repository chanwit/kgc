// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package gopt_test

import (
    "testing"
    "strings"
    "parse/gopt"
)

func TestGetOpt(t *testing.T) {

    getopt := gopt.New()

    getopt.BoolOption("-h -help --help")
    getopt.BoolOption("-v -version --version")
    getopt.StringOption("-f -file --file --file=")
    getopt.StringOption("-num=")
    getopt.StringOption("-I")

    argv := strings.Split("-h -num=7 -version not-option -fsomething -I/dir1 -I/dir2", " ", -1)

    args := getopt.Parse(argv)

    if !getopt.IsSet("-help") {
        t.Fatal("! getopt.IsSet('-help')\n")
    }

    if !getopt.IsSet("-version") {
        t.Fatal(" ! getopt.IsSet('-version')\n")
    }

    if !getopt.IsSet("-file") {
        t.Fatal(" ! getopt.IsSet('-file')\n")
    } else {
        if getopt.Get("-f") != "something" {
            t.Fatal(" getopt.Get('-f') != 'something'\n")
        }
    }

    if !getopt.IsSet("-num=") {
        t.Fatal(" ! getopt.IsSet('-num=')\n")
    }else{
        n,e := getopt.GetInt("-num=")
        if e != nil {
            t.Fatalf(" getopt.GetInt error = %s\n", e)
        }
        if n != 7 {
            t.Fatalf(" getopt.GetInt != 7 (%d)\n", n)
        }
    }

    if !getopt.IsSet("-I") {
        t.Fatal(" ! getopt.IsSet('-I')\n")
    } else {
        elms := getopt.GetMultiple("-I")
        if len(elms) != 2 {
            t.Fatal("getopt.GetMultiple('-I') != 2\n")
        }
        if elms[0] != "/dir1" {
            t.Fatal("getopt.GetMultiple('-I')[0] != '/dir1'\n")
        }
        if elms[1] != "/dir2" {
            t.Fatal("getopt.GetMultiple('-I')[1] != '/dir2'\n")
        }
    }

    if len(args) != 1 {
        t.Fatal("len(remaining) != 1\n")
    }

    if args[0] != "not-option" {
        t.Fatal("remaining[0] != 'not-something'\n")
    }
}
