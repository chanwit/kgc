// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package gopt

/*

The flag package provided in the go standard
library will only allow options before regular
input arguments, which is not ideal. So this
is another take on the old getopt.

Most notable difference:

  - Multiple 'option-strings' for a single option ('-r -rec -R')
  - Non-option arguments can come anywhere in argv
  - Option arguments can be in juxtaposition with flag
  - Only two types of options: string, bool


Usage:

 getopt := gopt.New();

 getopt.BoolOption("-h -help --help");
 getopt.BoolOption("-v -version --version");
 getopt.StringOption("-f -file --file --file=");
 getopt.StringOption("-l -list --list");
 getopt.StringOption("-I");

 args := getopt.Parse(os.Args[1:]);

 // getopt.IsSet("-h") == getopt.IsSet("-help") ..

 if getopt.IsSet("-help"){ println("-help"); }
 if getopt.IsSet("-v")   { println("-version"); }
 if getopt.IsSet("-file"){ println("--file ",getopt.Get("-f")); }
 if getopt.IsSet("-list"){ println("--list ",getopt.Get("-list")); }

 if getopt.IsSet("-I"){
     elms := getopt.GetMultiple("-I");
     for y := range elms { println("-I ",elms[y]);  }
 }

 for i := range args{
     println("remaining:",args[i]);
 }


*/


import (
    "strings"
    "log"
    "os"
    "strconv"
)

type GetOpt struct {
    options []Option
    cache   map[string]Option
}

func New() *GetOpt {
    g := new(GetOpt)
    g.options = make([]Option, 0)
    g.cache = make(map[string]Option)
    return g
}

func (g *GetOpt) isOption(o string) Option {
    _, ok := g.cache[o]
    if ok {
        element, _ := g.cache[o]
        return element
    }
    return nil
}

func (g *GetOpt) getStringOption(o string) *StringOption {

    opt := g.isOption(o)

    if opt != nil {
        sopt, ok := opt.(*StringOption)
        if ok {
            return sopt
        } else {
            log.Fatalf("[ERROR] %s: is not a string option\n", o)
        }
    } else {
        log.Fatalf("[ERROR] %s: is not an option at all\n", o)
    }

    return nil
}

func (g *GetOpt) Get(o string) string {

    sopt := g.getStringOption(o)

    switch sopt.count {
    case 0:
        log.Fatalf("[ERROR] %s: is not set\n", o)
    case 1: // fine do nothing
    default:
        log.Printf("[WARNING] option %s: has more arguments than 1\n", o)
    }
    return sopt.values[0]
}

func (g *GetOpt) GetFloat32(o string) (float32, os.Error) {
    return strconv.Atof32( g.Get(o) )
}

func (g *GetOpt) GetFloat64(o string) (float64, os.Error) {
    return strconv.Atof64( g.Get(o) )
}

func (g *GetOpt) GetInt(o string) (int, os.Error) {
    return strconv.Atoi( g.Get(o) )
}

func (g *GetOpt) Reset() {
    for _, v := range g.cache {
        v.reset()
    }
}

func (g *GetOpt) GetMultiple(o string) []string {

    sopt := g.getStringOption(o)

    if sopt.count == 0 {
        log.Fatalf("[ERROR] %s: is not set\n", o)
    }

    return sopt.values[0:sopt.count]
}

func (g *GetOpt) Parse(argv []string) (args []string) {

    var count int = 0
    // args cannot be longer than argv, if no options
    // are given on the command line it is argv
    args = make([]string, len(argv))

    for i := 0; i < len(argv); i++ {

        opt := g.isOption(argv[i])

        if opt != nil {

            switch opt.(type) {
            case *BoolOption:
                bopt, _ := opt.(*BoolOption)
                bopt.setFlag()
            case *StringOption:
                sopt, _ := opt.(*StringOption)
                if i+1 >= len(argv) {
                    log.Fatalf("[ERROR] missing argument for: %s\n", argv[i])
                } else {
                    sopt.addArgument(argv[i+1])
                    i++
                }
            }

        } else {

            // arguments written next to options
            start, ok := g.juxtaOption(argv[i])

            if ok {
                stropt := g.getStringOption(start)
                stropt.addArgument(argv[i][len(start):])
            } else {
                args[count] = argv[i]
                count++
            }
        }
    }

    return args[0:count]
}

func (g *GetOpt) juxtaOption(opt string) (string, bool) {

    var tmpmax string = ""

    for i := 0; i < len(g.options); i++ {

        sopt, ok := g.options[i].(*StringOption)

        if ok {
            s := sopt.startsWith(opt)
            if s != "" {
                if len(s) > len(tmpmax) {
                    tmpmax = s
                }
            }
        }
    }

    if tmpmax != "" {
        return tmpmax, true
    }

    return "", false
}

func (g *GetOpt) IsSet(o string) bool {
    _, ok := g.cache[o]
    if ok {
        element, _ := g.cache[o]
        return element.isSet()
    } else {
        log.Fatalf("[ERROR] %s not an option\n", o)
    }
    return false
}

func (g *GetOpt) BoolOption(optstr string) {
    ops := strings.Split(optstr, " ", -1)
    boolopt := newBoolOption(ops)
    for i := range ops {
        g.cache[ops[i]] = boolopt
    }
    g.options = append(g.options, boolopt)
}

func (g *GetOpt) StringOption(optstr string) {
    ops := strings.Split(optstr, " ", -1)
    stringopt := newStringOption(ops)
    for i := range ops {
        g.cache[ops[i]] = stringopt
    }
    g.options = append(g.options, stringopt)
}
