# 07 TCP

## 圧縮

HTTPの速度アップ手法としてよく使われるのが圧縮です。昔よりインターネットやWiFiの性能は向上しましたが、それでもCPUを使って圧縮することにより通信量を減らす方が、メリットが大きいことが多々あります。今回は一般的なブラウザでも使われているgzip圧縮を実装してみましょう。

圧縮してもパケット伝達の速度は変わりませんが、転送開始から終了までの時間は短くなります。

### gzip圧縮に対応したクライアント

```go
package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

func main() {
	sendMessages := []string{
		"ASCII",
		"PROGRAMMING",
		"PLUS",
	}
	current := 0
	var conn net.Conn
	// リトライ用にループで全体を囲う
	for {
		var err error
		// まだコネクションを張ってない / エラーでリトライ時はDialから行う
		if conn == nil {
			conn, err = net.Dial("tcp", "localhost:8888")
			if err != nil {
				panic(err)
			}
			fmt.Printf("Access: %d\n", current)
		}
		// POSTで文字列を送るリクエストを作成
		request, err := http.NewRequest("POST", "http://localhost:8888", nil)
		strings.NewReader(sendMessages[current])
		if err != nil {
			panic(err)
		}
		request.Header.Set("Accept-Encoding", "gzip")
		err = request.Write(conn)
		if err != nil {
			panic(err)
		}
		// サーバから読み込む。タイムアウトはここでエラーになるのでリトライ
		response, err := http.ReadResponse(bufio.NewReader(conn), request)
		if err != nil {
			fmt.Println("Retry")
			conn = nil
			continue
		}
		// 結果を表示
		dump, err := httputil.DumpResponse(response, false)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(dump))
		defer response.Body.Close()
		if response.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(response.Body)
			if err != nil {
				panic(err)
			}
			io.Copy(os.Stdout, reader)
			reader.Close()
		} else {
			io.Copy(os.Stdout, response.Body)
		}
		// 全送信完了していれば終了
		current++
		if current == len(sendMessages) {
			break
		}
	}
}
```

#### コード解説

```go
// POSTで文字列を送るリクエストを作成
request, err := http.NewRequest("POST", "http://localhost:8888", nil)
strings.NewReader(sendMessages[current])
if err != nil {
  panic(err)
}
request.Header.Set("Accept-Encoding", "gzip")
```

リクエスト生成部分を改造して、自分が対応しているアルゴリズムを宣言するようにしています。サーバから自分が理解出来ない圧縮フォーマットでデータを送りつけられてもクライアントではそれを読み込めないからです。


```go
dump, err := httputil.DumpResponse(response, false)
if err != nil {
  panic(err)
}
fmt.Println(string(dump))
defer response.Body.Close()
if response.Header.Get("Content-Encoding") == "gzip" {
  reader, err := gzip.NewReader(response.Body)
  if err != nil {
    panic(err)
  }
  io.Copy(os.Stdout, reader)
  reader.Close()
} else {
  io.Copy(os.Stdout, response.Body)
}
```

httputil.DumpResponse()は圧縮された内容を理解してくれません。2つ目のパラメータをfalseにすることでbodyを無視するように指示しています。

Accept-Encodingで表明した圧縮メソッドにサーバーが対応しているかどうかは、Content-Encodingを見ればわかります。今回は一種類ですが、複数の候補を提示してサーバに選ばせることも出来ます。

### gzip圧縮に対応したサーバ

```go
package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

func isGzipAceeptable(request *http.Request) bool {
	return strings.Index(strings.Join(request.Header["Accept-Encoding"], ","), "gzip") != -1
}

func processSession(conn net.Conn) {
	fmt.Printf("Accept %v\n", conn.RemoteAddr())
	defer conn.Close()
	for {
		// リクエストを読み込む
		// タイムアウトを設定
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		// リクエストを読み込む
		request, err := http.ReadRequest(bufio.NewReader(conn))
		if err != nil {
			// タイムアウトもしくはソケットクローズ時は終了
			// それ以外はエラーにする
			neterr, ok := err.(net.Error) // ダウンキャスト
			if ok && neterr.Timeout() {
				fmt.Println("Timeout")
				break
			} else if err == io.EOF {
				break
			}
			panic(err)
		}
		dump, err := httputil.DumpRequest(request, true)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(dump))
		response := http.Response{
			StatusCode: 200,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header),
		}
		if isGzipAceeptable(request) {
			content := "Hello World\n"
			// コンテンツをgzip化して転送
			var buffer bytes.Buffer
			writer := gzip.NewWriter(&buffer)
			io.WriteString(writer, content)
			writer.Close()
			response.Body = ioutil.NopCloser(&buffer)
			response.ContentLength = int64(buffer.Len())
			response.Header.Set("Content-Encoding", "gzip")
		} else {
			content := "Hello World\n"
			response.Body = ioutil.NopCloser(strings.NewReader(content))
			response.ContentLength = int64(len(content))
		}
		response.Write(conn)
	}
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}
	fmt.Println("Server is running at localhost:8888")
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go processSession(conn)
	}
}
```

