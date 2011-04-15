package main_new

import "fmt"
import "korat/ast"
import "strconv"
import "korat/parser"
import "korat/token"
// import "dump"

/*
    if __m,ok := e.(*Mul); ok {
        x,m_right := __m.left,__m.right
        if __c,ok := m_right.(*Const); ok {
            _c_value := __c.value
            if _c_value == 1 {
                b.data[0] = x
                return true
            }
        }
    }
*/
func callToStmt(varname string, root string, c *ast.CallExpr) (r string) {
    r = ""
    r = r + fmt.Sprintf("if %s,ok := %s.(*%s); ok {\n",varname, root, c.Fun.(*ast.Ident).Name)
    subPrefix := "c"
    rgen, lhs := genAssign(subPrefix, c, varname, c.Args)
    r = r + rgen
    for i := 0; i < len(c.Args); i++ {
        a := c.Args[i]
        if cA,ok := a.(*ast.CallExpr); ok {
            r = r + callToStmt("c" + strconv.Itoa(i),lhs[i],cA)
        } else if cV,ok := a.(*ast.BasicLit); ok {
            r = r + fmt.Sprintf("if %s != %s { return false }\n",
                lhs[i], string(cV.Value))
        }
    }
    r = r + fmt.Sprint("} else {\n")
    r = r + fmt.Sprintf("  return false\n")
    r = r + fmt.Sprintf("}\n")
    return
}


//
// mock function
//
func lookupField(structName string, fieldIndex int) string {
    if structName == "Mul"   && fieldIndex == 0 { return "left"  }
    if structName == "Mul"   && fieldIndex == 1 { return "right" }
    // if structName == "Mul"   && fieldIndex == 2 { return "next"  }
    if structName == "Const" && fieldIndex == 0 { return "value" }
    return ""
}

func join(strs []string, delim string) (r string) {
    r = strs[0]
    for i := 1; i<len(strs); i++ {
        r = r + delim + strs[i]
    }
    return
}

func getBindingIndexOf(s string) string {
    if s == "x" { return "0" }
    if s == "y" { return "1" }
    return ""
}

func genAssign(prefix string, c *ast.CallExpr, varname string, args []ast.Expr) (r string, lhs []string) {
    lhs = make([]string, len(args))
    rhs := make([]string, len(args))
    for i:=0; i<len(args); i++ {
        a := args[i]
        fname := lookupField(c.Fun.(*ast.Ident).Name, i)
        assignOp := ":="
        if id,ok := a.(*ast.Ident); ok {
            lhs[i] = "b.data[" + getBindingIndexOf(id.Name) + "]" // 
            assignOp = "="
        } else if _,ok := a.(*ast.CallExpr); ok {
            lhs[i] = prefix + "_" + fname
        } else if _,ok := a.(*ast.BasicLit); ok {
            lhs[i] = prefix + "_" + fname
        }
        rhs[i] = varname + "." + fname
        r = r + fmt.Sprintf("%s %s %s\n", lhs[i], assignOp, rhs[i])
    }
    // r = fmt.Sprintf("%s := %s\n", join(lhs,","), join(rhs,","))
    return
}

func main() {
/*
    p1 := &ast.CallExpr {
  Fun: &ast.Ident {
    NamePos: 385,
    Name: "Mul",
    Obj: nil },
  Lparen: 388,
  Args: []ast.Expr {
    &ast.CallExpr {
      Fun: &ast.Ident {
        NamePos: 389,
        Name: "Const",
        Obj: nil },
      Lparen: 394,
      Args: []ast.Expr {
        &ast.BasicLit {
          ValuePos: 395,
          Kind: token.INT,
          Value: []uint8 {
            49,
            48 } } },
      Ellipsis: 0,
      Rparen: 397 },
    &ast.CallExpr {
      Fun: &ast.Ident {
        NamePos: 399,
        Name: "Const",
        Obj: nil },
      Lparen: 404,
      Args: []ast.Expr {
        &ast.BasicLit {
          ValuePos: 405,
          Kind: token.INT,
          Value: []uint8 {
            50,
            48 } } },
      Ellipsis: 0,
      Rparen: 407 } },
  Ellipsis: 0,
  Rparen: 408 }
*/

    // fmt.Printf("%s \n", 
    c,_ := parser.ParseExpr(token.NewFileSet(),"temp.go", "Mul(Const(1),x)")
    p := c.(*ast.CallExpr)
    fmt.Print(callToStmt("m", "e", p))
    c2,_ := parser.ParseExpr(token.NewFileSet(),"temp.go", "Mul(x,Const(1))")
    p2 := c2.(*ast.CallExpr)
    callToStmt("m", "e", p2)
    // fmt.Printf"%s \n", p2)
}

