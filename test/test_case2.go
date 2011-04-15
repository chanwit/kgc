package main

import "fmt"

type Some struct {
	value interface{}
}

type Expr interface {
	asType() Expr
}

type Const struct {
	value int
}
type Var struct {
	name string
}
type Mul struct {
	left	Expr
	right	Expr
}

func main() {
	input := NewMul(NewConst(10), NewConst(20))
	fmt.Printf("%s\n", input)
}
func NewConst(value int) (r *Const) {

	r = &Const{value}
	return
}
func (u *Const) Unapply() (s *Some) {

	s = &Some{u.value}
	return

}
func (this *Const) asType() Expr {
	return this
}
func NewVar(name string) (r *Var) {

	r =
		&Var{name}
	return
}
func (u *Var) Unapply() (s *Some) {

	s = &Some{u.name}
	return

}
func (this *Var) asType() Expr {
	return this
}
func NewMul(left Expr,
right Expr) (r *Mul) {

	r = &Mul{left, right}
	return

}
func (u *Mul) Unapply() (s *Some) {

	s = &Some{[]interface{}{u.left, u.
		right}}
	return

}
func (this *Mul) asType() Expr {
	return this
}
