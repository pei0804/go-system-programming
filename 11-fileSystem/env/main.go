package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

func main() {
	path := os.ExpandEnv("${GOPATH}/src/github.com/pei0804/go-system-programming")
	fmt.Println(path)
	// /Users/jumpei/go/src/github.com/pei0804/go-system-programming

	// ~はOSではなくシェルが提供しているものなので、OSが解釈出来る形にする
	my, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Println(my.HomeDir)
	//  /Users/jumpei

	fmt.Println(pathClean("~/bin/goenv"))
	// /Users/jumpei/bin/goenv
}

// ~ を置換、環境変数の展開、パスのクリーン化
func pathClean(path string) string {
	if len(path) > 1 && path[0:2] == "~/" {
		my, err := user.Current()
		if err != nil {
			panic(err)
		}
		path = my.HomeDir + path[1:]
	}
	path = os.ExpandEnv(path)
	return filepath.Clean(path)
}
