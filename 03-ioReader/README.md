# 03 io.Reader

Goはファイルの読み込みやネットワークアクセスが簡単に出来る関数が用意されている。  
これらのAPIは、io.Readerやio.Writerを隠蔽し、特定の用途で簡単に使えるようにしたものです。  

- ioutil.WriteFile() ファイルに書き込める
- ioutil.ReadFile() ファイルを読み込める
- http.Get() HTTPのGETメソッドが使える
- http.Post() HTTPのPOSTメソッドが使える

ソフトウェアは基本的に何らかの入力に対して、書こうを行なってから出力をするというパイプ構造になっています。  
これらを思い通りに使えることは、開発者として強みになると言えます。  

## io.Readerとio.Writer、その仲間たち

io.Writerは様々なデータを出力することが出来るものでした。io.Readerは逆に様々なものを読み込むものです。

```go
typeeader interface {
	Read(p []byte) (n int, err error)
}
```

Readインターフェースは、
引数：読み込んだ内容を一時的に入れておくバッファです。嗚呼かじめメモリを用意しておいてそれを使います。（makeなどを使うと良い）  
返り値：読み込んだバイト数とエラーを返す

```go
// 1024バイトのバッファをmakeで作る
buffer := make([]byte, 1024)
// sizeは実際に読み込んだバイト数、errはエラー
size, err := r.Read(buffer)
```

書き込みに比べると、バッファを用意したり、その長さを管理したりと面倒です。しかし、低レベルなインターフェースだけではなく、簡単に扱うための機能も用意されています。

## io.Readerの補助関数

そのまま使うと多少不便なので、ヘルパーを使うと良い

### 読み込みの補助関数

`ioutil.ReadAll()`  
終端記号に当たるまでデータを読み込んで返す。メモリに収まらない可能生のあるものでは使えません。

```go
// すべて読み込む
buffer, err := ioutil.ReadAll(reader)
```

`ioutilReadFull()`  
決まったバイト数で確実に読み込む。

```go
// 4バイト読み込めないとエラー
buffer := make([]byte, 4)
size, err := io.ReadFull(reader, buffer)
```

### コピー補助関数

`io.Copy`  
io.Reader -> io.Writerにそのまま渡したい時などに使う。

```go
// すべてコピー
writeSize, err := io.Copy(writer, reader)
// 指定したサイズだけコピー
writeSize, err := io.CopyN(writer, reader, size)
```

`io.CopyBuffer`  
あらかじめコピーする量が決まっていて、無駄なバッファを使いたくない時は、コピーするバッファを使いまわしたい時に有効。自身が作った作業バッファを渡すことが出来る。

```go
// 8KBのバッファを使う
buffer := make([]byte, 8 * 1024)
io.CopyBuffer(writer, reader, buffer)
```

## io.Readerを満たす構造体でよく使うもの

### 標準入力

|変数    |io.Reader|io.Writer|io.Seeker|io.Closer|
|:-------|:--------|:--------|:--------|:--------|
|os.Stdin|✔        |         |         |✔        |

標準入力に対応するオブジェクトがos.Stdinです.
以下のプログラムは、実行すると入力街となり、エンターが押される度に結果が返ってきます。

```go
package main
　
import (
　　　"fmt"
　　　"io"
　　　"os"
)
　
func main() {
　　　for {
　　　　　　buffer := make([]byte, 5)
            // 標準入力から入力内容を受け取る
　　　　　　size, err := os.Stdin.Read(buffer)
　　　　　　if err == io.EOF {
　　　　　　　　　fmt.Println("EOF")
　　　　　　　　　break
　　　    }
　　　fmt.Printf("size=%d input='%s'\n", size, string(buffer))
　　　}
}
```


### ファイル入力

|変数    |io.Reader|io.Writer|io.Seeker|io.Closer|
|:-------|:--------|:--------|:--------|:--------|
|os.Stdin|✔        |✔        |✔        |✔        |

