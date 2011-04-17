package main

import (
    "fmt"
    "os"
    "exec"
    "log"

    "korat/parser"
    "korat/token"
    "korat/printer"
    . "korat/ast"

    "github.com/droundy/goopt"
    "container/vector"
    "strconv"

    // "dump"
)

type VisitorImpl struct {
    f *File
    node Node
    parent Node
    history *vector.Vector
    switchStack *vector.Vector
    bindingIndex int
    bindings map[int]Stmt
    caseStmtCount int
}

func NewStruct0(src string) *GenDecl {
    fs := token.NewFileSet()
    if d,ok := parser.ParseDeclList(fs, "temp.go", src); ok==nil {
        return d[0].(*GenDecl)
    }
    return nil
}

func NewStruct(structName string) *GenDecl {
    src := fmt.Sprintf("type %s struct{}", structName)
    fs := token.NewFileSet()
    if d,ok := parser.ParseDeclList(fs, "temp.go", src); ok==nil {
        return d[0].(*GenDecl)
    }
    return nil
}

func NewStmtList(src string) []Stmt {
    fs := token.NewFileSet()
    if e,ok := parser.ParseStmtList(fs, "temp.go", src); ok==nil {
        return e
    }
    return nil
}

func NewExpr(src string) Expr {
    fs := token.NewFileSet()
    if e,ok := parser.ParseExpr(fs, "temp.go", src); ok==nil {
        return e
    }
    return nil
}

func NewMethod(src string) *FuncDecl {
    fs := token.NewFileSet()
    if d,ok := parser.ParseDeclList(fs, "temp.go", src); ok==nil {
        return d[0].(*FuncDecl)
    }
    return nil
}

func MakeMethod(name string, typeName string) *Field {
    return &Field{Names: []*Ident{NewIdent(name)},
                Type: &FuncType{Params: &FieldList{},
                    Results: &FieldList{List:[]*Field{
                        &Field{Type:NewIdent(typeName)}}}}}
}

// World
var interfaces = map[string]*InterfaceType{}
var structs    = map[string]*StructType{}
var patternImported = false

func addPatternImport(f *File) {
    if patternImported {
        return
    }

    i := &ImportSpec {
      Path: &BasicLit {
        Kind: token.STRING,
        Value: []byte("\"korat/pattern\"") }}

    /*
    r := &ImportSpec {
      Path: &BasicLit {
        Kind: token.STRING,
        Value: []byte("\"reflect\"") }}
    */

    gd  := &GenDecl{Tok: token.IMPORT, Specs: []Spec{i}}
    // gd2 := &GenDecl{Tok: token.IMPORT, Specs: []Spec{r}}

    decls := make([]Decl, len(f.Decls)+1)
    decls[0] = gd
    // decls[1] = gd2
    for i:=1; i<len(decls); i++ {
        decls[i] = f.Decls[i-1]
    }
    f.Decls = decls;
    patternImported = true
    return
}

func formatCall(name string, nodes string) string {
    return fmt.Sprintf(
        `&pattern.Node{reflect.Typeof(&%s{}),-1,nil,[]*pattern.Node{%s}}`,
        name,nodes[1:len(nodes)])
}

func formatBasicLit(typ string, value string) string {
    return fmt.Sprintf(
        `&pattern.Node{nil,-1,%s,nil}`,
        /*typ,*/value)
}

func formatIdent(typ string, index int) string {
    return fmt.Sprintf(
        `&pattern.Node{nil,%d,nil,nil}`,
        /*typ,*/index)
}

func lookupType(s string, index int) string {
    st := structs[s]
    return st.Fields.List[index].Type.(*Ident).Name
}

func lookupField(s string, index int) string {
    st := structs[s]
    return st.Fields.List[index].Names[0].Name
}

type BindingSymTab struct {
    count int
    table map[int]string
}

