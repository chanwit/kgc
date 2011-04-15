// © Knug Industries 2010 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package global

// Be a man and use some globals.

// This is not threadsafe, unless maps are, i.e. it's not
// Since I only use the globals in a single thread, I did
// not bother with the 'sync' package Lock/Unlock..

// The idea of this package is to have 'global' variables
// naturally, but also to use it as some sort of multi-map. 
// I.e. a map that can hold different types, naturally we 
// use one map for each type, but still from a callers 
// viewpoint, it's all: global.SetXxx and global.GetXxx

var intMap map[string]int
var stringMap map[string]string
var float64Map map[string]float64
var float32Map map[string]float32
var boolMap map[string]bool
var interfaceMap map[string]interface{}


func init() {
    intMap = make(map[string]int)
    stringMap = make(map[string]string)
    float64Map = make(map[string]float64)
    float32Map = make(map[string]float32)
    boolMap = make(map[string]bool)
    interfaceMap = make(map[string]interface{})
}

// setters

func SetInt(key string, value int) {
    intMap[key] = value
}

func SetString(key, value string) {
    stringMap[key] = value
}

func SetFloat64(key string, value float64) {
    float64Map[key] = value
}

func SetFloat32(key string, value float32) {
    float32Map[key] = value
}

func SetBool(key string, value bool) {
    boolMap[key] = value
}

func SetInterface(key string, value interface{}) {
    interfaceMap[key] = value
}

// getters

func GetIntSafe(key string) (value int, ok bool) {
    value, ok = intMap[key]
    return value, ok
}

func GetInt(key string) int {
    return intMap[key]
}

func GetStringSafe(key string) (value string, ok bool) {
    value, ok = stringMap[key]
    return value, ok
}

func GetString(key string) string {
    return stringMap[key]
}

func GetFloat64Safe(key string) (value float64, ok bool) {
    value, ok = float64Map[key]
    return value, ok
}

func GetFloat64(key string) float64 {
    return float64Map[key]
}

func GetFloat32Safe(key string) (value float32, ok bool) {
    value, ok = float32Map[key]
    return value, ok
}

func GetFloat32(key string) float32 {
    return float32Map[key]
}

func GetBoolSafe(key string) (value, ok bool) {
    value, ok = boolMap[key]
    return value, ok
}

func GetBool(key string) bool {
    return boolMap[key]
}
