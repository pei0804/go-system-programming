package main

import (
	"compress/gzip"
	"os"
)

func main() {
	file, err := os.Create("test.txt.gz")
	if err != nil {
		panic(err)
	}
	writer := gzip.NewWriter(file)
	writer.Header.Name = "test.txt"
	writer.Write([]byte("gzip.Writer example\n"))
	writer.Close()
}
