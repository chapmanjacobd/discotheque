package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	fPath := "/home/user/disco/sync"
	relativePath := "../../etc/passwd"
	joined := filepath.Join(fPath, relativePath)
	fmt.Printf("Base: %s\n", fPath)
	fmt.Printf("Relative: %s\n", relativePath)
	fmt.Printf("Joined: %s\n", joined)
}