新規作成は、os.Create()、os.Openを使うとファイルを開くことが出来ますが、内部的にはos.OpenFile関数がフラグの違うだけで、同じ関数を呼び出しています。

```go
func Open(name string) (*File, error) {
　　　return OpenFile(name, O_RDONLY, 0)
}
　
func Create(name string) (*File, error) {
　　　return OpenFile(name, O_RDWR|O_CREATE|O_TRUNC, 0666)
}
```

ファイルの読み込み方法の例  
ファイルを一度開いたらClose()する必要があります。Goでは確実に行う後処理を実行するのに便利なのがdeferです。これは現在のスコープが終了したら実行するというものです。

```go
package main
　
import (
　　　"io"
　　　"os"
)
　
func main() {
　　　file, err := os.Open("file.go")
　　　if err != nil {
　　　　　　panic(err)
　　　}
　　　defer file.Close()
      // 読み込んだ内容を標準出力にコピーする
　　　io.Copy(os.Stdout, file)
}
```

### インターネット通信

|変数    |io.Reader|io.Writer|io.Seeker|io.Closer|
|:-------|:--------|:--------|:--------|:--------|
|os.Stdin|✔        |✔        |         |✔        |

インターネット上でのデータのやり取りは、送信データを送信者から見ると書き込みで、受信側から見ると読み込みです。  
以下のようにすることで、シンプルなHTTPのやり取りを実装することが出来ます。


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
	    // 送信データ
　　　conn.Write([]byte("GET / HTTP/1.0\r\nHost: ascii.jp\r\n\r\n"))
　　　io.Copy(os.Stdout, conn)
}
```

しかし、毎回上記のようにRFCの規約に乗っ取りやり取りを書くのは現実的ではありません。  
そこで使えるのが、http.ReadResponseです。これでヘッダーやボディーなどが簡単に取り出すことが出来ます。

```go
package main
　
import (
　　　"bufio"
　　　"fmt"
　　　"io"
　　　"net"
　　　"net/http"
　　　"os"
)
　
func main() {
　　　conn, err := net.Dial("tcp", "ascii.jp:80")
　　　if err != nil {
　　　　　　panic(err)
　　　}
　　　conn.Write([]byte("GET / HTTP/1.0\r\nHost: ascii.jp\r\n\r\n"))
　　　res, err := http.ReadResponse(bufio.NewReader(conn), nil)
　　　// ヘッダーを表示してみる
　　　fmt.Println(res.Header)
　　　// ボディーを表示してみる。最後にはClose()すること
　　　defer res.Body.Close()
　　　io.Copy(os.Stdout, res.Body)
}
```

実際には、NewRequestというメソッドなどを使って簡単な方法でやり取りが出来ます。実際に自分でWriteメソッドなどを使ってやり取りすることは稀です。

```go
req, err := http.NewRequest("GET", "http://ascii.jp", nil)
req.Write(conn)
```

## メモリに蓄えた内容をio.Readerとして読み出すバッファ

最初の初期化は実体を渡す必要がある

```go
// 空のバッファ
var buffer1 bytes.Buffer
// バイト列で初期化
buffer2 := bytes.NewBuffer([]{byte{0x10, 0x20, 0x30})
// 文字列で初期化
buffer3 := bytes.NewBufferString("初期文字列")
```

```go
// bytes.Readerはbytes.NewReaderで作成
bReader1 := bytes.NewReader([]byte{0x10, 0x20, 0x30})
bReader2 := bytes.NewReader([]byte("文字列をバイト配列にキャストして設定")
　
// strings.Readerはstrings.NewReader()関数で作成
sReader := strings.NewReader("Readerの出力内容は文字列で渡す")
```

## io.Readerとio.Writerを考えるためのモデル図

![モデル図](http://ascii.jp/elem/000/001/252/1252955/img.html)
