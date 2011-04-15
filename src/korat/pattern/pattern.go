package pattern

import "reflect"
// import "fmt"
// import "dump"

type Error string

func (e Error) String() string {
    return string(e)
}
var (
    ErrNotMatch      = Error("Extractable not matched")
    ErrValueNotMatch = Error("Value not matched")
)

type Some struct {
	Value interface{}
}
var None = &Some{nil}

type Extractable interface {
	Unapply() *Some
}

type Node struct {
	TypeName reflect.Type
	Bind	 int
    Value    interface{}
    Children []*Node
}

type Binding struct {
    Data map[int]interface{}
}

func NewBinding() *Binding{
    return &Binding{map[int]interface{}{}}
}

//
// Mul(Const(x),Const(y))
// input := &Mul{&Const{10},&Const{2}}
//
func matchValue(i interface{}, node *Node, binding *Binding) {
    // fmt.Printf("match Value\n")
    // fmt.Printf(">> binding: %v\n", binding)
    // fmt.Printf(">> node:    %v\n", node)
    if node.Bind >= 0 {
        binding.Data[node.Bind] = i
        // fmt.Printf("+> binding: %v\n", binding)
        return
    }
    if i != node.Value {
        panic(ErrValueNotMatch)
    }
    return
}

func _case(e Extractable, node *Node, binding *Binding) {
    // fmt.Printf("calling case\n")
    typ := reflect.Typeof(e)
    // fmt.Printf(">> e: %v\n", typ.String())
    // dump.Dump(e)
    // fmt.Printf(">> node: %v\n", node.TypeName)
    // fmt.Printf(">> binding: %v\n", binding)
    // dump.Dump(node)
    // if typ != node.TypeName {
    //     panic(ErrNotMatch)
	// }
    // fmt.Printf("===== Matched\n")
    if node.Bind >= 0 { // perform binding and exit
        binding.Data[node.Bind] = e
        // fmt.Printf("+> binding: %v\n", binding)
        return
    }
    if typ != node.TypeName {
        panic(ErrNotMatch)
	}

    s := e.Unapply() // got *Some here

    // multiple child 
    if child, ok := s.Value.([]interface{}); ok {
        for i,c := range child {
            ee,ext := c.(Extractable)
            if(ext) {
                _case(ee, node.Children[i], binding)
            } else { // match value
                matchValue(c, node.Children[i], binding)
            }
        }
    } else {
        ee,ext := s.Value.(Extractable)
        if ext {
            _case(ee, node.Children[0], binding)
        } else { // match value
            matchValue(s.Value, node.Children[0], binding)
        }
    }
	return
}

//
// case ok,b := MatchAndBind(input, &Node{}); x,y := b[0],b[1];
//
func Match(v interface{}, root *Node, binding *Binding) (matched bool) {
    defer func() { if e := recover(); e != nil {
        matched = false
        binding = nil
        // fmt.Printf("%s\n", e)
    }}()
    //binding.Data = map[int]interface{}{}
    if e,ok := v.(Extractable); ok {
        _case(e, root, binding)
        matched = true
        // dump.Dump(binding)
        return
    }
    return
}

