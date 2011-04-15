package main

import "fmt"

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

type Binding struct {
    data map[int]interface{}
}

var b = &Binding{}

// Mul(x,Const(1))
func matchCase01(e Expr) bool {
    b.data = map[int]interface{}{}
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
    return false
}

// Mul(Const(1),y)
func matchCase02(e Expr) bool {
    b.data = map[int]interface{}{}
    if __m,ok := e.(*Mul); ok {
        m_left,y := __m.left,__m.right
        if __c,ok := m_left.(*Const); ok {
            _c_value := __c.value
            if _c_value == 1 {
                b.data[0] = y
                return true
            }
        }
    }
    return false
}

func case01 (e Expr, b *Binding) bool {
    if m,ok := e.(*Mul); ok {
        c_left :=   m.left
        b.data[0] = m.right
        if c0,ok := c_left.(*Const); ok {
            c_value := c0.value
            if c_value != 1 { return false }
        } else {
            return false
        }
    } else {
        return false
    }
    return true
}

func case02 (e Expr, b *Binding) bool {
    if m,ok := e.(*Mul); ok {
        b.data[0] = m.left
        c_right :=  m.right
        if c1,ok := c_right.(*Const); ok {
            c_value := c1.value
            if c_value != 1 { return false }
        } else {
            return false
        }
    } else {
        return false
    }
    return true
}

func Simplify(e Expr) Expr {
    b.data = map[int]interface{}{}
    switch {
        case case01(e, b): x := b.data[0].(Expr); return x
        case case02(e, b): y := b.data[0].(Expr); return y
    }
    return e
}

func main() {
    e1 := &Mul{&Const{1},&Const{10}}
    for i := 1; i < 1000000; i++ {
        Simplify(e1)
    }
    fmt.Printf("%v\n", Simplify(e1))
}
