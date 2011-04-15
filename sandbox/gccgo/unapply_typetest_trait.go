package main

import "fmt"

type Expr trait {}

type Mul casestruct borrows Expr {
    left  Expr
    right Expr
}
type Const casestruct borrows Expr {
    value int
}
type Var casestruct borrows Expr {
    name  string
}

func Simplify(e Expr) Expr {
    match e {
        case Mul(x,Const(1)): return x
    }

    if __m,ok := e.(*Mul); ok {
        b[0],m_right := __m.left,__m.right
        if __c,ok := m_right.(*Const); ok {
            _c_value := __c.value
            if _c_value == 1 {
                return true
            }
        }
    }
    return false

    if m,ok := e.(*Mul); ok {
        l,r := m.left, m.right
        if c,ok := r.(*Const); ok && c.value == 1 {
            return m.left
        }
        if c,ok := l.(*Const); ok && c.value == 1 {
            return m.right
        }
    }
    return e
}

func main() {
    e1 := Mul(Const(20),Const(1))
    for i := 1; i < 1000000; i++ {
        Simplify(e1)
    }
}
