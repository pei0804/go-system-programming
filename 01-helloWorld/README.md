# 01

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello World!")
}
```

```console
Printlnから実行されているのは、Fprintln
ここで渡しているos.Stdoutは、標準出力のこと。
また、aには"Hello World!"の文字列が入っているが、数値や他の様々な型に対応するためinterface{}となっている。
Stdout = NewFile(uintptr(syscall.Stdout), "/dev/stdout")

> fmt.Println() /Users/jumpei/.anyenv/envs/goenv/versions/1.8/src/fmt/print.go:256 (PC
: 0x10800bf)
   251: }
   252:
   253: // Println formats using the default formats for its operands and writes to st
andard output.
   254: // Spaces are always added between operands and a newline is appended.
   255: // It returns the number of bytes written and any write error encountered.
=> 256: func Println(a ...interface{}) (n int, err error) {
   257:         return Fprintln(os.Stdout, a...)
   258: }
   259:

> fmt.Fprintln() /Users/jumpei/.anyenv/envs/goenv/versions/1.8/src/fmt/print.go:245 (P
C: 0x107ffc3)
   240: // after the last operand.
   241:
   242: // Fprintln formats using the default formats for its operands and writes to w
.
   243: // Spaces are always added between operands and a newline is appended.
   244: // It returns the number of bytes written and any write error encountered.
=> 245: func Fprintln(w io.Writer, a ...interface{}) (n int, err error) {
   246:         p := newPrinter()
                // 出力フォーマットを整える
   247:         p.doPrintln(a)
                // 標準出力に書き込む
   248:         n, err = w.Write(p.buf)
   249:         p.free()
   250:         return
(dlv)

Writeで呼び出されているのFile#Write

// Write writes len(b) bytes to the File.
// It returns the number of bytes written and an error, if any.
// Write returns a non-nil error when n != len(b).
func (f *File) Write(b []byte) (n int, err error) {
	if err := f.checkValid("write"); err != nil {
		return 0, err
	}
  // ここを見る
	n, e := f.write(b)
	if n < 0 {
		n = 0
	}
	if n != len(b) {
		err = io.ErrShortWrite
	}

	epipecheck(f, e)

	if e != nil {
		err = &PathError{"write", f.name, e}
	}
	return n, err
}

環境によってここで呼び出されるメソッドが変わる

// write writes len(b) bytes to the File.
// It returns the number of bytes written and an error, if any.
func (f *File) write(b []byte) (n int, err error) {
	for {
		bcap := b
		if needsMaxRW && len(bcap) > maxRW {
			bcap = bcap[:maxRW]
		}
    // syscallが呼ばれる（違うOSでもここは同じ）
		m, err := fixCount(syscall.Write(f.fd, bcap))
		n += m

		// If the syscall wrote some data but not all (short write)
		// or it returned EINTR, then assume it stopped early for
		// reasons that are uninteresting to the caller, and try again.
		if 0 < m && m < len(bcap) || err == syscall.EINTR {
			b = b[m:]
			continue
		}

		if needsMaxRW && len(bcap) != len(b) && err == nil {
			b = b[m:]
			continue
		}

		return n, err
	}
}

![フロー](http://ascii.jp/elem/000/001/234/1234873/19x_1200x525.jpg)