func callToStmt(varname string, root string,
                c *CallExpr, symtab *BindingSymTab) (r string) {
    r = ""
    r = r + fmt.Sprintf("if %s,ok := %s.(*%s); ok {\n",
        varname, root, c.Fun.(*Ident).Name)
    subPrefix := "c"
    rgen, lhs := genAssign(subPrefix, c, varname, c.Args, symtab)
    r = r + rgen
    for i := 0; i < len(c.Args); i++ {
        a := c.Args[i]
        if cA,ok := a.(*CallExpr); ok {
            r = r + callToStmt("c" + strconv.Itoa(i),lhs[i],cA, symtab)
        } else if cV,ok := a.(*BasicLit); ok {
            r = r + fmt.Sprintf("if %s != %s { r = false; return }\n",
                lhs[i], string(cV.Value))
        }
    }
    r = r + fmt.Sprint("} else {\n")
    r = r + fmt.Sprintf("  r = false; return\n")
    r = r + fmt.Sprintf("}\n")
    return
}

func genAssign(prefix string, c *CallExpr,
               varname string, args []Expr,
               symtab *BindingSymTab) (r string, lhs []string) {
    lhs = make([]string, len(args))
    rhs := make([]string, len(args))
    for i:=0; i<len(args); i++ {
        a := args[i]
        fname := lookupField(c.Fun.(*Ident).Name, i)
        assignOp := ":="
        if id,ok := a.(*Ident); ok {
            typ := lookupType(c.Fun.(*Ident).Name,i)
            count := symtab.count
            symtab.table[count] =
                id.Name + " := __b%d.Data[" + strconv.Itoa(count) + "].(" + typ + ");"
            lhs[i] = "b.Data[" + strconv.Itoa(count) + "]" // 
            symtab.count++
            assignOp = "="
        } else if _,ok := a.(*CallExpr); ok {
            lhs[i] = prefix + "_" + fname
        } else if _,ok := a.(*BasicLit); ok {
            lhs[i] = prefix + "_" + fname
        }
        rhs[i] = varname + "." + fname
        r = r + fmt.Sprintf("%s %s %s\n", lhs[i], assignOp, rhs[i])
    }
    return
}

/*
func callToPattern(symtab *BindingSymTab, c *CallExpr) string {
    fmt.Printf("==============\n")
    dump.Dump2(c)
    nodes := ""
    structName := c.Fun.(*Ident).Name
    for i,a := range c.Args {
        switch a.(type) {
            case *CallExpr:
                nodes = nodes + "," + callToPattern(symtab, a.(*CallExpr))
            case *BasicLit:
                bl := a.(*BasicLit)
                nodes = nodes + "," +
                    formatBasicLit(bl.Kind.String(),string(bl.Value))
            case *Ident:
                typ := lookupType(structName,i)
                id  := a.(*Ident).Name

                c := symtab.count
                symtab.table[symtab.count] =
                    id + " := __b%d.Data[" + strconv.Itoa(c) + "].(" + typ + ");"

                symtab.count++
                nodes =  nodes + "," + formatIdent(typ, c)
        }
    }
    return formatCall(structName, nodes)
}
*/

func (v *VisitorImpl) SetParent(node Node) {
    v.parent = node
}

func (v *VisitorImpl) PushSwitchStmt(node Node) {
    v.switchStack.Push(node)
}

func (v *VisitorImpl) PopSwitchStmt() {
    sw := v.switchStack.Pop().(*SwitchStmt)
    sw.Tag = nil
}

func makeMatchCall(v *VisitorImpl, input Expr, index int, caseCount int) *CallExpr {

    b := v.bindings[index].(*AssignStmt).Lhs[0]

    return &CallExpr {
        Fun: &Ident {Name: "__case_"+strconv.Itoa(caseCount)},
        Args: []Expr{input, b},
        Ellipsis: 0 }
}

