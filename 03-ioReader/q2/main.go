package main

import (
	"crypto/rand"
	"io"
	"os"
)

func main() {
	file, err := os.Create("rand.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	io.CopyN(file, rand.Reader, 1024)
}

// func main() {
// 	buf := make([]byte, 1024)
// 	_, err := io.ReadFull(rand.Reader, buf)
// 	if err != nil {
// 		panic(err)
// 	}
// 	file, err := os.Create("rand.txt")
// 	if err != nil {
// 		panic(err)
// 	}
// 	_, err = file.Write(buf)
// 	if err != nil {
// 		panic(err)
// 	}
// }