#### コード解説

```go
if isGzipAceeptable(request) {
  content := "Hello World\n"
  // コンテンツをgzip化して転送
  var buffer bytes.Buffer
  writer := gzip.NewWriter(&buffer)
  io.WriteString(writer, content)
  writer.Close()
  response.Body = ioutil.NopCloser(&buffer)
  response.ContentLength = int64(buffer.Len())
  response.Header.Set("Content-Encoding", "gzip")
} else {
  content := "Hello World\n"
  response.Body = ioutil.NopCloser(strings.NewReader(content))
  response.ContentLength = int64(len(content))
}
response.Write(conn)
```

クライアントがgzip受け入れ可能かどうかに応じて中に入れるコンテンツを変えています。圧縮には、gzip.NewWriterで作成したio.Writerを使います。圧縮した内容はbytes.Bufferに書き出しています。さらにContent-Lengthヘッダーに圧縮後にボディサイズを指定します。

このコードを見て分かる通り、ヘッダーは圧縮されません。そのため少量のデータを通信するほど効率がわるくなります。20ばいと足らずのサンプルの文字列ではgzipのオーバーヘッドの方が大きく、サイズが倍増してしまいます。大きいサイズでは降下が出てきます。ヘッダーの圧縮はHTTP/2になって導入されました。

## チャンク形式のボディ送信

これまで紹介してきた通信処理は、一度のリクエストに対して、必要な情報を一度に全て送るというものでした。そのため、全部のデータが出来上がるまでレスポンスのスタートが遅れます。結果として最終的な終了時間も伸び、実行効率は下がります。

大きいファイルなどをクライントに一度に返すと、メモリがそれに全て取られてしまうことなどが起きるため、チャンク形式をサポートして、これらの対処する必要があります。

### サーバー

```go
package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
)

// 青空文庫: ごんぎつねより
// http://www.aozora.gr.jp/cards/000121/card628.html
var contents = []string{
	"これは、私わたしが小さいときに、村の茂平もへいというおじいさんからきいたお話です。",
	"むかしは、私たちの村のちかくの、中山なかやまというところに小さなお城があって、",
	"中山さまというおとのさまが、おられたそうです。",
	"その中山から、少しはなれた山の中に、「ごん狐ぎつね」という狐がいました。",
	"ごんは、一人ひとりぼっちの小狐で、しだの一ぱいしげった森の中に穴をほって住んでいました。",
	"そして、夜でも昼でも、あたりの村へ出てきて、いたずらばかりしました。",
}

func processSession(conn net.Conn) {
	fmt.Printf("Accept %v\n", conn.RemoteAddr())
	defer conn.Close()
	for {
		// リクエストを読み込む
		request, err := http.ReadRequest(bufio.NewReader(conn))
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		dump, err := httputil.DumpRequest(request, true)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(dump))
		// レスポンスを書き込む
		fmt.Fprintf(conn, strings.Join([]string{
			"HTTP/1.1 200 OK",
			"Content-Type: text/plain",
			"Transfer-Encoding: chunked",
			"", "",
		}, "\r\n"))
		for _, content := range contents {
			bytes := []byte(content)
			fmt.Fprintf(conn, "%x\r\n%s\r\n", len(bytes), content)
		}
		fmt.Fprintf(conn, "0\r\n\r\n")
	}
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}
	fmt.Println("Server is running at localhost:8888")
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go processSession(conn)
	}
}
```

```console
75
これは、私わたしが小さいときに、村の茂平もへいというおじいさんからきいたお話です。
45
中山さまというおとのさまが、おられたそうです。
:
0
```

#### コード解説

```go
for _, content := range contents {
  bytes := []byte(content)
  fmt.Fprintf(conn, "%x\r\n%s\r\n", len(bytes), content)
}
```

http.Responseはファイルサイズ指定がないとConnection: closeを送ってしまうため、ここではfmt.FprintfでHTTPレスポンスを直接書き込んでいます。


### クライント

