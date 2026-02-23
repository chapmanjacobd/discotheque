package main
import (
	"fmt"
	"path/filepath"
)
func main() {
	p := ".hidden"
	fmt.Printf("Base: %s, Ext: %s\n", filepath.Base(p), filepath.Ext(p))
}
