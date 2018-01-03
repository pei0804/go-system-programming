# Unixドメインソケット

UnixドメインソケットはPOSIX系OSで提供されている機能です。コンピュータ内部でしか使えない代わりに、高速に通信が行えます。TCP型（ストリーム型）とUDP型（データグラム型）の両方の使い方が出来ます。

## 基本

TCPとUDPによるソケット通信は、外部のネットワークに繋がるインターフェースに接続します。これに対し、Unixドメインソケットでは外部インターフェースへの接続を行いません。その代わり、カーネル内部で完結する高速なネットワークインターフェースを作成します。Unixドメインソケットを使うことで、ウェブサーバーとNginxなどのリバースプロキシとの間、あるいはウェブサーバとの間の接続を高速に出来る場合があります。  

クライアント側のコード構成  
```go
conn, err := net.Dial("unix", "socketfile")
if err != nil {
  panic(err)
}
// connを使った処理
```

サーバ側のコード構成  
```go
listener, err := net.Listen("unix", "socketfile")
if err != nil {
  panic(err)
}
defer listener.Close()
conn, err := listener.Accept()
if err != nil {
  panic(err)
}
// connを使った処理
```

TCPとの違いとしての注意すべき点として、サーバ側でnet.Listener.Close()を呼ばないとソケットファイルが残り続けてしまうことが挙げられます。Ctrl+cで停める場合には、何らかのシグナルをトラップする必要があります。今回はとりあえず、ソケットファイルがある場合は削除するようにしてください。きちんとCloseしたい場合は、graceful shutdownあたりを検索すれば実現方法が見つかります。

## ストリーム型のUnixドメインソケット

### クライアント

```go
package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
)

func main() {
	conn, err := net.Dial("unix", filepath.Join(os.TempDir(), "unixdomainsocket-sample"))
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

### サーバ

```go
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	path := filepath.Join(os.TempDir(), "unixdomainsocket-sample")
	os.Remove(path)
	listener, err := net.Listen("unix", path)
	if err != nil {
		panic(err)
	}
	fmt.Println("Server is running at " + path)
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
Server is running at /var/folders/dh/c6l
wd3680000gn/T/unixdomainsocket-sample
Accept
GET / HTTP/1.1
Host: localhost:8888
User-Agent: Go-http-client/1.1
```

```console
HTTP/1.0 200 OK
Connection: close

Hello World


```

## データグラム型のUnixドメインソケット

### サーバ実装

```go
package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func main() {
	path := filepath.Join(os.TempDir(), "unixdomainsocket-server")
	// エラーチェックは削除（存在しなかったらしなかったで問題ないため）
	os.Remove(path)
	fmt.Println("Server is running at " + path)
	conn, err := net.ListenPacket("unixgram", path)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	buffer := make([]byte, 1500)
	for {
		length, remoteAddress, err := conn.ReadFrom(buffer)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Received from %v: %v\n", remoteAddress, string(buffer[:length]))
		_, err = conn.WriteTo([]byte("Hello from Server"), remoteAddress)
		if err != nil {
			panic(err)
		}
	}
}
```

### クライアント実装

```go
package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func main() {
	clientPath := filepath.Join(os.TempDir(), "unixdomainsocket-server")
	// エラーチェックは不要なので削除（存在しなかったらしなかったで問題ないため）
	os.Remove(clientPath)
	conn, err := net.ListenPacket("unixgram", clientPath)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// 送信先アドレス
	unixServerAddr, err := net.ResolveUnixAddr("unixgram", filepath.Join(os.TempDir(), "unixdomainsocket-server"))
	var serverAddr net.Addr = unixServerAddr
	fmt.Println("Sending to server")
	_, err = conn.WriteTo([]byte("Hello from Client"), serverAddr)
	if err != nil {
		panic(err)
	}
	fmt.Println("Receiving from server")
	buffer := make([]byte, 1500)
	length, _, err := conn.ReadFrom(buffer)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received: %s\n", string(buffer[:length]))
}
```

### 注意点

net.Dial()で開いたソケットは一方的な送信用で、アドレスと結び付けられていないので、返信を受けられません。なので、クライアント側もサーバと同じ初期化を行いnet.PacketConnインターフェースのWriteTo()メソッドとReadFrom()メソッドを使えば送受信が出来ます。
