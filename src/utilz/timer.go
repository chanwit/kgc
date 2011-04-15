// © Knug Industries 2011 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package timer

import(
    "time"
    "os"
    "io"
    "fmt"
)

// a very simple timer package, perhaps to simple

// nanosecond: Millisecond, Second, Minute, Hour
const(
    Millisecond = 1e6
    Second = 1e9
    Minute = 60*Second
    Hour   = 60*Minute
)

type Time struct {
  Milliseconds, Seconds, Minutes, Hours int64
}

// jobname -> epoch-ns [ns used when stopped]
var jobs map[string]int64
// jobname -> running true or false
var running map[string]bool

func init(){
    jobs = make(map[string]int64)
    running = make(map[string]bool)
}

func Start(name string) {
    jobs[name] = time.Nanoseconds()
    running[name] = true
}

func Stop(name string) os.Error {

    started, ok := running[name]

    if !ok {
        return os.NewError("[utilz/timer] unknown job: "+name)
    }

    if !started {
        return os.NewError("[utilz/timer] job not running: "+name)
    }

    jobs[name] = time.Nanoseconds() - jobs[name]
    running[name] = false

    return nil
}

func Resume(name string) os.Error {

    started, ok := running[name]

    if !ok {
        return os.NewError("[utilz/timer] unknown job: "+name)
    }

    if started {
        return os.NewError("[utilz/timer] job is running: "+name)
    }

    jobs[name] = time.Nanoseconds() - jobs[name]
    running[name] = true

    return nil
}

func Delta(name string) (ns int64, err os.Error) {

    delta, ok := jobs[name]

    if !ok {
        return 0, os.NewError("[utilz/timer] unknown job: "+name)
    }

    return delta, nil
}

// positive time interval to Time struct
func Nano2Time(delta int64) *Time {
    if delta < 0 { panic("negative time interval") }
    t := new(Time)
    t.Hours, delta = chunk(Hour, delta)
    t.Minutes, delta = chunk(Minute, delta)
    t.Seconds, delta = chunk(Second, delta)
    t.Milliseconds, delta = chunk(Millisecond, delta)
    return t
}

func chunk(unit, total int64) (units, rest int64) {
    if total > unit {
        units = total/unit
        rest  = total%unit
    }else{
        units = 0
        rest = total
    }
    return units, rest
}

func (t *Time) String() string {

    var r string

    if t.Hours > 0 {
        r = fmt.Sprintf("%dh ",t.Hours)
    }
    if t.Minutes > 0 {
        r = fmt.Sprintf("%s%2dm ", r, t.Minutes)
    }

    return fmt.Sprintf("%s%d.%03ds", r,t.Seconds,t.Milliseconds)
}

func Print(w io.Writer) {
    fmt.Fprintf(w, "--------------------------------\n")
    for k,v := range jobs {
        fmt.Fprintf(w,"%11s   : %14s\n", k, Nano2Time(v))
    }
    fmt.Fprintf(w, "--------------------------------\n")
}