```go
package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}
	request, err := http.NewRequest(
		"GET",
		"http://localhost:8888",
		nil,
	)
	if err != nil {
		panic(err)
	}
	err = request.Write(conn)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(conn)
	response, err := http.ReadResponse(reader, request)
	if err != nil {
		panic(err)
	}
	dump, err := httputil.DumpResponse(response, false)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(dump))
	if len(response.TransferEncoding) < 1 || response.TransferEncoding[0] != "chunked" {
		panic("wrong transfer encoding")
	}
	for {
		// サイズを取得
		sizeStr, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		// 16進数のサイズをパース。サイズがゼロならクローズ
		size, err := strconv.ParseInt(string(sizeStr[:len(sizeStr)-2]), 16, 64)
		if size == 0 {
			break
		}
		if err != nil {
			panic(nil)
		}
		// サイズ数分バッファを確保して読み込み
		line := make([]byte, int(size))
		reader.Read(line)
		reader.Discard(2)
		fmt.Printf(" %d bytes: %s\n", size, string(line))
	}
}
```

```console
HTTP/1.1 200 OK
Transfer-Encoding: chunked
Content-Type: text/plain


 123 bytes: これは、私わたしが小さいときに、村の茂平もへ
いというおじいさんからきいたお話です。
 117 bytes: むかしは、私たちの村のちかくの、中山なかやま
というところに小さなお城があって、
 69 bytes: 中山さまというおとのさまが、おられたそうです
。
 108 bytes: その中山から、少しはなれた山の中に、「ごん狐
ぎつね」という狐がいました。
 132 bytes: ごんは、一人ひとりぼっちの小狐で、しだの一ぱ
いしげった森の中に穴をほって住んでいました。
 102 bytes: そして、夜でも昼でも、あたりの村へ出てきて、
いたずらばかりしました。
```

```console
Server is running at localhost:8888
Accept 127.0.0.1:57334
GET / HTTP/1.1
Host: localhost:8888
User-Agent: Go-http-client/1.1
```

#### コード解説

```go
for {
  // サイズを取得
  sizeStr, err := reader.ReadBytes('\n')
  if err == io.EOF {
    break
  }
  // 16進数のサイズをパース。サイズがゼロならクローズ
  size, err := strconv.ParseInt(string(sizeStr[:len(sizeStr)-2]), 16, 64)
  if size == 0 {
    break
  }
  if err != nil {
    panic(nil)
  }
  // サイズ数分バッファを確保して読み込み
  line := make([]byte, int(size))
  reader.Read(line)
  reader.Discard(2)
  fmt.Printf(" %d bytes: %s\n", size, string(line))
}
```

for文内でチャンクを読み込む。  
改行を探し、サイズを取得したら、そのサイズ文だけ読み込む。

## パイプラインニング

送受信を非同期化することでトータルの通信にかかる時間を大幅に減らす方法です。  
この機能はパイプランニングと呼ばれ、HTTP/1.1の規格に含まれています。パイプランニングでは、レスポンスがくる前に次から次にリクエストを多重で飛ばすことで、最終的に通信が完了するまでの時間を短くする。（通常だと待ち時間がある）

実はこのパイプランニングはHTTPの歴史における最も不幸な機能でした。規格には入りましたが、後方互換がない機能であるため、HTTP/1.0しか解釈できないプロキシがとゆうにあると通信が完了しなくなる問題があります。またサーバの実装が不十分な場合もありました。

ブラウザの対応も十分とは言えませんでした。Netscape Navigatiorでは、サーバが自称するX-Powerd-Byヘッダーを見てパイプランニングを使うかどうか決める条件分岐ロジックが組まれていました。  
Chromeブラウザでは、一旦実装されたものの、後で削除されました。Safariでもがぞうが入れ替わるバグなどを誘発しました。サポートされているブラウザでも、何度か通信してサーバーの対応を確認してから有効となるようになっています。

### パイプラインのサーバー実装

```go
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

func writeToConn(sessionResponses chan chan *http.Response, conn net.Conn) {
	defer conn.Close()
	// 順番に取り出す
	for sessionResponse := range sessionResponses {
		// 選択された仕事が終わるまで待つ
		response := <-sessionResponse
		response.Write(conn)
		close(sessionResponse)
	}
}

func handleRequest(request *http.Request, resultReceiver chan *http.Response) {
	dump, err := httputil.DumpRequest(request, true)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(dump))
	content := "Hello World\n"
	// レスポンスをかきこむ
	// セッションを維持するためにKeep-Aliveでないといけない
	response := &http.Response{
		StatusCode:    200,
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: int64(len(content)),
		Body:          ioutil.NopCloser(strings.NewReader(content)),
	}
	// 処理が終わったらチャネルに書き込み
	// ブロックされていたwriteToConnの処理を再始動する
	resultReceiver <- response
}

func processSession(conn net.Conn) {
	fmt.Printf("Accept %v\n", conn.RemoteAddr())
	// セッション内のリクエストを順に処理するためのチャネル
	sessionResponses := make(chan chan *http.Response, 50)
	defer close(sessionResponses)
	// レスポンスを直列化してソケットにかき出す専用のゴルーチン
	go writeToConn(sessionResponses, conn)
	reader := bufio.NewReader(conn)
	defer conn.Close()
	for {
		// レスポンスを受け取ってセッションのキューに入れる
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		// リクエストを読み込む
		request, err := http.ReadRequest(reader)
		if err != nil {
			neterr, ok := err.(net.Error)
			if ok && neterr.Timeout() {
				fmt.Println("Timeout")
				break
			} else if err == io.EOF {
				break
			}
			panic(err)
		}
		sessionResponse := make(chan *http.Response)
		sessionResponses <- sessionResponse
		// 非同期でレスポンスを実行
		go handleRequest(request, sessionResponse)
	}
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}
	fmt.Println("Server is running at localhost:8888")
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go processSession(conn)
	}
}
```

