package dump

import (
    r "reflect"
    "fmt"
    "strconv"
    "io"
    "os"
)

var emptyString = ""

func Fdump(out io.Writer, v_ interface{}) {
    // forward decl
    var dump0 func(r.Value, int)
    var dump func(r.Value, int, *string, *string)

    done := make(map[string]bool)

    dump = func(v r.Value, d int, prefix *string, suffix *string) {
        pad := func() {
            res := ""
            for i := 0; i < d; i++ {
                res += "  "
            }
            fmt.Fprintf(out, res)
        }

        padprefix := func() {
            if prefix != nil {
                fmt.Fprintf(out, *prefix)
            } else {
                res := ""
                for i := 0; i < d; i++ {
                    res += "  "
                }
                fmt.Fprintf(out, res)
            }
        }

        printv := func(o interface{}) { fmt.Fprintf(out, "%v", o) }

        printf := func(s string, args ...interface{}) { fmt.Fprintf(out, s, args...) }

        // prevent circular for composite types
        switch v.Kind() {
        // case nil:
        // do nothing
        case r.Array, r.Slice, r.Map, r.Ptr, r.Struct, r.Interface:
            addr := v.Addr()
            key := fmt.Sprintf("%x %v", addr, v.Type())
            if _, exists := done[key]; exists {
                padprefix()
                printf("<%s>", key)
                return
            } else {
                done[key] = true
            }
        default:
            // do nothing
        }

        switch v.Kind() {
        case r.Array:
            padprefix()
            printf("[%d]%s {\n", v.Len(), v.Type().Elem())
            for i := 0; i < v.Len(); i++ {
                dump0(v.Index(i), d+1)
                if i != v.Len()-1 {
                    printf(",\n")
                }
            }
            print("\n")
            pad()
            print("}")

        case r.Slice:
            padprefix()
            if v.Len()==0 {
                printf("[]%s (len=%d)", v.Type().Elem(), v.Len())
            } else {
                printf("[]%s (len=%d) {\n", v.Type().Elem(), v.Len())
                for i := 0; i < v.Len(); i++ {
                    dump0(v.Index(i), d+1)
                    if i != v.Len()-1 {
                        printf(",\n")
                    }
                }
                print("\n")
                pad()
                print("}")
            }

        case r.Map:
            padprefix()
            t := v.Type()
            printf("map[%s]%s {\n", t.Key(), t.Elem())
            for i, k := range v.MapKeys() {
                dump0(k, d+1)
                printf(": ")
                dump(v.MapIndex(k), d+1, &emptyString, nil)
                if i != v.Len()-1 {
                    printf(",\n")
                }
            }
            print("\n")
            pad()
            print("}")

        case r.Ptr:
            padprefix()
            if v.Elem().IsNil() {
                printf("(*%s) nil", v.Type().Elem())
            } else {
                print("&")
                dump(v.Elem(), d, &emptyString, nil)
            }

        case r.Struct:
            padprefix()
            t := v.Type()
            printf("%s {\n", t)
            d += 1
            for i := 0; i < v.NumField(); i++ {
                pad()
                printv(t.Field(i).Name)
                printv(": ")
                dump(v.Field(i), d, &emptyString, nil)
                if i != v.NumField()-1 {
                    printf(",\n")
                }
            }
            d -= 1
            print("\n")
            pad()
            print("}")

        case r.Interface:
            padprefix()
            t := v.Type()
            printf("(%s) ", t)
            dump(v.Elem(), d, &emptyString, nil)

        case r.String:
            padprefix()
            printv(strconv.Quote(v.String()))

        case r.Bool,
            r.Int,
            r.Int8,
            r.Int16,
            r.Int32,
            r.Int64,
            r.Uint,
            r.Uint8,
            r.Uint16,
            r.Uint32,
            r.Uint64,
            r.Uintptr,
            r.Float32,
            r.Float64,
            r.Complex64,
            r.Complex128:
            padprefix()
            //printv(o.Interface());
            i := v.Interface()
            if stringer, ok := i.(interface {
                String() string
            }); ok {
                printf("(%v) %s", v.Type(), stringer.String())
            } else {
                printv(i)
            }

        default:
            padprefix()
            if v.IsNil() {
                printv("nil")
            } else {
                printf("(%v) %v", v.Type(), v.Interface())
            }            
        }
    }

    dump0 = func(v r.Value, d int) { dump(v, d, nil, nil) }

    v := r.ValueOf(v_)
    dump0(v, 0)
    fmt.Fprintf(out, "\n")
}

