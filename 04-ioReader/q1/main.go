package main

import (
	"io"
	"os"
	"strings"
)

var (
	computer   = strings.NewReader("COMPUTER")
	system     = strings.NewReader("SYSTEM")
	programing = strings.NewReader("PROGRAMMING")
)

// PROGRAMING の A
// SYSTEMの S
// COMPUTER の C
// PROGRAMING の I
// PROGRAMING の I

func main() {
	var stream io.Reader
	a := io.NewSectionReader(programing, 5, 1)
	s := io.LimitReader(system, 1)
	c := io.LimitReader(computer, 1)
	i := io.NewSectionReader(programing, 8, 1)
	pr, pw := io.Pipe()
	// 2つのwriterをひとつのwriterにする
	writer := io.MultiWriter(pw, pw)
	// ここで、2つのwriterにiが書き込まれる
	go io.CopyN(writer, i, 1)
	defer pw.Close()
	stream = io.MultiReader(a, s, c, io.LimitReader(pr, 2))
	io.Copy(os.Stdout, stream)
}

// func main() {
// 	var stream io.Reader
// 	aSectionReader := io.NewSectionReader(programing, 5, 1)
// 	sSectionReader := io.NewSectionReader(system, 0, 1)
// 	cSectionReader := io.NewSectionReader(computer, 0, 1)
// 	iSectionReader := io.NewSectionReader(programing, 8, 1)
// 	i2SectionReader := io.NewSectionReader(programing, 8, 1)
//
// 	mReader := io.MultiReader(
// 		aSectionReader,
// 		sSectionReader,
// 		cSectionReader,
// 		iSectionReader,
// 		i2SectionReader,
// 	)
// 	stream = mReader
// 	io.Copy(os.Stdout, stream)
// }
