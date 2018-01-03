package main

import "os"

func main() {
	os.Mkdir("setting", 0644)
	os.MkdirAll("setting/myapp/network", 0644)
}
