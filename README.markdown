README
======

Korat Golang Compiler, KGC.
Copyright (c) 2011 Chanwit Kaewkasi / SUT.

This program is a front-end compiler for 8g and GCCGO.
It additionally supports the following language constructs:

  * trait
  * case struct
  * pattern matching

Trait
-----

    type Object trait {}


Case struct
-----------

    type Something  casestruct {
        left  Object
        right Object
    }
    type SomeObject casestruct borrows Object {}


Pattern Matching
----------------

    match s {
        case Something(x, SomeObject(y)):
            return Some(x,y)
        case Something(SomeObject(), x):
            return Some(x)
    }

