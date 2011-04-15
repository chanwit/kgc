package main

import "fmt"

type Some struct {
    value interface{}
}

type Expr trait {}

type Const casestruct borrows Expr {
    value  int
}
type Var casestruct borrows Expr {
    name   string
}
type Mul casestruct borrows Expr {
    left   Expr
    right  Expr
}

func main() {
    input := Mul(Const(10),Const(20))
    fmt.Printf("%s\n", input)
}
