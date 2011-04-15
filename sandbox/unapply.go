package main_not

//import "fmt"
//import "dump"
import "reflect"

type Error string

func (e Error) String() string {
    return string(e)
}
var (
    ErrNotMatch      = Error("Extractable not matched")
    ErrValueNotMatch = Error("Value not matched")
)

type Some struct {
	value interface{}
}
var None = &Some{nil}

type Expr interface {
	Cast() Expr
}

type Const struct {
    value  int
}
type Var struct {
    name   string
}
type Mul struct {
    left   Expr
    right  Expr
}

type Extractable interface {
	Unapply() *Some
}

// func Const(value int) *Const {
//    return &Const{value}
// }

func (c *Const) Cast() Expr {
	return Expr(c)
}

func (c *Const) Unapply() *Some {
    return &Some{c.value}
}

func (m *Mul) Cast() Expr {
	return Expr(m)
}

func (m *Mul) Unapply() *Some {
    return &Some{[]interface{}{m.left, m.right}}
}

func (v *Var) Cast() Expr {
    return Expr(v)
}

func (v *Var) Unapply() *Some {
    return &Some{v.name}
}

type Node struct {
	typeName string
	bind	 int
    value    interface{}
    children []*Node
}

//
// Mul(Const(x),Const(y))
// input := &Mul{&Const{10},&Const{2}}
//
func matchValue(i interface{}, node *Node, binding map[int]interface{}) {
    if node.bind >= 0 {
        binding[node.bind] = i
        return
    }
    if i != node.value {
        panic(ErrValueNotMatch)
    }
    return
}

func init() {
    // register(&Mul{})
    // register(&Const{})
}

func _case(e Extractable, node *Node, binding map[int]interface{}) {
    // fmt.Printf("calling case\n")
    typ := reflect.Typeof(e)
    // fmt.Printf(">> e:\n")
    // dump.Dump(e)
    // fmt.Printf(">> node:\n")
    // dump.Dump(node)
    if typ.String() != node.typeName {
        panic(ErrNotMatch)
	}
    // fmt.Printf("===== Matched\n")
    if node.bind >= 0 { // perform binding and exit
        binding[node.bind] = e
        return
    }

    s := e.Unapply() // got *Some here

    // multiple child 
    if child, ok := s.value.([]interface{}); ok {
        for i,c := range child {
            ee,ext := c.(Extractable)
            if(ext) {
                _case(ee, node.children[i], binding)
            } else { // match value
                matchValue(c, node.children[i], binding)
            }
        }
    } else {
        ee,ext := s.value.(Extractable)
        if ext {
            _case(ee, node.children[0], binding)
        } else { // match value
            matchValue(s.value, node.children[0], binding)
        }
    }
	return
}

type Binding map[int]interface{}

func matchAndBind(s *Some, root *Node) (matched bool, binding Binding) {
    defer func() { if e := recover(); e != nil {
        matched = false
        binding = nil
    }}()
    binding = Binding{}
    if e,ok := s.value.(Extractable); ok {
        _case(e, root, binding)
        matched = true
        // dump.Dump(binding)
        return
    }
    return
}

func match(input Extractable, caseFunc func(*Some)interface{}) *Some {
	return &Some{caseFunc(&Some{input})}
}

// Mul(x,y)
var root1 = &Node{"*main.Mul",-1, nil,
    []*Node{&Node{"*main.Const", 0, nil, nil},
            &Node{"*main.Const", 1, nil, nil}}}

// Mul(Const(x),Const(y))
var root2 = &Node{"*main.Mul",  -1, nil,
            []*Node{&Node{"*main.Const",-1, nil,
                    []*Node{&Node{"int", 0, nil, nil}}},
                    &Node{"*main.Const",-1, nil,
                    []*Node{&Node{"int", 1, nil, nil}}}}}

// Mul(Const(10),Const(2))
var root3 = &Node{"*main.Mul",-1, nil,
            []*Node{&Node{"*main.Const",-1, nil,
                []*Node{&Node{"int", -1, 10,nil}}},
                    &Node{"*main.Const",-1, nil,
                []*Node{&Node{"int", -1, 2, nil}}}}}

// Mul(Const(1),x)
var case1 = &Node{"*main.Mul",-1, nil,[]*Node{

}}
// Mul(x,Const(1))
// var case2 =
// Mul(Const(0),x)
// Mul(x,Const(0))
// _

func main() {
    // input := Mul(Const(10),Const(3))
    input := &Mul{&Const{10},&Const{3}}
    /*
    a := match(input, func(e *Some) interface{} {
		return _case(e,"Mul(Const(x),Const(y))", func(x, y interface{}) interface{} {
			ax,_ := x.(int)
			ay,_ := y.(int)
			return &Const{ax * ay}
		})
    })
    */
    // for i:=0; i<10000000; i++ {
    matchAndBind(&Some{input},root1)
        //if !m1 { fmt.Printf("case#1 failed\n")}
    matchAndBind(&Some{input},root2)
        //if !m2 { fmt.Printf("case#1 failed\n")}
    matchAndBind(&Some{input},root3)
        //if m3  { fmt.Printf("case#1 failed\n")}
        //if m1 && m2 && !m3 {
        //    fmt.Printf("All passed\n")
        //}
    // }
	//  fmt.Printf("%s\n", reflect.Typeof(&Mul{}).String())
	//  dump.Dump(a)
	//	fmt.Printf("%s\n", a.value)
}