func (v *VisitorImpl) Visit(node Node) (w Visitor) {
    w = nil
    if node == nil {
        return
    }

    switch node.(type) {
        case *ImportSpec:
            i := node.(*ImportSpec)
            if string(i.Path.Value) == "\"korat/pattern\"" {
                patternImported = true
            }

        case *InterfaceType:
            t := node.(*InterfaceType)
            if t.IsTrait {
                ident := v.node.(*Ident)
                iName := ident.Name
                interfaces[iName] = t
                t.Methods.List = append(t.Methods.List,
                    MakeMethod("asType", iName))
            }

        case *StructType:
            st := node.(*StructType)
            ident := v.node.(*Ident)
            iName := ident.Name
            structs[iName] = st
            if st.IsCaseStruct {

                addPatternImport(v.f)

                params := ""
                values := ""
                unValues := ""
                for i,f := range st.Fields.List {
                    params = params + f.Names[0].Name
                    values = values + f.Names[0].Name
                    unValues = unValues + "u." + f.Names[0].Name
                    for j := 1; j<len(f.Names);j++ {
                        params = params + "," + f.Names[j].Name
                        values = values + "," + f.Names[j].Name
                        unValues = unValues + "u." + f.Names[j].Name
                    }
                    params = params + " " + f.Type.(*Ident).Name
                    if i != len(st.Fields.List)-1 {
                        params = params + ","
                        values = values + ","
                        unValues = unValues + ","
                    }
                }
                v.f.Decls = append(v.f.Decls,
                    NewMethod(fmt.Sprintf(`
                        func New%s(%s) (r *%s) {
                            r = &%s{%s}; return
                        }`,iName,params,iName,iName,values)))
                if len(st.Fields.List) > 1 {
                    v.f.Decls = append(v.f.Decls,
                        NewMethod(fmt.Sprintf(`
                            func (u *%s) Unapply() (s *pattern.Some) {
                                s = &pattern.Some{[]interface{}{%s}}; return
                            }`,iName,unValues)))
                } else {
                    v.f.Decls = append(v.f.Decls,
                        NewMethod(fmt.Sprintf(`
                            func (u *%s) Unapply() (s *pattern.Some) {
                                s = &pattern.Some{%s}; return
                            }`,iName,unValues)))
                }
            }

            if st.BorrowedTrait != nil {
                n := st.BorrowedTrait.(*Ident).Name
                t := interfaces[n]
                for _,l := range t.Methods.List {
                    meth := l.Names[0].Name
                    if meth == "asType" {
                        v.f.Decls = append(v.f.Decls,
                            NewMethod(fmt.Sprintf(`
                                func (this *%s) asType() %s {
                                    return this;
                                }`,iName,n)))
                    } else {
                        v.f.Decls = append(v.f.Decls,
                            NewMethod(fmt.Sprintf(`
                                func (this *%s) %s() {
                                    _%sImpl_%s(this)
                                }`,iName,meth,n,meth)))
                    }
                }
            }
        case *SwitchStmt:
            sw := node.(*SwitchStmt)
            if sw.IsMatch {
                list := v.parent.(*BlockStmt).List
                found := -1
                for i,e := range list {
                    if e == node {
                        found = i
                        break
                    }
                }
                // fmt.Printf("%d\n", found)
                v.bindingIndex++
                bStmt := fmt.Sprintf("__b%d := pattern.NewBinding()",
                    v.bindingIndex)
                toInsert := NewStmtList(bStmt)[0]
                v.bindings[v.bindingIndex] = toInsert

                //
                // insert item, ex: 99 after 2nd of x[0,1,2,3]
                // 1. allocate slice of size 2+1
                // 2. copy; so it's [0,1,2]
                // 3. put 99 at 3rd; so it's [0,1,99]
                // 4. append, [0,1,99] + x[2:], which is [2,3]
                // so it becomes [0,1,99,2,3]
                //
	            newList := make([]Stmt, found+1)
	            copy(newList,list[0:found])
	            newList[found] = toInsert
	            newList = append(newList, list[found:]...)

                v.parent.(*BlockStmt).List = newList
            }

        case *CallExpr:
            call := node.(*CallExpr)
            // fmt.Printf(">> v.node\n"); dump.Dump(v.node)
            if cc,ok := v.node.(*CaseClause); ok {
                // fmt.Printf(">> v.node\n"); dump.Dump(v.node)
                symtab := NewBindingSymTab()
                patternCode := callToStmt("m", "e", call, symtab)
                v.f.Decls = append(v.f.Decls,
                    NewMethod(fmt.Sprintf(`
                func __case_%d (e interface{}, b *pattern.Binding) (r bool) {
                        %s
                        r = true; return
                    }`,v.caseStmtCount,patternCode)))

                sw := v.switchStack.Last().(*SwitchStmt)
                newCall := makeMatchCall(v, sw.Tag, v.bindingIndex, v.caseStmtCount)
                v.caseStmtCount++
                cc.Values[0] = newCall
                s := ""
                for _,val := range symtab.table {
                    s = s + fmt.Sprintf(val, v.bindingIndex)
                }
                if s != "" {
                    newcc := append(NewStmtList(s), cc.Body...)
                    cc.Body = newcc
                }
                // dump.Dump2(cc.Body)
                // fmt.Printf("%s\n", newCall)
                // dump.Dump2(NewExpr(pattern))
            } else {
                if id,ok := call.Fun.(*Ident); ok {
                    if _,e := structs[id.Name]; e {
                        id.Name = "New" + id.Name
                    }
                }
            }
    }
    v.node = node
    w = v
    return
}

