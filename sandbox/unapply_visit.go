package main

// import "dump"

type Visitor interface{
    otherwise(t Expr)   Expr
    caseMul(t *Mul)     Expr
    caseConst(t *Const) Expr
}

type VisitorImpl struct { }
func (this *VisitorImpl) otherwise(t Expr) Expr {
    panic("Match error")
}
func (this *VisitorImpl) caseMul(t *Mul) Expr {
    return this.otherwise(t)
}
func (this *VisitorImpl) caseConst(t *Const) Expr {
    return this.otherwise(t)
}

type Expr interface {
    matchWith(v Visitor)
}
type Mul struct {
    left Expr
    right Expr
}
type Const struct {
    value int
}

func (e *Mul)   matchWith(v Visitor) Expr { return v.caseMul(e)   }
func (e *Const) matchWith(v Visitor) Expr { return v.caseConst(e) }

type V1 struct {
    *VisitorImpl
}
type V2 struct {
    *VisitorImpl
}
func (this *V1) caseMul(m *Mul) Expr {
    return m.right.matchWith(&V2{})
}
func (this *V2) caseConst(c *Const) Expr {
    if c.value == 1 {
        return m.left
    }
    return e
}
func (this *V2) otherwise(e Expr) {
    return e
}

// func (e *Var)   matchWith(v Visitor) *Expr { return v.caseVar(e)   }

func NewExpr() *Expr {
    return &Expr{isConst: false, isVar: false, isMul:false}
}
func NewConst(value int) *Const {
    e := NewExpr()
    e.value = value
    e.isConst = true
    return &Const{e}
}
func NewVar(name string) *Var {
    e := NewExpr()
    e.name = name
    e.isVar = true
    return &Var{e}
}
func NewMul(left, right IExpr) *Mul {
    e := NewExpr()
    e.left  = left.Cast()
    e.right = right.Cast()
    e.isMul = true
    return &Mul{e}
}

func Simplify(e Expr) Expr {

}

func main() {
    e1 := NewMul(NewConst(2),NewConst(1))
    dump.Dump(e1.Simplify())
    e2 := NewMul(NewConst(1),NewConst(3))
    dump.Dump(e2.Simplify())
}
