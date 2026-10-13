package main
import (
	"fmt"
	"os"
	"strconv"
)
func main() {
	countFile := "attempt.txt"
	count := 0
	b, err := os.ReadFile(countFile)
	if err == nil {
		c, _ := strconv.Atoi(string(b))
		count = c
	}
	count++
	_ = os.WriteFile(countFile, []byte(fmt.Sprintf("%d", count)), 0644)
	if count < 3 {
		fmt.Printf("Simulated transient failure on attempt %d\n", count)
		os.Exit(1)
	}
	fmt.Printf("Success on attempt %d!\n", count)
	os.Exit(0)
}