package main

import (
	"fmt"

	"golang.org/x/example/hello/reverse"
)

func main() {
	sourceStr := "Hello, OTUS!"
	reversedStr := Reverse(sourceStr)
	fmt.Print(reversedStr)
}

// Reverse - выполняет переворот строки.
func Reverse(s string) string {
	return reverse.String(s)
}
