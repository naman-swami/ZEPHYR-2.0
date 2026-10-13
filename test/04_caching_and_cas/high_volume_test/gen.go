package main
import (
	"fmt"
	"os"
	"path/filepath"
)
func main() {
	_ = os.MkdirAll("build", 0755)
	for i := 1; i <= 50; i++ {
		_ = os.WriteFile(filepath.Join("build", fmt.Sprintf("file_%03d.txt", i)), []byte(fmt.Sprintf("DATA_%03d", i)), 0644)
	}
}