#### コード解説

```go
func handleRequest(request *http.Request, resultReceiver chan *http.Response) {
	dump, err := httputil.DumpRequest(request, true)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(dump))
	content := "Hello World\n"
	// レスポンスをかきこむ
	// セッションを維持するためにKeep-Aliveでないといけない
	response := &http.Response{
		StatusCode:    200,
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: int64(len(content)),
		Body:          ioutil.NopCloser(strings.NewReader(content)),
	}
	// 処理が終わったらチャネルに書き込み
	// ブロックされていたwriteToConnの処理を再始動する
	resultReceiver <- response
}
```

リクエストごとに非同期処理でレスポンスを返す処理(handleRequest())を呼び出しています。今回はただ文字列を返しているだけなので処理時間が大きく変わることはありません。

レスポンスの順番の制御はGo言語のデータ構造のチャネルを使っています。チャネルはFIFOのキューで、バッファなしとありの2種類があります。利用するには下記のように型を指定して初期化します。

```go
// バッファなし
unbuffered := make(chan string)
// バッファあり
buffered := make(chan string, 10)
```

データの入出力には<-演算子を使います。
バッファありの場合は、指定した個数までは自由に投入出来ますが、指定した個数のデータが入っている時は、さらにデータを追加しようとしてもブロックされます。もし、他のゴルーチンで受信が行われ空きが発生したら再開される。  
バッファなしの場合は、投入しようとしたスレッドは即座にブロックされます。


```go
func writeToConn(sessionResponses chan chan *http.Response, conn net.Conn) {
	defer conn.Close()
	// 順番に取り出す
	for sessionResponse := range sessionResponses {
		// 選択された仕事が終わるまで待つ
		response := <-sessionResponse
		response.Write(conn)
		close(sessionResponse)
	}
}
```

パイプランニング対応サーバ実装では、まず並列処理でレスポンスを書き込むwriteToConn()関数が順序を守って書けるように、先頭から1つずつデータを処理出来るように、バッファなしの
チャネルで、キューの仕組みを作っています。また、リクエスト処理が終わるまで待つために送信データをためるためにバッファなしのチャネルを内部のもう一つ用意しています。待つ側のコードはwriteToConn()の中に、送信側のコードはhandleRequest()の最後にあります。 チャネルまわりの構成を図にするとこんなイメージです。

### パイプランニングのクライアント

```go
package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
)

func main() {
	sendMessages := []string{
		"ASCII",
		"PROGRAMMING",
		"PLUS",
	}
	current := 0
	var conn net.Conn
	var err error
	requests := make(chan *http.Request, len(sendMessages))
	conn, err = net.Dial("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Access: %d\n", current)
	defer conn.Close()
	// リクエストだけ先に送る
	for i := 0; i < len(sendMessages); i++ {
		lastMessage := i == len(sendMessages)-1
		request, err := http.NewRequest(
			"GET",
			"http://localhost:8888?message="+sendMessages[i],
			nil,
		)
		if lastMessage {
			request.Header.Add("Connection", "close")
		} else {
			request.Header.Add("Connection", "keep-alive")
		}
		if err != nil {
			panic(err)
		}
		err = request.Write(conn)
		if err != nil {
			panic(err)
		}
		fmt.Println("send: ", sendMessages[i])
		requests <- request
	}
	close(requests)
	// レスポンスをまとめて受信
	reader := bufio.NewReader(conn)
	for request := range requests {
		response, err := http.ReadResponse(reader, request)
		if err != nil {
			panic(err)
		}
		dump, err := httputil.DumpResponse(response, true)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(dump))
		if current == len(sendMessages) {
			break
		}
	}
}```

クライアントでは、まずリクエストだけを先行してすべて送ります。 その後、結果を1つずつ読み込んで表示しています。 レスポンスをダンプするのにリクエストが必要なため、後から取得できるようにチャネルを使っています。

今回は簡易実装なので、POSTなどの安全ではない処理が混ざった場合の対処を省略しています。 ですが、パイプライニングのだいたいの雰囲気はつかめるでしょう。

