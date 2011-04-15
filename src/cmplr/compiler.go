// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package compiler

import (
    "os"
    "fmt"
    "path"
    "log"
    "exec"
    "strings"
    "regexp"
    "utilz/walker"
    "utilz/stringset"
    "utilz/handy"
    "utilz/say"
    "utilz/global"
    "cmplr/dag"
)


var includes []string
var srcroot string
var libroot string
var pathLinker string
var pathCompiler string
var suffix string


func Init(srcdir, arch string, include []string){

    var A string // a:architecture
    var err os.Error

    if arch == "" {
        A = os.Getenv("GOARCH")
    } else {
        A = arch
    }

    var S, C, L string // S:suffix, C:compiler, L:linker

    switch A {
    case "arm":
        S = ".5"
        C = "5g"
        L = "5l"
    case "amd64":
        S = ".6"
        C = "6g"
        L = "6l"
    case "386":
        S = ".8"
        C = "8g"
        L = "8l"
    default:
        log.Fatalf("[ERROR] unknown architecture: %s\n", A)
    }

    pathCompiler, err = exec.LookPath(C)

    if err != nil {
        log.Fatalf("[ERROR] could not find compiler: %s\n", C)
    }

    pathLinker, err = exec.LookPath(L)

    if err != nil {
        log.Fatalf("[ERROR] could not find linker: %s\n", L)
    }

    suffix   = S
    srcroot  = srcdir
    includes = include

    if global.GetString("-lib") != "" {
        libroot = global.GetString("-lib")
    }else{
        libroot = srcroot
    }
}


func CreateArgv(pkgs []*dag.Package) {

    var argv []string

    includeLen := len(includes)

    for y := 0; y < len(pkgs); y++ {

        argv = make([]string, 5 + len(pkgs[y].Files) + (includeLen*2))
        i := 0
        argv[i] = pathCompiler
        i++
        argv[i] = "-I"
        i++
        argv[i] = libroot
        i++
        for y := 0; y < includeLen; y++ {
            argv[i] = "-I"
            i++
            argv[i] = includes[y]
            i++
        }
        argv[i] = "-o"
        i++
        argv[i] = path.Join(libroot, pkgs[y].Name) + suffix
        i++

        for z := 0; z < len(pkgs[y].Files); z++ {
            argv[i] = pkgs[y].Files[z]
            i++
        }

        pkgs[y].Argv = argv
    }
}

func CreateLibArgv(pkgs []*dag.Package) {

    ss := stringset.New()
    for i := range pkgs {
        if len(pkgs[i].Name) > len(pkgs[i].ShortName) {
            ss.Add(pkgs[i].Name[:(len(pkgs[i].Name) - len(pkgs[i].ShortName))])
        }
    }
    slice := ss.Slice()
    for i := 0; i < len(slice); i++ {
        slice[i] = path.Join(libroot, slice[i])
        handy.DirOrMkdir(slice[i])
    }

    CreateArgv(pkgs)

}

func SerialCompile(pkgs []*dag.Package) {

    var oldPkgFound bool = false

    for y := 0; y < len(pkgs); y++ {

        if global.GetBool("-dryrun") {
            fmt.Printf("%s || exit 1\n", strings.Join(pkgs[y].Argv, " "))
        } else {
            if oldPkgFound || !pkgs[y].UpToDate() {
                say.Println("compiling:", pkgs[y].Name)
                handy.StdExecve(pkgs[y].Argv, true)
                oldPkgFound = true
            } else {
                say.Println("up 2 date:", pkgs[y].Name)
            }
        }
    }
}

func ParallelCompile(pkgs []*dag.Package) {

    var localDeps *stringset.StringSet
    var compiledDeps *stringset.StringSet
    var y, z, count int
    var parallel []*dag.Package
    var oldPkgFound bool = false
    var zeroFirst []*dag.Package

    localDeps = stringset.New()
    compiledDeps = stringset.New()

    for y = 0; y < len(pkgs); y++ {
        localDeps.Add(pkgs[y].Name)
        pkgs[y].ResetIndegree()
    }

    zeroFirst = make([]*dag.Package, len(pkgs))

    for y = 0; y < len(pkgs); y++ {
        if pkgs[y].Indegree == 0 {
            zeroFirst[count] = pkgs[y]
            count++
        }
    }

    for y = 0; y < len(pkgs); y++ {
        if pkgs[y].Indegree > 0 {
            zeroFirst[count] = pkgs[y]
            count++
        }
    }

    parallel = make([]*dag.Package, 0)

    for y = 0; y < len(zeroFirst); {

        if !zeroFirst[y].Ready(localDeps, compiledDeps) {

            oldPkgFound = compileMultipe(parallel, oldPkgFound)

            for z = 0; z < len(parallel); z++ {
                compiledDeps.Add(parallel[z].Name)
            }

            parallel = make([]*dag.Package, 0)

        } else {
            parallel = append(parallel, zeroFirst[y])
            y++
        }
    }

    if len(parallel) > 0 {
        _ = compileMultipe(parallel, oldPkgFound)
    }

}

