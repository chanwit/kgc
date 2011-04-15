package main

import "fmt"

type Expr interface {
	a()
	B()
	asType() Expr
}

func _ExprImpl_a(this Expr) {
	fmt.Printf("a() %s\n", this)
}

func _ExprImpl_B(this Expr) {
	fmt.Printf("B() %s\n", this)
}

type Mul struct {
	value, x	int
	y, z		string
}

func main() {
	a := NewMul(10, 20, "a", "b")
	a.a()
}
func NewMul(value,

x int, y, z string) (r *Mul) {
	r = &Mul{value,
		x,
		y,
		z}
	return
}
func (this *Mul) a() {
	_ExprImpl_a(this)
}
func (this *Mul) B() {
	_ExprImpl_B(this)
}
func (this *Mul) asType() Expr {
	return this
}
