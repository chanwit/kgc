package main

import "fmt"
import "reflect"

type Expr interface {
    Cast() Expr
}

type Const struct {
    value int
}
type Mul struct {
    left  Expr
    right Expr
}

func main() {
    v := a.(type)
    fmt.Printf("%s\n", v)
}
