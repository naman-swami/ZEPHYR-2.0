package main

import "fmt"

func main() {
	fmt.Println("🚀 Playground App Initialized!")
	sum := Add(10, 20)
	fmt.Printf("10 + 20 = %d\n", sum)
	product := Multiply(6, 7)
	fmt.Printf("6 * 7 = %d\n", product)
	diff := Subtract(100, 35)
	fmt.Printf("100 - 35 = %d\n", diff)
}
