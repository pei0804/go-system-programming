# 08 UDP

# UDPが使われる場面は昔と今で変わってきている。

## 昔

UDPはTCPと同じトランスポート層ですが、TCPと違ってコネクションレス型です。誰と通信しているか、データの順番やデータロスの管理などは一切行いません。一方的に投げるだけというシンプルなプロトコルですが、そのかわりに、複数コンピュータに同時にメッセージ送信が可能なマルチキャストやブロードキャストをサポートしています。

UDPが使われているアプリケーション例  

- DNS
- NTP
- 動画配信
- 音声配信
- P2P

かつては、VPNなどの仮想ネットワークの土台にもUDPは使われていました。理由としては、仮想ネットワークでは、そこで張られたTCPコネクションがエラー訂正や順番制御を行うため。その接続プロトコルにTCPを使うと、TCP over TCPとなり、無駄が多かったためです。

UDPは高速といわれるので、独自プロトコルを開発する時に、UDPを土台として選ぶことがあります。伝送ロスがあまりないことが期待できる機内LAN専用高速プロトコルなど。

しかし、現在では以上のような使い分けは正しいとは言い切れません。

## 今

セキュリティ上の理由から、VPN接続でも暗号化のためにTLS経由するSSL-VPNが使われることが増えています。SSL-VPNにも3通りの方式があり、その中にはパケットをHTTPS上にくるんで送信するものがあります。この場合には、上で使うプロトコルがTCPの場合、どうしてもTCP over TCPとなります。  
独自プロトコルを開発するには、かなりの作り込みが必要となります。Googleでは独自にQUICというトランスポート層のプロトコルを開発していますが、彼らのようにネットワークを知り尽くした人たちが設計して大規模なフィールドテストが出来る状態でないと、そもそも独自プロトコルを作るべきでないと言えます。TCPも最近では輻輳制御は高性能になっています。  

