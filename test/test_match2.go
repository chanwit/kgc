package main

import "korat/pattern"
import "reflect"

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
	left  Expr
	right Expr
}

func main() {
	input := NewMul(NewConst(1), NewConst(20))
	fmt.Printf("%s\n", input)
	__b1 := pattern.

		NewBinding()

	switch {
	case pattern.Match(input, &pattern.Node{
		reflect.TypeOf(&Mul{}),
		-1, nil, []*pattern.Node{&pattern.
			Node{reflect.TypeOf(&Const{}),
			-1, nil, []*pattern.Node{&pattern.
				Node{nil, -1, 10,

				nil}}},
			&pattern.Node{reflect.TypeOf(&Const{}),
				-1, nil,
				[]*pattern.Node{&pattern.Node{nil, -1,
					20, nil}}}}}, __b1):
		fmt.Printf("matched\n")
	case pattern.Match(input, &pattern.Node{
		reflect.TypeOf(&Mul{}),
		-1, nil, []*pattern.Node{&pattern.
			Node{reflect.TypeOf(&Const{}),
			-1, nil, []*pattern.Node{&pattern.
				Node{nil, 0, nil,

				nil}}},
			&pattern.Node{nil, 1, nil, nil},
		}}, __b1):
		x := __b1.Data[0].(int)
		y :=

			__b1.Data[1].(Expr)

		fmt.Printf("matched x=%v, y=%v\n", x, y)
	}

}
func NewConst(value int) (r *Const) {

	r = &Const{value}
	return
}
func (u *Const) Unapply() (s *pattern.Some) {

	s = &pattern.Some{u.value}
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
func (u *Var) Unapply() (s *pattern.Some) {

	s = &pattern.Some{u.name}
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
func (u *Mul) Unapply() (s *pattern.Some) {

	s = &pattern.Some{[]interface{}{u.left,
		u.
			right}}
	return
}
func (this *Mul) asType() Expr {

	return this

}