const (
    PROG_NAME = "Korat Golang Compiler"
    VER_MAJOR = 0
    VER_MINOR = 1
)

var ver = goopt.Flag([]string{"-v","--version"},nil,
                     "show version and exit",   "")

func NewBindingSymTab() *BindingSymTab {
    return &BindingSymTab{0, map[int]string{}}
}

func NewVisitor(astf *File) *VisitorImpl {
    return &VisitorImpl{astf, nil, nil,
            &vector.Vector{}, &vector.Vector{}, 0, map[int]Stmt{}, 0}
}

func StdExecve(argv []string, stopOnTrouble bool) (ok bool) {

    var err os.Error
    var cmd *exec.Cmd
    var pt int = exec.PassThrough
    var wmsg *os.Waitmsg
    ok = true
    cmd, err = exec.Run(argv[0], argv, os.Environ(), "", pt, pt, pt)

    if err != nil {
        if stopOnTrouble {
            log.Fatalf("[ERROR] %s\n", err)
        } else {
            log.Printf("[ERROR] %s\n", err)
        }
        ok = false

    } else {

        wmsg, err = cmd.Wait(0)

        if err != nil || wmsg.WaitStatus.ExitStatus() != 0 {

            if err != nil {
                log.Printf("[ERROR] %s\n", err)
            }

            if stopOnTrouble {
                os.Exit(1)
            }

            ok = false
        }
    }

    return ok
}

func main() {

    goopt.Version = fmt.Sprintf("%d.%d",VER_MAJOR,VER_MINOR)
    goopt.Summary = PROG_NAME
    goopt.Parse(nil)

    if *ver {
        fmt.Printf("\n%s version %d.%d",PROG_NAME,VER_MAJOR,VER_MINOR)
        fmt.Printf("\nCopyright (c) 2011 Chanwit Kaewkasi / SUT\n\n")
        return
    }

    var filename string = ""
    if len(goopt.Args) == 1 {
        filename = goopt.Args[0]
    } else {
        fmt.Print(goopt.Usage())
        return
    }

    fset := token.NewFileSet()
    astf,err := parser.ParseFile(fset, filename, nil, 0)
    if err == nil  {
        v := NewVisitor(astf)
        Walk(v, astf)
        tempfile,err := os.Open(filename + "k", os.O_WRONLY | os.O_CREATE, 0665)
        if err == nil {
            printer.Fprint(tempfile,fset,astf)
            tempfile.Close()
            newArgs := make([]string, len(os.Args))
            copy(newArgs, os.Args)
            for i,v := range newArgs {
                if v == filename {
                    newArgs[i] = filename + "k"
                }
            }
            newArgs[0] = os.Getenv("GOROOT") + "/bin/" + "8g"
            StdExecve(newArgs, true)
            os.Remove(filename + "k")
        }
    } else {
        fmt.Printf("%s\n", err)
    }
}
