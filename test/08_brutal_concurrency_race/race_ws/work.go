package main
import (
	"os"
	"time"
)
func main() {
	time.Sleep(100 * time.Millisecond)
	_ = os.WriteFile("output.dat", []byte("CONCURRENT_HEAVY_PAYLOAD_DATA"), 0644)
}