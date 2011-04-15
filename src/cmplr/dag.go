// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package dag

import (
    "exec"
    "go/parser"
    "go/token"
    "path"
    "go/ast"
    "os"
    "fmt"
    "time"
    "log"
    "strings"
    "regexp"
    "utilz/stringset"
    "utilz/stringbuffer"
    "utilz/handy"
    "utilz/global"
    "utilz/say"
)


type Dag map[string]*Package // package-name -> Package object

type Package struct {
    Indegree        int
    Name, ShortName string               // absolute path, basename
    Argv            []string             // command needed to compile package
    Files           []string             // relative path of files
    dependencies    *stringset.StringSet
    children        []*Package           // packages that depend on this
}

type TestCollector struct {
    Names []string
}

func New() Dag {
    return make(map[string]*Package)
}

func newPackage() *Package {
    p := new(Package)
    p.Indegree = 0
    p.Files = make([]string, 0)
    p.dependencies = stringset.New()
    p.children = make([]*Package, 0)
    return p
}

func newTestCollector() *TestCollector {
    t := new(TestCollector)
    t.Names = make([]string, 0)
    return t
}


func (d Dag) Parse(root string, files []string) {

    root = addSeparatorPath(root)

    var e, pkgname string

    for i := 0; i < len(files); i++ {
        e = files[i]
        tree := getSyntaxTreeOrDie(e, parser.ImportsOnly)
        dir, _ := path.Split(e)
        unroot := dir[len(root):len(dir)]
        shortname := tree.Name.String()

        // if package name == directory name -> assume stdlib organizing
        if len(unroot) > 1 && path.Base(dir) == shortname {
            pkgname = unroot[:len(unroot)-1]
        }else{
            pkgname = path.Join(unroot, shortname)
        }

        _, ok := d[pkgname]
        if !ok {
            d[pkgname] = newPackage()
            d[pkgname].Name = pkgname
            d[pkgname].ShortName = shortname
        }

        ast.Walk(d[pkgname], tree)
        d[pkgname].Files = append(d[pkgname].Files, e)
    }
}

func (d Dag) addEdge(from, to string) {
    fromNode := d[from]
    toNode := d[to]
    fromNode.children = append(fromNode.children, toNode)
    toNode.Indegree++
}
// note that nothing is done in order to check if dependencies
// are valid if they are not part of the actual source-tree.

func (d Dag) GraphBuilder() {

    for k, v := range d {
        for dep := range v.dependencies.Iter() {
            if d.localDependency(dep) {
                d.addEdge(dep, k)
                ///fmt.Printf("local:  %s \n", dep);
            }
        }
    }
}

func (d Dag) External() {

    var err os.Error
    var argv []string
    var argc int = 2
    var i int = 0


    set := stringset.New()

    for _, v := range d {
        for dep := range v.dependencies.Iter() {
            if ! d.localDependency(dep) {
                set.Add(dep)
            }
        }
    }

    for u := range set.Iter() {
        if ! seemsExternal( u ) {
            set.Remove( u )
        }
    }

    if global.GetBool("-verbose") {
        argc++
    }

    argv = make([]string, argc)

    argv[i], err = exec.LookPath("goinstall")
    i++

    if err != nil {
        log.Fatalf("[ERROR] %s\n", err)
    }

    if global.GetBool("-verbose") {
        argv[i] = "-v=true"
        i++
    }

    for u := range set.Iter() {
        argv[i] = u
        if global.GetBool("-dryrun") {
            fmt.Printf("%s || exit 1\n", strings.Join(argv, " "))
        }else{
            say.Printf("goinstall: %s\n", u)
            handy.StdExecve(argv, true)
        }
    }

}

// If import starts with one of these, it seems legal...
//
//  bitbucket.org/
//  github.com/
//  [^.]+\.googlecode\.com/
//  launchpad.net/
func seemsExternal(imprt string) (bool) {

    if strings.HasPrefix(imprt, "bitbucket.org/") {
        return true
    }else if strings.HasPrefix(imprt, "github.com/") {
        return true
    }else if strings.HasPrefix(imprt, "launchpad.net/") {
        return true
    }

    ok, _ := regexp.MatchString("[^.]\\.googlecode\\.com\\/.*", imprt)

    return ok
}

func (d Dag) MakeDotGraph(filename string) {

    var file *os.File
    var fileinfo *os.FileInfo
    var e os.Error
    var sb *stringbuffer.StringBuffer

    fileinfo, e = os.Stat(filename)

    if e == nil {
        if fileinfo.IsRegular() {
            e = os.Remove(fileinfo.Name)
            if e != nil {
                log.Fatalf("[ERROR] failed to remove: %s\n", filename)
            }
        }
    }

    sb = stringbuffer.NewSize(500)
    file, e = os.Open(filename, os.O_WRONLY|os.O_CREAT, 0644)

    if e != nil {
        log.Fatalf("[ERROR] %s\n", e)
    }

    sb.Add("digraph depgraph {\n\trankdir=LR;\n")

    for _, v := range d {
        v.DotGraph(sb)
    }

    sb.Add("}\n")

    file.WriteString(sb.String())

    file.Close()

}