func compileMultipe(pkgs []*dag.Package, oldPkgFound bool) bool {

    var ok bool
    var max int = len(pkgs)
    var trouble bool = false

    if max == 0 {
        log.Fatal("[ERROR] trying to compile 0 packages in parallel\n")
    }

    if max == 1 {
        if oldPkgFound || !pkgs[0].UpToDate() {
            say.Println("compiling:", pkgs[0].Name)
            handy.StdExecve(pkgs[0].Argv, true)
            oldPkgFound = true
        } else {
            say.Println("up 2 date:", pkgs[0].Name)
        }
    } else {

        ch := make(chan bool, max)

        for y := 0; y < max; y++ {
            if oldPkgFound || !pkgs[y].UpToDate() {
                say.Println("compiling:", pkgs[y].Name)
                oldPkgFound = true
                go gCompile(pkgs[y].Argv, ch)
            } else {
                say.Println("up 2 date:", pkgs[y].Name)
                ch <- true
            }
        }

        // drain channel (make sure all jobs are finished)
        for z := 0; z < max; z++ {
            ok = <-ch
            if !ok {
                trouble = true
            }
        }
    }

    if trouble {
        log.Fatal("[ERROR] failed batch compile job\n")
    }

    return oldPkgFound
}

func gCompile(argv []string, c chan bool) {
    ok := handy.StdExecve(argv, false) // don't exit on error
    c <- ok
}

// for removal of temoprary packages created for testing and so on..
func DeletePackages(pkgs []*dag.Package) bool {

    var ok = true
    var e os.Error

    for i := 0; i < len(pkgs); i++ {

        for y := 0; y < len(pkgs[i].Files); y++ {
            e = os.Remove(pkgs[i].Files[y])
            if e != nil {
                ok = false
                log.Printf("[ERROR] %s\n", e)
            }
        }
        if ! global.GetBool("-dryrun") {
            pcompile := path.Join(libroot, pkgs[i].Name) + suffix
            e = os.Remove(pcompile)
            if e != nil {
                ok = false
                log.Printf("[ERROR] %s\n", e)
            }
        }
    }

    return ok
}

func ForkLink(output string, pkgs []*dag.Package) {

    var mainPKG *dag.Package

    gotMain := make([]*dag.Package, 0)

    for i := 0; i < len(pkgs); i++ {
        if pkgs[i].ShortName == "main" {
            gotMain = append(gotMain, pkgs[i])
        }
    }

    if len(gotMain) == 0 {
        log.Fatal("[ERROR] (linking) no main package found\n")
    }

    if len(gotMain) > 1 {
        choice := mainChoice(gotMain)
        mainPKG = gotMain[choice]
    } else {
        mainPKG = gotMain[0]
    }

    staticXtra := 0
    if global.GetBool("-static") {
        staticXtra++
    }

    compiled := path.Join(libroot, mainPKG.Name) + suffix

    argv := make([]string, 6+(len(includes)*2)+staticXtra)
    i := 0
    argv[i] = pathLinker
    i++
    argv[i] = "-L"
    i++
    argv[i] = libroot
    i++
    argv[i] = "-o"
    i++
    argv[i] = output
    i++
    if global.GetBool("-static") {
        argv[i] = "-d"
        i++
    }
    for y := 0; y < len(includes); y++ {
        argv[i] = "-L"
        i++
        argv[i] = includes[y]
        i++
    }
    argv[i] = compiled
    i++

    if global.GetBool("-dryrun") {
        fmt.Printf("%s || exit 1\n", strings.Join(argv, " "))
    } else {
        say.Println("linking  :", output)
        handy.StdExecve(argv, true)
    }
}

