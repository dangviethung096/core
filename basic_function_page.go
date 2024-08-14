package core

import "html/template"

func add(x int, y int) int {
	return x + y
}

func subtract(x int, y int) int {
	return x - y
}

func multiply(x int, y int) int {
	return x * y
}

func divide(x int, y int) int {
	return x / y
}

func seq(start int, end int) []int {
	result := make([]int, end-start+1)
	for i := range result {
		result[i] = start + i
	}
	return result
}

var basicFunctionMap = template.FuncMap{
	"add":      add,
	"subtract": subtract,
	"seq":      seq,
	"multiply": multiply,
	"divide":   divide,
}
