// © Knug Industries 2011 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package say // Perl6 inspiration here :-)

import(
    "fmt"
    "io"
    "os"
)

// package to turn of all print statements

var mute bool = false

func Mute() {
    mute = true
}

func Sound() {
    mute = false
}

func Print(args ...interface{})(int, os.Error) {
    if ! mute {
        return fmt.Print(args...)
    }
    return 0, nil
}

func Println(args ...interface{})(int, os.Error) {
    if ! mute {
        return fmt.Println(args...)
    }
    return 0, nil
}

func Printf(f string, args ...interface{})(int, os.Error) {
    if ! mute {
        return fmt.Printf(f, args...)
    }
    return 0, nil
}


func Fprint(w io.Writer, args ...interface{})(int, os.Error) {
    if ! mute {
        return fmt.Fprint(w, args...)
    }
    return 0, nil
}

func Fprintln(w io.Writer, args ...interface{})(int, os.Error) {
    if ! mute {
        return fmt.Fprintln(w, args...)
    }
    return 0, nil
}

func Fprintf(w io.Writer, f string, args ...interface{})(int, os.Error) {
    if ! mute {
        fmt.Fprintf(w, f, args...)
    }
    return 0, nil
}
