# 06 TCP

# プロトコルとレイヤー

通信を行うためには、送信側と受信側で通信のルールを共有する必要があります。このルールのことを「プロトコル（通信規約）」と呼びます。

プロトコルは、通常、役割に応じていくつかを組み合わせて使います。OSI 7階層モデルやTCP/IP参照モデルなどです。これらは、ネットワーク通信を実現するための様々な機能を階層に分け、それぞれのレイヤーを担うプロトコルを規定したものです。インターネット通信で採用されているのは、TCP/IP参照モデルです。TCP/IP参照モデルは以下のようになっています。

|レイヤーの名称|代表的なプロトコル|
|:----|:----|
|アプリケーション層|HTTP|
|トランスポート層|TCP/UDP/QUIC  |
|インターネット層|IP|
|リンク層|WiFi,イーサネット |

この内、アプリケーションを作る上で気にする必要があるのは、トランスポート層より上のレイヤです。実際のインターネット通信では、ケーブルや無線を通してIPぱけっとの形でデータがやり取りされますが、アプリケーションで直接IPパケットを作ったりはしません。HTTPやTCPで決められているルールに従って通信すれば、下のレイヤのことを気にしなくても、ネットワークの向こうにあるアプリケーションとやり取りが出来るようになっています。

## HTTPとその上のプロトコルたち

HTTP, GraphQL, RESTなど

## ソケットとは

HTTPがどのような仕組みで下位レイヤーを使っているのでしょうか。現在、ほとんどのOSではアプリケーション層からトランスポート層のプロトコルを利用する時のAPIとしてソケットという仕組みを利用しています。  
一般に、他のアプリケーションとの通信のことをプロセス間通信（IPC: Inter Process Communication）と呼びます。OSには、シグナル、メッセージキュー、パイプ、共有メモリなど、数多くのプロセス間通信機能が用意されています。ソケットもそのようなものの一種です。さっき紹介したプロセス間通信と違うのは、ローカルのコンピュータだけではなく、外部のコンピュータとも通信が行える点です。  
アプリケーション間のインターネット通信も、この祖kェッとを通じて行います。例えばブラウザの場合は、TCP80番に対して、ソケットを使ったプロセス間通信を行います。

### ソケット通信の基本構造

どんなソケット通信も基本となる構成は次のような形態です。

- サーバ(Listen())：ソケットを開いて待ち受ける
- クライアント(Dial())：開いているソケットに接続し、通信を行う

通信手順はプロトコルによって異なります。一方的な送信しか出来ないUDPのようなプロトコルもあれば、接続時にサーバーがクライアントを認知(Accept())して双方向にやり取りが出来るようになるTCPやUnixドメインソケットなどのプロトコルもあります。

1. Server Listen()で待受
2. Client Dial()で通信しようとする
3. Server Accept()してセッション成立させる
4. Server, Client Close()してセッション終了

TCP クライアントコード

```go
conn, err := net.Dial("tcp", "localhost:8080")
if err != nil {
    panic(err)
}
// connを使った読み書き
```

TCPサーバー側の最低限

```go
ln, err := net.Listen("tcp", ":8080")
if err != nil {
  panic(err)
}
conn, err := ln.Accept()
if err != nil {
  // handle error
}
// connを使った読み書き
```


上記の例だと、一度アクセスされたら終了してしまいます。1つのリクエストの処理中に他のリクエストを受け付ける。CPUが許す限り並列タスクをこなしたい。そのためには以下のようなコードを書くことになります。

```go
ln, err := net.Listen("tcp", ":8080")
if err != nil {
  panic(err)
}
// 一度で終了しないためにAccept()を何度も繰り返し呼ぶ
for {
  conn, err := ln.Accept()
  if err != nil {
    // handle error
  }
  // 1リクエスト処理中に他のリクエストのAccept()が行えるように
  // Goroutineを使って非同期にレスポンスを処理する
  go func() {
    // connを使った読み書き
  }()
}
```

