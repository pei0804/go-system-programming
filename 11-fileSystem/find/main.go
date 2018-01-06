package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	fmt.Println(filepath.Match("image-*.png", "image-100.png"))
	// true <nil>

	files, err := filepath.Glob("./*.png")
	if err != nil {
		panic(err)
	}
	fmt.Println(files)
	// [golang.png image-100.png]
}
