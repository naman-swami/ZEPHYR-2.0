package main
import (
	"fmt"
	"os"
	"path/filepath"
)
func main() {
	_ = os.MkdirAll("dist", 0755)
	for i := 1; i <= 10; i++ {
		_ = os.WriteFile(filepath.Join("dist", fmt.Sprintf("bundle_%02d.js", i)), []byte(fmt.Sprintf("console.log('bundle %d');", i)), 0644)
	}
}