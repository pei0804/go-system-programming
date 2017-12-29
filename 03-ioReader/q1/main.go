package main

import (
	"io"
	"os"
)

func main() {
	oldFile, err := os.Open("old.txt")
	if err != nil {
		panic(err)
	}
	defer oldFile.Close()
	newFile, err := os.Create("new.txt")
	if err != nil {
		panic(err)
	}
	defer newFile.Close()
	io.Copy(newFile, oldFile)
}

// func main() {
// 	oldFile, err := os.Open("old.txt")
// 	if err != nil {
// 		panic(err)
// 	}
// 	v, err := ioutil.ReadAll(oldFile)
// 	if err != nil {
// 		panic(err)
// 	}
// 	newFile, err := os.Create("new.txt")
// 	if err != nil {
// 		panic(err)
// 	}
// 	_, err = newFile.Write(v)
// 	if err != nil {
// 		panic(err)
// 	}
// }
