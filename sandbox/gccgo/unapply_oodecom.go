package main

import "os"
import "fmt"
import "time"
import "strconv"

type Expr interface {
    IsConst() bool
    IsVar()   bool
    IsMul()   bool

    Value()   int
    Name()    string
    Left()    Expr
    Right()   Expr
}

type ExprImpl struct {
    isConst bool
    isVar   bool
    isMul   bool

    value   int
    name    string
    left    Expr
    right   Expr
}

func (e *ExprImpl) IsConst() bool   { return e.isConst}
func (e *ExprImpl) IsVar()   bool   { return e.isVar  }
func (e *ExprImpl) IsMul()   bool   { return e.isMul  }
func (e *ExprImpl) Value()   int    { return e.value  }
func (e *ExprImpl) Name()    string { return e.name   }
func (e *ExprImpl) Left()    Expr   { return e.left   }
func (e *ExprImpl) Right()   Expr   { return e.right  }

type Const struct { *ExprImpl }
type Var   struct { *ExprImpl }
type Mul   struct { *ExprImpl }

func NewExpr() *ExprImpl {
    return &ExprImpl{isConst: false, isVar: false, isMul:false}
}
func NewConst(value int) Expr {
    e := NewExpr()
    e.value = value
    e.isConst = true
    return &Const{e}
}
func NewVar(name string) Expr {
    e := NewExpr()
    e.name = name
    e.isVar = true
    return &Var{e}
}
func NewMul(left, right Expr) Expr {
    e := NewExpr()
    e.left  = left
    e.right = right
    e.isMul = true
    return &Mul{e}
}

func Simplify(e Expr) Expr {
    if e.IsMul() {
        r,l := e.Right(),e.Left()
        switch {
            case l.IsConst() && l.Value() == 1: return r
            case r.IsConst() && r.Value() == 1: return l
            default: return e
        }
    }
    return e
}

func bench(input Expr, max int) {
    st := time.Nanoseconds()
    for i := 0; i < max; i++ {
        Simplify(input)
    }
    stop := time.Nanoseconds() - st
    fmt.Print(stop)
    fmt.Print("\n")
}

func main() {
    ROUNDS := 1000000
    if len(os.Args) > 1 {
        ROUNDS,_ = strconv.Atoi(os.Args[1])
    }
    bench(NewMul(NewConst(5), NewConst(1) ), ROUNDS) // second matched
    bench(NewMul(NewConst(1) ,NewConst(10)), ROUNDS) // first  matched
    bench(NewMul(NewConst(20),NewConst(20)), ROUNDS) // not    matched
}