[TCPの輻輳制御機能](http://ascii.jp/elem/000/001/411/1411547/#fn1)  

## どちらを選択するべきか

UDPが高速と言われる理由はコネクション接続時間がかからないからです。TCPでは1.5RTTの時間がかかりますが、UDPでは接続の時間は不要なので、短時間で完了するメッセージを短時間で大量に送受信する場合には、メリットが大きいでしょう。一方で、一度の通信で1パケットに収まらないような大きなデータをやり取りする場合、自前でエラー処理も含めて実装する場合、そこまで差はないでしょう。UDPが高速かどうかは通信するアプリケーションの性質に左右されます。

現在では一部を除いて、アプリケーションレイヤーで使われるプロトコルの多くがTCPを土台にしています。主にUDPを使っているDNSも512KBを超えるレスポンスの場合はTCPにフォールバックする仕組みがあったりします。他のUDPを使っていたアプリケーションもTCPを採用している例が増えています。特別な理由がない限りは、すぐに使えて安全性の高いTCPを採用する方がいいでしょう。

# UDPとTCPの処理の違い

1. サーバ ListenPacket()
2. クライント Dial()
3. クライアント Write()
4. サーバ ReadForm()

## サーバー実装

```go
package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("Server is running at localhost:8888")
	conn, err := net.ListenPacket("udp", "localhost:8888")
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

### コード解説

```go
fmt.Println("Server is running at localhost:8888")
conn, err := net.ListenPacket("udp", "localhost:8888")
if err != nil {
  panic(err)
}
defer conn.Close()
```

TCPソケットの場合、接続を受け付けるサーバはnet.Listen()を呼び出し、返ってきたnet.Listenerインターフェースでクライントが接続してくるのを待ち、接続されたら、Accept()メソッドを呼び、お互いのデータを送受信するためのnet.Connインターフェースオブジェクトを得たりしました。  

UDPでは、net.ListenPacket()を使って、クライントを待つのではなくて、データの送受信のためのnet.PacketConnというインターフェースが即座に返されます。このオブジェクトもio.Readerインターフェースを実装しているため、圧縮やファイル入出力などの高度なAPIと簡単に接続出来ます。

```go
length, remoteAddress, err := conn.ReadFrom(buffer)
if err != nil {
  panic(err)
}
fmt.Printf("Received from %v: %v\n", remoteAddress, string(buffer[:length]))
```

接続処理の注目すべきポイントはconn.ReadFrom()です。ReadFrom()メソッドを使うと、通信内容を読み込むと同時に、接続してきた相手のアドレス情報を受け取れます。net.PacketConnは、サーバ側でクライアントを知らない状態で開かれるソケットなので、このインターフェースを使ってサーバから先にメッセージを送信することは出来ません。サーバにクライアントから接続があったときに初めてクライアントのアドレスが分かります。通信内容だけを取得し、通信の内容を認識しないRead()メソッドを使ってしまうと、通信相手に通信を送り返す必要がある時に対処出来なくなってしまいます。  

```go
buffer := make([]byte, 1500)
```

ReadFrom()では、TCPの時の紹介したデータ終了を探りながら受信といった高度な読み込みは出来ません。そのため、データサイズが決まらないデータに対してはフレームサイズ分のバッファや最大サイズ分のバッファを作り、そこにデータをまとめて読み込むことになります。あるいは、バイナリ形式のデータにして、ヘッダにデータ長などを格納しておき、それを先読みしてから必要なバッファを確保して読み込むといったコードになるでしょう。

```go
_, err = conn.WriteTo([]byte("Hello from Server"), remoteAddress)
if err != nil {
  panic(err)
}
```

ReadFrom()で取得したアドレスに対しては、net.PacketConnインターフェースのWriteTo()メソッドを使ってデータを返送することが出来ます。

## クライアント実装

```go
package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("udp4", "localhost:8888")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("Sending to server")
	_, err = conn.Write([]byte("Hello from Client"))
	if err != nil {
		panic(err)
	}
	fmt.Println("Receiving from server")
	buffer := make([]byte, 1500)
	length, err := conn.Read(buffer)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received: %s\n", string(buffer[:length]))
}
```

```console
Server is running at localhost:8888
Received from 127.0.0.1:51493: Hello from Client
```

```console
Sending to server
Receiving from server
Received: Hello from Server
```

### コード解説

クライアント側ではDial()を使うので、TCPと同じようにio.Reader, io.Writerインターフェースをそのまま使うことが出来ます。

## コラム：抽象インターフェースと具象実装

UDPの通信サンプルをネットで調べると、ほとんどのコードはnet.Listen()やnet.Dial()ではなく、net.ListenUDP()やnet.DialUDP()という関数を使っています。結論から言うとどちらを使っても同じです。しかし、これから説明するようなUDPのマルチキャストや、TCPのKeepAliveなどはそれぞれの固有のプロトコルのものなので、それぞれの型の関数を明示的に使う必要があります。

```go
udp, ok := conn.(*net.UDPConn)
if ok {
  // UDP固有処理
}
```

上記の例で使っているnet.Listen()やnet.ListenPacket(), net.Dial()はプロトコルの種類を文字列で指定するだけで具体的なインターフェースを隠して通信を抽象的に書くためのインターフェースです。明示的な実装が必要な場合は、net.ListenUDP()やnet.ListenTCP()などの関数を使って通信してもいいし、net.Connやnet.PacketConnのインターフェースから具象型にキャストする方法もあります。

## UDPのマルチキャストの実装例

マルチキャストは、リクエスト側の負担を増やすことなく多くのクライアントに同時にデータを送信出来る仕組みです。

### サーバ実装

```go
package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("Start tick server at 224.0.0.1:9999")
	conn, err := net.Dial("udp", "224.0.0.1:9999")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	start := time.Now()
	wait := 10*time.Second - time.Nanosecond*time.Duration(start.UnixNano()%(10*1000*1000*1000))
	time.Sleep(wait)
	ticker := time.Tick(10 * time.Second)
	for now := range ticker {
		conn.Write([]byte(now.String()))
		fmt.Println("Tick: ", now.String())
	}
}
```

#### コード解説

例は時報のようなもので、あくまで例題なので、電話による時報を聞く場合と同様に遅延などによる誤差は気にしないものとします。

```go
fmt.Println("Start tick server at 224.0.0.1:9999")
conn, err := net.Dial("udp", "224.0.0.1:9999")
if err != nil {
  panic(err)
}
defer conn.Close()
```

UDPのマルチキャストでは、サービスをウケる側（クライアント）がソケットをオープンにして待受、そこにサービス提供者（サーバ）がデータを送信します。

```go
start := time.Now()
wait := 10*time.Second - time.Nanosecond*time.Duration(start.UnixNano()%(10*1000*1000*1000))
time.Sleep(wait)
ticker := time.Tick(10 * time.Second)
for now := range ticker {
  conn.Write([]byte(now.String()))
  fmt.Println("Tick: ", now.String())
}
```

10秒単位に端数を取り出しているだけです。そして、決まった時間に接続しているクライアントにマルチキャストでデータを流しています。

### クライアント実装

```go
package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("Listen tick server at 224.0.0.1:9999")
	address, err := net.ResolveUDPAddr("udp", "224.0.0.1:9999")
	if err != nil {
		panic(err)
	}
	listener, err := net.ListenMulticastUDP("udp", nil, address)
	defer listener.Close()
	buffer := make([]byte, 1500)
	for {
		length, remoteAddress, err := listener.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Server %v\n", remoteAddress)
		fmt.Printf("Now %s\n", string(buffer[:length]))
	}
}
```

```console
Start tick server at 224.0.0.1:9999
Tick:  2018-01-03 12:00:20.006369656 +0900 JST
Tick:  2018-01-03 12:00:30.002014987 +0900 JST
Tick:  2018-01-03 12:00:40.004010592 +0900 JST
Tick:  2018-01-03 12:00:50.005536195 +0900 JST
Tick:  2018-01-03 12:01:00.001898152 +0900 JST
Tick:  2018-01-03 12:01:10.006129842 +0900 JST
Tick:  2018-01-03 12:01:20.006278843 +0900 JST
Tick:  2018-01-03 12:01:30.006248287 +0900 JST
```

```console
Listen tick server at 224.0.0.1:9999
Server 192.168.0.170:59671
Now 2018-01-03 12:00:20.006369656 +0900 JST
Server 192.168.0.170:59671
Now 2018-01-03 12:00:30.002014987 +0900 JST
```

クライアント側で複数のネットワーク接続がある時に、特定のLAN環境のマルチキャストを受信するには、net.InterfaceByName("en0")のように書いて、イーサネットのインターフェース情報を取得して、それをnet.ListenMulticastUDP()の第二引数に渡す必要があります。

## TCPとUDPとの違い

### TCPには再送処理とフロー処理がある

TCPでは送信するメッセージにシーケンス番号が入っているので、受信側ではこの数値を見て、もしパケット順序が入れ替わっていた時は順序を並べ直します。そして、受信側は受け取ったデータのシーケンス番号とサイズの合計を確認応答番号として返します。ここで誤りがある場合は再送処理を行います。  
また、TCPにはウィンドウ制御という機能があり、受信側が用意出来ていない状態で送信リクエストが集中して通信内容が失われたりするのを防ぎます。具体的には、送受信用のバッファを予め決めておき、送信側はこのサイズまで受信側からの確認を待たずに送信が出来ます。このサイズは最初のコネクション確立時に決定されます。もし、受信側のデータの読み込み処理が間に合わない場合には、受信できるウィンドウサイズを受信側から送信側に伝えて送信量を制御することが出来ます。これをフロー制御といいます。  
UDPにはこれらのことを考えずに、一方的に送りつけるので、処理自体は高速になっています。もちろん自前で実装もできます。

### UDPではフレームサイズも気にしよう

TCPもUDPも、その下のデータリンク層の影響を受けます。ひとかたまりで送信出来るデータの大きさは、通信経路の種類やルータなどの設定によって変わり、ある経路でひとかたまりで送信出来るデータ条件のことをその経路の最大転送単位（MTU）といいます。  
一般的に使われるイーサネットのMTUは1500オクテットですが、現在の市販のルータで「ジャンボフレーム対応」と書かれているものだと、それ以上のサイズを扱えます。しかし、UDPやTCPヘッダ、PPP,VPNなどでカプセル化されるとヘッダーが増えると、実データで確保出来る容量も小さくなります。  
MTUに収まらないデータは、IPレベル（TCP/UDPの下のレイヤー）で複数のパケットに分割されます。これをIPフラグメンテーションと呼びます。IPフラグメンテーション自体は、IPレイヤーで再結合をしてくれますが、分割された最後のパケットが来るまで、UDPパケットとして未完成のままあので、アプリケーション側にデータが流れてくることがありますせん。しかもデータの消失などが起きたら、タイムアウトが発生します。なので、UDPの売りである応答性の高さをカーネル内部の結合待ちで無駄にしないためには、イーサネットのフレームサイズを意識したアプリケーションプロトコルの実装が必須でしょう。  
具体的には、UDPを利用する場合には、データ構造を1フレームで収まるサイズにし、毎フレームにフレームの内容を識別するヘッダーをつける必要があるでしょう。またデータが欠落しても支障がないストリーミングデータであっても、順序ぐらいは守りたいでしょうから、何らかのカウンターが必要だったり、フレームに収まらないデータを格納するための仕組みも必要になるでしょう。  
また、巨大なデータをUDPとして送信するデメリットはもう一つあります。IPレイヤーでデータを結合してくれるとはいっても、IPレイヤーやその上のUDPレイヤーで取り扱えるデータは約64キロバイトまでなので、それ以上になるとパケットを分割する必要があります。TCPであれば気にする必要はありませんが、UDPだと対策が必要です。逆に言うと64キロバイト以下に収める場合は、データの取扱がシンプルになります。

### 輻輳制御とフェアネス

輻輳制御とは、ネットワークの輻輳（渋滞）を避けるように流量を調整して、そのネットワークの最大効率で通信できるようにするとともに、複数の通信をお互いにフェアに行える仕組みです。

TCPには輻輳制御は備わっており、そのアルゴリズムには様々な種類があります。どのアルゴリズムもゆっくり通信料を増やしていき、通信量の限界値をさぐりつつ、パケット消失などの渋滞発生を検知すると、流量を絞ったり増やしたりしながら、最適な通信料を探ります。最初の通信の限界を探る段階では、2倍、4倍、8倍と指数的に増やしていきます。このステップをスロースタートと呼びます。このような仕組みで最大速度を全体で出すことが出来るTCPですが、UDPは気にせず投げまくるので、輻輳を気にするTCPだけ速度が落ち込むといったことが起きたりします。
