package main

import "os"
import "fmt"
import "time"
import "strconv"

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

func Simplify(e Expr) Expr {
    match e {
        case Mul(Const(1), r): return r
        case Mul(l, Const(1)): return l
        default:               return e
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
    bench(Mul(Const(5), Const(1) ), ROUNDS) // second matched
    bench(Mul(Const(1) ,Const(10)), ROUNDS) // first  matched
    bench(Mul(Const(20),Const(20)), ROUNDS) // not    matched
}