func mainChoice(pkgs []*dag.Package) int {

    var cnt int
    var choice int

    for i := 0; i < len(pkgs); i++ {
        ok, _ := regexp.MatchString(global.GetString("-main"), pkgs[i].Name)
        if ok {
            cnt++
            choice = i
        }
    }

    if cnt == 1 {
        return choice
    }

    fmt.Println("\n More than one main package found\n")

    for i := 0; i < len(pkgs); i++ {
        fmt.Printf(" type %2d  for: %s\n", i, pkgs[i].Name)
    }


    fmt.Printf("\n type your choice: ")

    n, e := fmt.Scanf("%d", &choice)

    if e != nil {
        log.Fatalf("%s\n", e)
    }
    if n != 1 {
        log.Fatal("failed to read input\n")
    }

    if choice >= len(pkgs) || choice < 0 {
        log.Fatalf(" bad choice: %d\n", choice)
    }

    fmt.Printf(" chosen main-package: %s\n\n", pkgs[choice].Name)

    return choice
}


func CreateTestArgv() []string {

    var numArgs int = 1

    pwd, e := os.Getwd()

    if e != nil {
        log.Fatal("[ERROR] could not locate working directory\n")
    }

    arg0 := path.Join(pwd, global.GetString("-test-bin"))

    if global.GetString("-bench") != "" {
        numArgs += 2
    }
    if global.GetString("-match") != "" {
        numArgs += 2
    }
    if global.GetBool("-verbose") {
        numArgs++
    }

    var i = 1
    argv := make([]string, numArgs)
    argv[0] = arg0
    if global.GetString("-bench") != "" {
        argv[i] = "-test.bench"
        i++
        argv[i] = global.GetString("-bench")
        i++
    }
    if global.GetString("-match") != "" {
        argv[i] = "-test.run"
        i++
        argv[i] = global.GetString("-match")
        i++
    }
    if global.GetBool("-verbose") {
        argv[i] = "-test.v"
    }
    return argv
}

func Remove865(dir string, alsoDir bool) {

    // override IncludeFile to make walker pick up only .[865] files
    walker.IncludeFile = func(s string) bool {
        return strings.HasSuffix(s, ".8") ||
            strings.HasSuffix(s, ".6") ||
            strings.HasSuffix(s, ".5")
    }

    handy.DirOrExit(dir)

    compiled := walker.PathWalk(path.Clean(dir))

    for i := 0; i < len(compiled); i++ {

        if ! global.GetBool("-dryrun") {

            e := os.Remove(compiled[i])
            if e != nil {
                log.Printf("[ERROR] could not delete file: %s\n", compiled[i])
            } else {
                say.Printf("rm: %s\n", compiled[i])
            }

        } else {
            fmt.Printf("[dryrun] rm: %s\n", compiled[i])
        }
    }

    if alsoDir {
        // remove entire dir if empty after objects are deleted
        walker.IncludeFile = func (s string) bool { return true }
        walker.IncludeDir = func (s string) bool { return true }
        if len( walker.PathWalk(dir) ) == 0 {
            if global.GetBool("-dryrun") {
                fmt.Printf("[dryrun] rm: %s\n",dir)
            }else{
                say.Printf("rm: %s\n", dir)
                e := os.RemoveAll(dir)
                if e != nil {
                    log.Fatalf("[ERROR] %s\n", e)
                }
            }
        }
    }
}


func FormatFiles(files []string) {

    var i, argvLen int
    var argv []string
    var tabWidth string = "-tabwidth=4"
    var useTabs string = "-tabindent=false"
    var comments string = "-comments=true"
    var rewRule string = global.GetString("-rew-rule")
    var fmtexec string
    var err os.Error

    fmtexec, err = exec.LookPath("gofmt")

    if err != nil {
        log.Fatal("[ERROR] could not find 'gofmt' in $PATH")
    }

    if global.GetString("-tabwidth") != "" {
        tabWidth = "-tabwidth=" + global.GetString("-tabwidth")
    }
    if global.GetBool("-no-comments") {
        comments = "-comments=false"
    }
    if rewRule != "" {
        argvLen++
    }
    if global.GetBool("-tab") {
        useTabs = "-tabindent=true"
    }

    argv = make([]string, 6+argvLen)

    if fmtexec == "" {
        log.Fatal("[ERROR] could not find: gofmt\n")
    }

    argv[i] = fmtexec
    i++
    argv[i] = "-w=true"
    i++
    argv[i] = tabWidth
    i++
    argv[i] = useTabs
    i++
    argv[i] = comments
    i++

    if rewRule != "" {
        argv[i] = fmt.Sprintf("-r='%s'", rewRule)
        i++
    }

    for y := 0; y < len(files); y++ {
        argv[i] = files[y]
        if ! global.GetBool("-dryrun") {
            say.Printf("gofmt : %s\n", files[y])
            _ = handy.StdExecve(argv, true)
        } else {
            fmt.Printf(" %s\n", strings.Join(argv, " "))
        }
    }
}
