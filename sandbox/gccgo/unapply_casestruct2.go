package main

import "korat/pattern"

import "os"
import "fmt"
import "time"
import "strconv"

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

var	__b1 *pattern.Binding

func Simplify(e Expr) Expr {
	switch {
	case __case_0(e, __b1):
		r := __b1.Data[0].(Expr)

		return r
	case __case_1(e, __b1):
		l := __b1.Data[0].(Expr)

		return l
	default:
		return e
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
    __b1 = pattern.NewBinding()
	ROUNDS := 1000000
	if len(os.Args) > 1 {
		ROUNDS, _ = strconv.Atoi(os.Args[1])
	}
	bench(NewMul(NewConst(5), NewConst(1)), ROUNDS)
	bench(NewMul(NewConst(1), NewConst(10)), ROUNDS)
	bench(NewMul(NewConst(20), NewConst(20)), ROUNDS)
}
func NewConst(value int) (r *Const) {

	r = &Const{value}
	return
}
func (u *Const) Unapply() (s *pattern.Some) {

	s = &pattern.Some{u.value,
	}
	return

}
func (this *Const) asType() Expr {

	return this

}
func NewVar(name string) (r *Var) {

	r = &Var{name}
	return
}
func (u *Var) Unapply() (s *pattern.Some) {

	s = &pattern.Some{u.name,
	}
	return

}
func (this *Var) asType() Expr {

	return this

}
func NewMul(left Expr, right Expr) (r *Mul) {

	r = &Mul{left, right}
	return

}
func (u *Mul) Unapply() (s *pattern.Some) {

	s = &pattern.Some{[]interface{}{u.left,
		u.
			right}}
	return
}
func (this *Mul) asType() Expr {

	return this

}
func __case_0(e interface{}, b *pattern.
	Binding) (r bool) {

	if m, ok := e.(*Mul); ok {
		c_left := m.left

		b.Data[0] = m.right
		if c0, ok := c_left.(*Const); ok {
			c_value := c0.value
			if c_value != 1 {
				r = false
				return
			}
		} else {
			r = false
			return
		}
	} else {
		r = false
		return
	}
	r = true
	return
}
func __case_1(e interface{}, b *pattern.
	Binding) (r bool) {

	if m, ok := e.(*Mul); ok {
		b.Data[0] = m.left

		c_right := m.right
		if c1, ok := c_right.(*Const); ok {
			c_value := c1.value
			if c_value !=
				1 {
				r = false
				return
			}
		} else {

			r = false
			return
		}
	} else {
		r = false
		return
	}
	r = true
	return
}
