package main

import "os"
import "fmt"
import "time"
import "strconv"

type Expr interface { Cast() Expr }

type Mul struct {
    left  Expr
    right Expr
}
type Const struct {
    value int
}
type Var struct {
    name  string
}

func (m *Mul)   Cast() Expr { return m }
func (c *Const) Cast() Expr { return c }
func (v *Var)   Cast() Expr { return v }

func Simplify(e Expr) Expr {
    if m,ok := e.(*Mul); ok {
        l,r := m.left, m.right
        if c,ok := l.(*Const); ok && c.value == 1 {
            return m.right
        }
        if c,ok := r.(*Const); ok && c.value == 1 {
            return m.left
        }
    }
    return e
}

func NewMul(l, r Expr) Expr { return &Mul{l, r} }
func NewConst(v int)   Expr { return &Const{v}  }
func NewVar(n string)  Expr { return &Var{n}    }

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
