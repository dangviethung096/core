package core

import "html/template"

func add(x int, y int) int {
	return x + y
}

func subtract(x int, y int) int {
	return x - y
}

var basicFunctionMap = template.FuncMap{
	"add":      add,
	"subtract": subtract,
}