// Prints to the writer the value with indentation.
func Fdump2(out io.Writer, v_ interface{}, linedem string) {
    // forward decl
    var dump0 func(r.Value, int)
    var dump func(r.Value, int, *string, *string)

    done := make(map[string]bool)

    dump = func(v r.Value, d int, prefix *string, suffix *string) {
        pad := func() {
            res := ""
            for i := 0; i < d; i++ {
                res += "  "
            }
            fmt.Fprintf(out, res)
        }

        padprefix := func() {
            if prefix != nil {
                fmt.Fprintf(out, *prefix)
            } else {
                res := ""
                for i := 0; i < d; i++ {
                    res += "  "
                }
                fmt.Fprintf(out, res)
            }
        }

        printv := func(o interface{}) { fmt.Fprintf(out, "%v", o) }

        printf := func(s string, args ...interface{}) { fmt.Fprintf(out, s, args...) }

        // prevent circular for composite types
        switch v.Kind() {
        // case nil:
            // do nothing
        case r.Array, r.Slice,
             r.Map, r.Ptr,
             r.Struct, r.Interface:
            addr := v.Addr()
            key := fmt.Sprintf("%x %v", addr, v.Type())
            if _, exists := done[key]; exists {
                padprefix()
                printf("<%s>", key)
                return
            } else {
                done[key] = true
            }
        default:
            // do nothing
        }

        switch v.Kind() {
        case r.Array:
            padprefix()
            printf("[%d]%s {\n", v.Len(), v.Type().Elem())
            for i := 0; i < v.Len(); i++ {
                dump0(v.Index(i), d+1)
                if i != v.Len()-1 {
                    printf(",\n")
                }
            }
            print(linedem)
            pad()
            print("}")

        case r.Slice:
            padprefix()
            if v.Len()==0 {
                printf("[]%s{}", v.Type().Elem())
            } else {
                printf("[]%s {\n", v.Type().Elem())
                for i := 0; i < v.Len(); i++ {
                    dump0(v.Index(i), d+1)
                    if i != v.Len()-1 {
                        printf(",\n")
                    }
                }
                print(linedem)
                print(" }")
            }

        case r.Map:
            padprefix()
            t := v.Type()
            printf("map[%s]%s {\n", t.Key(), t.Elem())
            for i, k := range v.MapKeys() {
                dump0(k, d+1)
                printf(": ")
                dump(v.MapIndex(k), d+1, &emptyString, nil)
                if i != v.Len()-1 {
                    printf(",\n")
                }
            }
            print(linedem)
            print(" }")

        case r.Ptr:
            padprefix()
            if v.Elem().IsNil() {
                printf("nil") // , o.Type().(*r.PtrType).Elem())
            } else {
                print("&")
                dump(v.Elem(), d, &emptyString, nil)
            }

        case r.Struct:
            padprefix()
            t := v.Type()
            printf("%s {\n", t)
            d += 1
            for i := 0; i < v.NumField(); i++ {
                pad()
                printv(t.Field(i).Name)
                printv(": ")
                dump(v.Field(i), d, &emptyString, nil)
                if i != v.NumField()-1 {
                    printf(",\n")
                }
            }
            d -= 1
            print(linedem)
            print(" }")

        case r.Interface:
            padprefix()
            dump(v.Elem(), d, &emptyString, nil)

        case r.String:
            padprefix()
            printv(strconv.Quote(v.String()))

        case r.Bool,
            r.Int,
            r.Int8,
            r.Int16,
            r.Int32,
            r.Int64,
            r.Uint,
            r.Uint8,
            r.Uint16,
            r.Uint32,
            r.Uint64,
            r.Uintptr,
            r.Float32,
            r.Float64,
            r.Complex64,
            r.Complex128:
            padprefix()
            i := v.Interface()
            if stringer, ok := i.(interface {
                String() string
            }); ok {
                printf("%s", v.Type(), stringer.String())
            } else {
                printv(i)
            }

        default:            
            padprefix()
            if v.IsNil() {
                printv("nil")
            } else {
                printf("%v", v.Interface())
            }
        }
    }

    dump0 = func(v r.Value, d int) { dump(v, d, nil, nil) }

    v := r.ValueOf(v_)
    dump0(v, 0)
    fmt.Fprintf(out, "\n")
}

// Print to standard out the value that is passed as the argument with indentation.
// Pointers are dereferenced.
func Dump2(v_ interface{}) { Fdump2(os.Stdout, v_,"") }
func Dump (v_ interface{}) { Fdump (os.Stdout, v_)    }

