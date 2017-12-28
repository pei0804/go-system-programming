# 02 io.Writer

システムコール = ファイルディスクリプタ  
ファイルディスクリプタとは、一種の識別子（数値）で、その数値を指定してコールすると、数値に対応したものにアクセスが出来る。

```go
func (f *File) write(b []byte) (n int, err error) {
    for {
        bcap := b
        if needsMaxRW && len(bcap) > maxRW {
            bcap = bcap[:maxRW]
        }
        // f.fdが数値に当たるもの
        m, err := fixCount(syscall.Write(f.fd, bcap))
        n += m
        :
    }
}
```

ファイルディスクリプタの種類（本当はもっとある）

- CPU
- ソケット
- 標準入出力

上記のものをファイルとして扱えるように抽象化してくれている。  
これらの仕組みはOSがカーネルレイヤで用意している抽象化の仕組みです。  
OSはプロセスが起動されれうとまず、3の疑似ファイルを作成します。

- 0 標準入力
- 1 標準出力
- 2 標準エラー出力

プロセスが増える度に1ずつ大きな数値が割り当てられる。
しかし、これらの仕組みはOSによって異なるので、Goでは違いを吸収してくれています。そのひとつが`io.Writer`です。

ファイルディスクリプタを直接指定して、ファイルを作成することも出来る

`file, err := os.NewFile(ファイルディスクリプタ, 名前)`

## io.Writerはインターフェース

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

Writerというインターフェースは  
引数：バイト列のbを読み込む  
返り値：書き込んだバイト数nと発生したエラーを返す  

POSIX系OSでは可能な限りファイルとして抽象化している。  
それと同じように様々な動きが、同じメソッド名で適用出来るように定義されているインターフェースです。


Hello World!の出力時に呼び出されていたWriteも同じ型です。  
この状態をインターフェースを満たすといいます。

```go
func (f *File) Write(b []byte) (n int, err error) {
      if f == nil {
            return 0, ErrInvalid
        }
        n, e := f.write(b)
        :
}
```

## io.Writerを使う構造体の例

write()というメソッドで、同じように書き出している例

### File

```go
package main
 
import (
    "os"
)
 
func main() {
    file, err := os.Create("test.txt")
    if err != nil {
        panic(err)
    }
    file.Write([]byte("os.File example\n"))
    file.Close()
}
```


```go
package main
 
import (
    "os"
)
 
func main() {
    os.Stdout.Write([]byte("os.Stdout example\n"))
}
```

### Buffer

```go
package main
 
import (
    "bytes"
    "fmt"
)
 
func main() {
    var buffer bytes.Buffer
    buffer.Write([]byte("bytes.Buffer example\n"))
    fmt.Println(buffer.String())
}
```

### TCP

```go
package main
 
import (
    "io"
    "os"
    "net"
)
 
func main() {
    conn, err := net.Dial("tcp", "ascii.jp:80")
    if err != nil {
        panic(err)
    }
    conn.Write([]byte("GET / HTTP/1.0\r\nHost: ascii.jp\r\n\r\n"))
    io.Copy(os.Stdout, conn)
}
```

### HTTP

```go
package main
 
import (
    "net/http"
)
 
func handler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("http.ResponseWriter sample"))
}
 
func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
```

## io.Writerのフィルタ

### io.MultiWriter

書き込む場所を複数指定する場合はMultiWriterを使う

```go
package main
 
import (
    "io"
    "os"
)
 
func main() {
    file, err := os.Create("multiwriter.txt")
    if err != nil {
        panic(err)
    }
    writer := io.MultiWriter(file, os.Stdout)
    io.WriteString(writer, "io.MultiWriter example\n")
}
```

### あらかじめ作成していたfileに操作を行う

```go
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
```

### bufio.Writer

出力結果を一時的にためておいて、ある程度の分量ごとにまとめて書き出す。  
他言語でいうところのバッファ付き出力にあたるもの。  
Flush()を呼ぶと、io.Writerに書き出す。Flushを呼ばなければデータを抱えたまま死ぬ。

```go
package main
 
import (
    "bufio"
    "os"
)
 
func main() {
    buffer := bufio.NewWriter(os.Stdout)
    buffer.WriteString("bufio.Writer ")
    buffer.Flush()
    buffer.WriteString("example\n")
    buffer.Flush()
}
```

### フォーマットしてio.Writerにかき出す

Formatを指定することが出来る。  
`%v`はプリミティブ型でもそうじゃない型でもString()メソッドがあればそれを表示に使って出力してくれます。
※プリミティブとは単純なもの。構造などを持たないものなど。

```go
package main
 
import (
    "os"
    "fmt"
    "time"
)
 
func main() {
    fmt.Fprintf(os.Stdout, "Write with os.Stdout at %v", time.Now())
}
```

#### JSON

JSONを生計して、io.Writerにかき出すことも出来ます。

```go
// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w, escapeHTML: true}
}
```

```go
package main
 
import (
    "os"
    "encoding/json"
)
 
func main() {
    encoder := json.NewEncoder(os.Stdout)
    encoder.SetIndent("", "    ")
    encoder.Encode(map[string]string{
        "example": "encoding/json",
        "hello": "world",
    })
}
```

### HTTP

あまり使わないですが、Request構造体もWriterが使えます。
Transfer-Encoding: chunked でチャンクに分けて送信したり、プロトコルのアップグレードで別のプロトコルと併用するようなHTTPリクエストを送るときには使うことになるでしょう（もっとも、そういったことをするケースもまれといえばまれですが）。

```go
// If Body is present, Content-Length is <= 0 and TransferEncoding
// hasn't been set to "identity", Write adds "Transfer-Encoding:
// chunked" to the header. Body is closed after it is sent.
func (r *Request) Write(w io.Writer) error {
	return r.write(w, false, nil, nil)
}
```

```go
package main
 
import (
    "os"
    "net/http"
)
 
func main() {
    request, err := http.NewRequest("GET", "http://ascii.jp", nil)
    if err != nil {
        panic(err)
    }
    request.Header.Set("X-TEST", "ヘッダーも追加できます")
    request.Write(os.Stdout)
}
```

## io.Writerの実装状況・利用状況を調べる

GoではJavaのインターフェースなどとは異なり、このインターフェースを持っているという宣言を構造体側には書きません。構造体がインターフェースを満たすメソッドを持っているかどうかは、インターフェースの変数に構造体のポインタを代入したり、メソッドの引数にポインタを渡したりした時に、自動的に認識されます。  
なので、単純にコード検索しただけでは分かりません。こういう場合は、godocというコマンドを使うと楽に見ることが出来ます。  

--analysis typeとつけるとインターフェースの分析を行なってくれます。  
 http://localhost:6060/pkg/io/#Writer を見ると。 -analysis type を付けて実行すると、golang.orgにはない、 implements という項目が追加されます。ここに io.Writer を実装した構造体などの一覧が表示されます。  

```console
godoc -http ":6060" -analysis type
```

