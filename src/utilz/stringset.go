// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package stringset

import "strings"


// Use the built-in hash tables to construct a set of strings.


type StringSet struct {
    elements map[string]interface{}
}

func New() *StringSet {
    s := new(StringSet)
    s.elements = make(map[string]interface{})
    return s
}

func (s *StringSet) Add(e string) bool {
    if s.Contains(e) {
        return false
    }
    s.elements[e] = nil
    return true
}

func (s *StringSet) Clear() {
    s.elements = make(map[string]interface{})
}

func (s *StringSet) Remove(e string) bool {
    if !s.Contains(e) {
        return false
    }
    s.elements[e] = nil, false
    return true
}

func (s *StringSet) Contains(e string) bool {
    _, here := s.elements[e]
    return here
}

func (s *StringSet) Len() int {
    return len(s.elements)
}

func (s *StringSet) iterate(c chan<- string) {
    for k, _ := range s.elements {
        c <- k
    }
    close(c)
}

func (s *StringSet) Iter() <-chan string {
    c := make(chan string)
    go s.iterate(c)
    return c
}

func (s *StringSet) String() string {
    sarray := make([]string, len(s.elements)+2)
    var i int = 1
    for e, _ := range s.elements {
        sarray[i] = e
        i++
    }
    sarray[0] = "["
    sarray[i] = "]"
    return strings.Join(sarray, " ")
}

func (s *StringSet) Slice() []string {
    i := 0
    slice := make([]string, len(s.elements))
    for k,_ := range s.elements {
        slice[i] = k
        i++
    }
    return slice
}
