package main

import (
	"fmt"
)

func main() {
	fmt.Println(func1())
}

func func1() string {
	defer func2()

	return func3()
}

func func2() string {
	fmt.Println("func2")
	return ""
}

func func3() string {
	fmt.Println("func3")
	return ""
}
