package main

import "fmt"

type Expr trait {
    a()
    B()
}

func Expr.a() {
    fmt.Printf("a() %s\n", this)
}

func Expr.B() {
    fmt.Printf("B() %s\n", this)
}

type Mul casestruct borrows Expr {
    value,x int
    y,z string
}

func main() {
    a := Mul(10,20,"a","b")
    a.a()
}

