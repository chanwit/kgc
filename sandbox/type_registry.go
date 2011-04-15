package main

import "reflect"

var nameToConcreteType = map[string]reflect.Type{}
var concreteTypeToName = map[reflect.Type]string{}

func registerName(name string, value interface{}) {
    if name == "" {
	    panic("attempt to register empty name")
    }
    // base := userType(reflect.Typeof(value)).base
    base := reflect.Typeof(value)
    // Check for incompatible duplicates.
    if t, ok := nameToConcreteType[name]; ok && t != base {
        return
    }
    if n, ok := concreteTypeToName[base]; ok && n != name {
        return
    }
    // Store the name and type provided by the user....
    nameToConcreteType[name] = reflect.Typeof(value)
    // but the flattened type in the type table, since that's what decode needs.
    concreteTypeToName[base] = name
}

func register(value interface{}) {
	// Default to printed representation for unnamed types
	rt := reflect.Typeof(value)
	name := rt.String()

	// But for named types (or pointers to them), qualify with import path.
	// Dereference one pointer looking for a named type.
	star := ""
	if rt.Name() == "" {
		if pt, ok := rt.(*reflect.PtrType); ok {
			star = "*"
			rt = pt
		}
	}
	if rt.Name() != "" {
		if rt.PkgPath() == "" {
			name = star + rt.Name()
		} else {
			name = star + rt.PkgPath() + "." + rt.Name()
		}
	}

	registerName(name, value)
}