func (d Dag) MakeMainTest(root string) ([]*Package, string) {

    var max, i int
    var isTest bool
    var sname, tmpdir, tmpstub, tmpfile string

    sbImports := stringbuffer.NewSize(300)
    imprtSet  := stringset.New()
    sbTests := stringbuffer.NewSize(1000)
    sbBench := stringbuffer.NewSize(1000)

    sbImports.Add("\n// autogenerated code\n\n")
    sbImports.Add("package main\n\n")
    imprtSet.Add("import \"regexp\"\n")
    imprtSet.Add("import \"testing\"\n")


    sbTests.Add("\n\nvar tests = []testing.InternalTest{\n")
    sbBench.Add("\n\nvar benchmarks = []testing.InternalBenchmark{\n")

    for _, v := range d {

        isTest = false
        sname = v.ShortName
        max = len(v.ShortName)

        if max > 5 && sname[max-5:] == "_test" {
            collector := newTestCollector()
            for i = 0; i < len(v.Files); i++ {
                tree := getSyntaxTreeOrDie(v.Files[i], 0)
                ast.Walk(collector, tree)
            }

            if len(collector.Names) > 0 {
                isTest = true
                imprtSet.Add(fmt.Sprintf("import \"%s\"\n", v.Name))
                for i = 0; i < len(collector.Names); i++ {
                    testFunc := collector.Names[i]
                    if len(testFunc) >= 4 && testFunc[0:4] == "Test" {
                        sbTests.Add(fmt.Sprintf("testing.InternalTest{\"%s.%s\", %s.%s },\n",
                            sname, testFunc, sname, testFunc))
                    } else if len(testFunc) >= 9 && testFunc[0:9] == "Benchmark" {
                        sbBench.Add(fmt.Sprintf("testing.InternalBenchmark{\"%s.%s\", %s.%s },\n",
                            sname, testFunc, sname, testFunc))

                    }
                }
            }
        }

        if !isTest {

            collector := newTestCollector()

            for i = 0; i < len(v.Files); i++ {
                fname := v.Files[i]
                if len(fname) > 8 && fname[len(fname)-8:] == "_test.go" {
                    tree := getSyntaxTreeOrDie(fname, 0)
                    ast.Walk(collector, tree)
                }
            }

            if len(collector.Names) > 0 {
                imprtSet.Add(fmt.Sprintf("import \"%s\"\n", v.Name))
                for i = 0; i < len(collector.Names); i++ {
                    testFunc := collector.Names[i]
                    if len(testFunc) >= 4 && testFunc[0:4] == "Test" {
                        sbTests.Add(fmt.Sprintf("testing.InternalTest{\"%s.%s\", %s.%s },\n",
                            sname, testFunc, sname, testFunc))
                    } else if len(testFunc) >= 9 && testFunc[0:9] == "Benchmark" {
                        sbBench.Add(fmt.Sprintf("testing.InternalBenchmark{\"%s.%s\", %s.%s },\n",
                            sname, testFunc, sname, testFunc))
                    }
                }
            }
        }
    }

    sbTests.Add("};\n")
    sbBench.Add("};\n\n")

    for im := range imprtSet.Iter() {
        sbImports.Add( im )
    }

    sbTotal := stringbuffer.NewSize(sbImports.Len() +
        sbTests.Len() +
        sbBench.Len() + 5)
    sbTotal.Add(sbImports.String())
    sbTotal.Add(sbTests.String())
    sbTotal.Add(sbBench.String())

    sbTotal.Add("func main(){\n")
    sbTotal.Add("testing.Main(regexp.MatchString, tests);\n")
    sbTotal.Add("testing.RunBenchmarks(regexp.MatchString, benchmarks);\n}\n\n")

    tmpstub = fmt.Sprintf("tmp%d", time.Seconds())
    tmpdir = fmt.Sprintf("%s%s", addSeparatorPath(root), tmpstub)

    dir, e1 := os.Stat(tmpdir)

    if e1 == nil && dir.IsDirectory() {
        log.Printf("[ERROR] directory: %s already exists\n", tmpdir)
    } else {
        e_mk := os.Mkdir(tmpdir, 0777)
        if e_mk != nil {
            log.Fatal("[ERROR] failed to create directory for testing")
        }
    }

    tmpfile = path.Join(tmpdir, "main.go")

    fil, e2 := os.Open(tmpfile, os.O_WRONLY|os.O_CREAT, 0777)

    if e2 != nil {
        log.Fatalf("[ERROR] %s\n", e2)
    }

    n, e3 := fil.WriteString(sbTotal.String())

    if e3 != nil {
        log.Fatalf("[ERROR] %s\n", e3)
    } else if n != sbTotal.Len() {
        log.Fatal("[ERROR] failed to write test")
    }

    fil.Close()

    p := newPackage()
    p.Name = path.Join(tmpstub, "main")
    p.ShortName = "main"
    p.Files = append(p.Files, tmpfile)

    vec := make([]*Package, 1)
    vec[0] = p
    return vec, tmpdir
}

