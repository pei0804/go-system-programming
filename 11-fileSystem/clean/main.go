package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	// 無駄なパス表記を消す
	fmt.Println(filepath.Clean("./path/filepath/../path.go"))
	// path/path.go

	// 絶対パス
	abspath, _ := filepath.Abs("path/filepath/path_unix.go")
	fmt.Println(abspath)
	// /usr/local/go/src/path/ilepath/path_unix.go

	// 基準のパスから相対パスを出す
	relpath, _ := filepath.Rel("/usr/local/go/src", "/usr/local/go/src/path/filepath/path.go")
	fmt.Println(relpath)
	// path/filepath/path.go
}