## TCPソケットを使ったHTTPサーバー

HTTP1.0相当の送受信を実現

```go
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
)

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
		go func() {
			fmt.Printf("Accept %v\n", conn.RemoteAddr())
			// リクエストを読み込む
			request, err := http.ReadRequest(bufio.NewReader(conn))
			if err != nil {
				panic(err)
			}
			dump, err := httputil.DumpRequest(request, true)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(dump))
			// レスポンスを書き込む
			response := http.Response{
				StatusCode: 200,
				ProtoMajor: 1,
				ProtoMinor: 0,
				Body:       ioutil.NopCloser(strings.NewReader("Hello World\n")),
			}
			response.Write(conn)
			conn.Close()
		}()
	}
}
```

```console
Server is running at localhost:8888
Accept 127.0.0.1:58665
GET / HTTP/1.1
Host: localhost:8888
Accept: */*
User-Agent: curl/7.43.0
```

```console
❯ curl localhost:8888 -vvv
* Rebuilt URL to: localhost:8888/
*   Trying ::1...
* connect to ::1 port 8888 failed: Connection refused
*   Trying 127.0.0.1...
* Connected to localhost (127.0.0.1) port 8888 (#0)
> GET / HTTP/1.1
> Host: localhost:8888
> User-Agent: curl/7.43.0
> Accept: */*
>
* HTTP 1.0, assume close after body
< HTTP/1.0 200 OK
<
Hello World
* Closing connection 0

```

※実際にGoでHTTPのコードを実装する時は、net/httpという高機能APIを使います。低レベルなnetパッケージを使うことはほとんどありません。

### コード解説  

```go
request, err := http.ReadRequest(bufio.NewReader(conn))
if err != nil {
  panic(err)
}
```

まずはクライアントから送られてきたリクエストの読み込みです。自分でテキスト解析してもいいですが、http.ReadRequest()関数を使ってHTTPリクエストのヘッダー、メソッド、パスなどの情報を切り出します。

```go
dump, err := httputil.DumpRequest(request, true)
if err != nil {
  panic(err)
}
fmt.Println(string(dump))
```

読み込んだリクエストは、httputil.DumpRequest()関数を使って取り出しています。この関数はhttputil以下にある便利なデバック関数です。ここまでで、io.Readerからバイト列を読み込んで分析してデバック出力するという処理を行なっています。

```go
response := http.Response{
  StatusCode: 200,
  ProtoMajor: 1,
  ProtoMinor: 0,
  Body:       ioutil.NopCloser(strings.NewReader("Hello World\n")),
}
response.Write(conn)
```

次はHTTPリクエストを送信してくれたクライアント向けにレスポンスを生成するコードです。これには、http.Response構造体を使います。http.Response構造体はwrite()メソッドを持っているので、作成したレスポンスのコンテツをio.Writerに直接書き込むことが出来ます。

```go
conn.Close()
```

最後にClose()

## TCPソケットを使ってHTTPクライアント

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
	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}
	request, err := http.NewRequest("GET", "http://localhost:8888", nil)
	if err != nil {
		panic(err)
	}
	request.Write(conn)
	response, err := http.ReadResponse(bufio.NewReader(conn), request)
	if err != nil {
		panic(err)
	}
	dump, err := httputil.DumpResponse(response, true)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(dump))
}
```

## HTTP/1.1のKeep-Aliveに対応させる

HTTP1.0をシンプルに実装した前のコードでは、1セットの通信が終わる度に通信が切れます。  
HTTP1.1ではKeep-Aliveが規格に入りました。Keep-Aliveを使うことで、HTTP/1.0のように一つのメッセージごとに切断するのではなく、しばらくの間はTCP接続のセッションを維持して使います。 
TCPではセッションを接続するのに、1.5RTT（ラウンドトリップタイム:1往復の通信で1RTT）の時間がかかります。切断にも1.5RTTの時間がかかります。物理的な距離や速度によって1RTTのじかんも代わりますが、RTTは多いほど影響を与えます。一度の送信（送信と確認の返信で1RTT）につき、1.5 + 1.5 = 3RTTのオーバーヘッドがあれば、実行速度は単純に考えると1/4です。Keep-Aliveを使えば、この分のオーバーヘッドをなくせるので、速度低下を防げます。

### Keep-Alive対応のHTTPサーバー

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
		go func() {
			fmt.Printf("Accept %v\n", conn.RemoteAddr())
			for {
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
				content := "Hello world\n"
				// レスポンスを書き込む
				// HTTP1.1かつ、ContentLengthの設定が必要
				response := http.Response{
					StatusCode:    200,
					ProtoMajor:    1,
					ProtoMinor:    1,
					ContentLength: int64(len(content)),
					Body:          ioutil.NopCloser(strings.NewReader(content)),
				}
				response.Write(conn)
			}
			conn.Close()
		}()
	}
}
```

