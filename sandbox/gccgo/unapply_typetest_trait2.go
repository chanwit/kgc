package main

import "korat/pattern"

type Expr interface {
	asType() Expr
}

type Mul struct {
	left	Expr
	right	Expr
}
type Const struct {
	value int
}
type Var struct {
	name string
}

func Simplify(e Expr) Expr {
	if m, ok := e.(*Mul); ok {
		l, r := m.left, m.right
		if c, ok := r.(*Const); ok && c.value == 1 {
			return m.left
		}
		if c, ok := l.(*Const); ok && c.value == 1 {
			return m.right
		}
	}
	return e
}

func main() {
	e1 := NewMul(NewConst(20), NewConst(1))
	for i := 1; i < 1000000; i++ {
		Simplify(e1)
	}
	// fmt.Print(reflect.Typeof(e1))
}
func NewMul(left Expr,
right Expr) (r *Mul) {

	r = &Mul{left, right,
	}
	return
}
func (u *Mul) Unapply() (s *pattern.Some) {

	s = &pattern.

		Some{[]interface{}{u.left, u.right},
	}
	return

}
func (this *Mul) asType() Expr {

	return this

}
func NewConst(value int) (r *Const) {
	r = &Const{value}
	return

}
func (u *Const) Unapply() (s *pattern.Some) {

	s = &pattern.

		Some{u.value}
	return

}
func (this *Const) asType() Expr {

	return this

}
func NewVar(name string) (r *Var) {
	r = &Var{
		name}
	return

}
func (u *Var) Unapply() (s *pattern.Some) {

	s = &pattern.

		Some{u.name}
	return

}
func (this *Var) asType() Expr {

	return this

}
