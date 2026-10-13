package main
import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
)
func main() {
	_ = os.MkdirAll("bin", 0755)
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	_ = os.WriteFile(filepath.Join("bin", "rand.bin"), []byte(hex.EncodeToString(b)), 0644)
}