#### コード解説

```go
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go func() {
			fmt.Printf("Accept %v\n", conn.RemoteAddr())
			for {
        ///
      }
```

このコードで重要なのは、Accept()を受信した後にforループがある点です。これによって、コネクションがはられた後に何度もリクエストを受けられるようにしています。  

```go
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
```

タイムアウトの設定も重要です。これを設定しておくと、通信がしばらくないとタイムアウトのエラーでRead()の呼び出しを終了します。設定しなければ相手からのレスポンスがあるまでずっとブロックします。ここでは現在時刻から＋5秒としています。  

タイムアウトは、標準のerrインターフェースの上位互換であるnet.Errorインターフェースの構造体から取得出来ます。net.Connをbuifo.Readerでラップして、それをhttp.ReadRequest()関数に渡しています。  
タイムアウト時のエラーはnet.Connが生成しますが、それいがいのio.Readerは最初に発生したエラーをそのまま伝搬します。そのため、errorからダウンキャストを行うことでタイムアウトかどうかを判断できます。  

```go
content := "Hello world\n"
// レスポンスを書き込む
// HTTP1.1かつ、ContentLengthの設定が必要
response := http.Response{
  StatusCode:    200,
  ProtoMajor:    1,
  ProtoMinor:    1,
  ContentLength: int64(len(content)),
  Body:          ioutil.NopCloser(strings.NewReader(content)),
}
response.Write(conn)
```

HTTPバージョンを1.1に設定しています。  
送信するデータのバイト長が書き込まれている点もポイントです。Go言語のResponse.Write()はHTTP/1.1よりも前もしくは長さがわからない場合は、Connnection: closeヘッダーを付与してしまいます。複数のレスポンスを取り扱うには、明確にレスポンスを区切れる必要があります。

### Keep-Alive対応のHTTPクライアント

```go
package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
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
		dump, err := httputil.DumpResponse(response, true)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(dump))
		// 全送信完了していれば終了
		current++
		if current == len(sendMessages) {
			break
		}
	}
}
```

```go
Access: 0
HTTP/1.1 200 OK
Content-Length: 12

Hello world

HTTP/1.1 200 OK
Content-Length: 12

Hello world

HTTP/1.1 200 OK
Content-Length: 12

Hello world
```

#### コード解説

```go
var err error
// まだコネクションを張ってない / エラーでリトライ時はDialから行う
if conn == nil {
  conn, err = net.Dial("tcp", "localhost:8888")
  if err != nil {
    panic(err)
  }
  fmt.Printf("Access: %d\n", current)
}
```

サーバ同様、一度通信を開始したソケットをなるべき再利用します。

```go
// サーバから読み込む。タイムアウトはここでエラーになるのでリトライ
response, err := http.ReadResponse(bufio.NewReader(conn), request)
if err != nil {
  fmt.Println("Retry")
  conn = nil
  continue
}
```

サーバー側と異なるのは、通信の起点はソケットなので、セッションが切れた場合の再接続はクライアント側にあるという点です。切れた場合は、net.Conn型の変数を一度クリアして再試行するようになっています。