func (d Dag) Topsort() []*Package {

    var node, child *Package
    var cnt int = 0

    zero := make([]*Package, 0)
    done := make([]*Package, 0)

    for _, v := range d {
        if v.Indegree == 0 {
            zero = append(zero, v)
        }
    }

    for len(zero) > 0 {

        node = zero[0]
        zero = zero[1:] // Pop

        for i := 0; i < len(node.children); i++ {
            child = node.children[i]
            child.Indegree--
            if child.Indegree == 0 {
                zero = append(zero, child)
            }
        }
        cnt++
        done = append(done, node)
    }

    if cnt < len(d) {
        log.Fatal("[ERROR] loop in dependency graph")
    }

    return done
}

func (d Dag) localDependency(dep string) bool {
    _, ok := d[dep]
    return ok
}

func (d Dag) PrintInfo() {

    var i int

    fmt.Println("--------------------------------------")
    fmt.Println("Packages and Dependencies")
    fmt.Println("p = package, f = file, d = dependency ")
    fmt.Println("--------------------------------------\n")

    for k, v := range d {
        fmt.Println("p ", k)
        for i = 0; i < len(v.Files); i++ {
            fmt.Println("f ", v.Files[i])
        }
        for ds := range v.dependencies.Iter() {
            fmt.Println("d ", ds)
        }
        fmt.Println("")
    }
}

func (p *Package) DotGraph(sb *stringbuffer.StringBuffer) {

    if p.dependencies.Len() == 0 {

        sb.Add(fmt.Sprintf("\t\"%s\";\n", p.Name))

    } else {

        for dep := range p.dependencies.Iter() {
            sb.Add(fmt.Sprintf("\t\"%s\" -> \"%s\";\n", p.Name, dep))
        }
    }
}


func (p *Package) UpToDate() bool {

    if p.Argv == nil {
        log.Fatalf("[ERROR] missing dag.Package.Argv\n")
    }

    var e os.Error
    var finfo *os.FileInfo
    var compiledModifiedTime int64
    var last, stop, i int
    var resultingFile string

    last = len(p.Argv) - 1
    stop = last - len(p.Files)
    resultingFile = p.Argv[stop]

    finfo, e = os.Stat(resultingFile)

    if e != nil {
        return false
    } else {
        compiledModifiedTime = finfo.Mtime_ns
    }

    for i = last; i > stop; i-- {
        finfo, e = os.Stat(p.Argv[i])
        if e != nil {
            panic(fmt.Sprintf("Missing go file: %s\n", p.Argv[i]))
        } else {
            if finfo.Mtime_ns > compiledModifiedTime {
                return false
            }
        }
    }

    return true
}

func (p *Package) Ready(local, compiled *stringset.StringSet) bool {

    for dep := range p.dependencies.Iter() {
        if local.Contains(dep) && !compiled.Contains(dep) {
            return false
        }
    }

    return true
}

func (p *Package) ResetIndegree(){
    for i := 0; i < len(p.children); i++ {
        p.children[i].Indegree++
    }
}

func (p *Package) Visit(node ast.Node) (v ast.Visitor) {

    switch node.(type) {
    case *ast.BasicLit:
        bl, ok := node.(*ast.BasicLit)
        if ok {
            stripped := string(bl.Value[1:len(bl.Value)-1])
            p.dependencies.Add(stripped)
        }
    default: // nothing to do if not BasicLit
    }
    return p
}

func (t *TestCollector) Visit(node ast.Node) (v ast.Visitor) {
    switch node.(type) {
    case *ast.FuncDecl:
        fdecl, ok := node.(*ast.FuncDecl)
        if ok {
            t.Names = append(t.Names, fdecl.Name.Name)
        }
    default: // nothing to do if not FuncDecl
    }
    return t
}

func addSeparatorPath(root string) string {
    if root[len(root)-1:] != "/" {
        root = root + "/"
    }
    return root
}

func getSyntaxTreeOrDie(file string, mode uint) *ast.File {
    absSynTree, err := parser.ParseFile(token.NewFileSet(), file, nil, mode)
    if err != nil {
        log.Fatalf("%s\n", err)
    }
    return absSynTree
}
