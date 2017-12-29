package main

import (
	"archive/zip"
	"io"
	"os"
	"strings"
)

func main() {
	// zipの内容を書き込むファイル
	file, err := os.Create("sample.zip")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// zipファイル
	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// ファイルの数だけ書き込み
	a, err := zipWriter.Create("a.txt")
	if err != nil {
		panic(err)
	}
	io.Copy(a, strings.NewReader("ひとつめ"))
}

// func main() {
// 	// 圧縮ファイル名
// 	dest, err := os.Create("test.txt.zip")
// 	if err != nil {
// 		panic(err)
// 	}
// 	zw := zip.NewWriter(dest)
// 	defer zw.Close()
//
// 	// 圧縮したいファイル名
// 	src, err := os.Open("test.txt")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer src.Close()
//
// 	// 解答した時のファイル名
// 	z, err := zw.Create("test.txt")
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// 内容のコピー
// 	_, err = io.Copy(z, src)
// 	if err != nil {
// 		panic(err)
// 	}
// }
