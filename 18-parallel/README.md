# 18 Parallel

アムダールの法則によると、並列化して効率がどれだけ改善出来るかは、並列化できる仕事の割合をどれだけ増やせるかです。逐次処理をなるべく分解して、同じ粒度のシンプルなたくさんのジョブに分ける必要があります。

ジョブそのものをスレッドで高速化する方法もいくつもあります。 オライリー・ジャパンの「並行コンピューティング技法」には、MapReduceやソート、検索をマルチスレッドで高速化する方法が紹介されています。 今回はそれらのロジックよりも粒度の大きな話のみを取り扱います。

## 同期->非同期

並行・並列化の第一歩は、「重い処理」をタスクに分けることです。

```go
package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	inputs := make(chan []byte)

	go func() {
		a, _ := ioutil.ReadFile("a.txt")
		inputs <- a
	}()

	go func() {
		b, _ := ioutil.ReadFile("b.txt")
		inputs <- b
	}()
}
```

## 非同期->同期化

非同期化したら、どこかで同期化する必要があります。そうでないとGoの場合、main()関数の処理が終わったタイミングでタスクが残っていても処理が終了してしまいます。なので、いい感じにしよう。  
selectとか、sync.WaitGroupなど。

## タスク生成と処理を分ける

タスクを生成する側と処理する側とそれぞれProducer(生産者)、Consumer（消費者）と呼ぶ。

このパターンは、Go言語であれば、チャネルでProducerとConsumerを接続することで簡単に実現できます。 チャネルは、複数のgoroutineで同時に読み込みを行っても、かならず1つのgoroutineだけが1つの結果を受け取れます（消失したり、複製ができてしまうことはない）。 したがって、Consumer側の数を増やすことで、安全に処理速度をスケールできます。

プロセスをまたいでProducer-Consumerパターンを実現するには、一般にメッセージキューと呼ばれるミドルウェアで仲介します。 シンプルなものではbeanstalkd1という、メッセージキューのミドルウェアがあり、beanstalkd公式のGo用のクライアントライブラリ2が提供されています。

Amazon SQSのような、メッセージキューのクラウドサービスもあります。 負荷に応じてConsumerプロセスの起動まで面倒見てくれるものはサーバーレスアーキテクチャと呼ばれ、AWS/GCP/Azureで提供されています。

## 開始した順で処理する

07のTCPパイプライニングで実装した。

## タスク処理が詰まったら待機：バックプレッシャー

ネットワーク用語です。本来は、LANのスイッチにおいて、パケットが溢れそうになったら、送信側に衝突が発生したという信号を意図的に送り、送信料を落とさせる仕組みです。  

Goの場合はgoroutineの入力にバッファつきチャネルを使うことで、バックプレッシャーを実現出来ます。

平常時につまらない程度のサイズにするのが良いです。

```go
tasks := make(chan string, 10)
```

## 並列Forループ

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	tasks := []string{"cmake ..", "cmake . --build Release", "cpack"}
	var wg sync.WaitGroup
	wg.Add(len(tasks))
	for _, task := range tasks {
		go func(task string) {
			// ジョブ実行
			fmt.Println(task)
			wg.Done()
		}(task)
	}
	wg.Wait()
}
```

ループの内部の処理が小さすぎると、オーバーヘッドの方が大きくなるので注意してください。また、計算速度はCPUのコア数以上はスケールしないので、CPUの負荷が大きい場合は、次に紹介するgoroutineループを量する方が良いでしょう。

主にI/O待ちが多い場合に有効です。

## 決まった数のgoroutineでタスクを消化： ワーカープール

OSスレッドやフォークしたプロセスで多数の処理をこなすときは、生成コストの問題があるため、事前にワーカーをいくつか作ってストックしておき、そのワーカーが並列でタスクを消化していく方法がよく取られます。事前に作られたワーカー群のことを、スレッドプールとかプロセスプール、あるいはワーカープールなどと呼びます。

runtime.NumCPU()の個数分起動しているので、CPUを目一杯回すことができます。

```go
package main

import (
	"fmt"
	"runtime"
	"sync"
)

// 計算: 元金均等
func calc(id, price int, interestRate float64, year int) {
	months := year * 12
	interest := 0
	for i := 0; i < months; i++ {
		balance := price * (months - i) / months
		interest += int(float64(balance) * interestRate / 12)
	}
	fmt.Printf("year=%d total=%d interest=%d id=%d\n", year, price+interest, interest, id)
}

// ワーカー
func worker(id, price int, interestRate float64, years chan int, wg *sync.WaitGroup) {
	// タスクがなくなってタスクのチャネルがcloseされるまで無限ループ
	for year := range years {
		calc(id, price, interestRate, year)
		wg.Done()
	}
}
func main() {
	// 借入額
	price := 40000000
	// 利子 1.1%固定
	interestRate := 0.011
	// タスクはchanに格納
	years := make(chan int, 35)
	for i := 1; i < 36; i++ {
		years <- i
	}
	var wg sync.WaitGroup
	wg.Add(35)
	// CPUコア数分のgoroutine起動
	for i := 0; i < runtime.NumCPU(); i++ {
		go worker(i, price, interestRate, years, &wg)
	}
	// すべてのワーカーが終了する
	close(years)
	wg.Wait()
}
```

## 依存関係のあるタスクを表現する Future/Promise

依存関係のあるタスクをパイプラインとしてスマートに表現し、実行可能なタスクから効率よく消化していくことで遅延を短縮します。

- Future 今はまだ得られてないけど、将来得られるはずの入力
- Promise 将来、値を提供するという約束

```go
package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func readFile(path string) chan string {
	// ファイルを読み込み、その結果を返すFutureを返す
	promise := make(chan string)
	go func() {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("read error %s\n", err.Error())
			close(promise)
		} else {
			// 約束を果たす
			fmt.Println("約束果たした")
			promise <- string(content)
		}
	}()
	return promise
}

func printFunc(futureSource chan string) chan []string {
	// 文字列中の関数一覧を返すFutureを返す
	promise := make(chan []string)
	go func() {
		var result []string
		// futureが解決するまで待って実行
		fmt.Println("信じて待つ")
		for _, line := range strings.Split(<-futureSource, "\n") {
			fmt.Println("きたー！")
			if strings.HasPrefix(line, "func ") {
				result = append(result, line)
			}
		}
		// 約束を果たした
		promise <- result
	}()
	return promise
}

func main() {
	futureSource := readFile("main.go")
	futureFuncs := printFunc(futureSource)
	fmt.Println(strings.Join(<-futureFuncs, "\n"))
}
```

```console
信じて待つ
約束果たした
きたー！
きたー！
きたー！
きたー！
きたー！
きたー！
きたー！
きたー！
きたー！
きたー！
```

Futureでは、結果を1回でまとめて送ります（この点は後述のReactiveXとは異なります）。 サーバー越しに取得してきたファイルを小分けにして10回送る、といった処理のことは考えられていません。 上記の実装はシンプルなものなので、実用的なものにするためには、途中で中断されたことを把握できるように全てのジョブにContextを渡すといったことが必要になるでしょう

## イベントの流れを定義する: ReactiveX

Rx（オブザーバーパターンを賢くしたもの）

オブザーバーパターンでは、監視している（オブザーバブル）が変更されると、監視している（オブザーバー）に確実に漏れなく通知を行うのが責務でした。ReactiveXでは、イベントやデータストアのストリーム（流れ）を定義し、何度も頻繁に発生するイベントも取り扱えるように拡張されています。

```go
package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/reactivex/rxgo/observable"
	"github.com/reactivex/rxgo/observer"
)

func main() {
	//observableを作成
	emitter := make(chan interface{})
	source := observable.Observable(emitter)

	// イベントを受け取るobserverを作成
	watcher := observer.Observer{
		NextHandler: func(item interface{}) {
			line := item.(string)
			if strings.HasPrefix(line, "func ") {
				fmt.Println(line)
			}
		},
		ErrHandler: func(err error) {
			fmt.Printf("Encountered error: %v\n", err)
		},
		DoneHandler: func() {
			fmt.Println("Done!")
		},
	}

	// observableとobserverを接続（購読）
	sub := source.Subscribe(watcher)

	// observable
	go func() {
		content, err := ioutil.ReadFile("main.go")
		if err != nil {
			emitter <- err
		} else {
			for _, line := range strings.Split(string(content), "\n") {
				emitter <- line
			}
		}
		close(emitter)
	}()
	// 終了待ち
	<-sub
}
```

## 自立した複数のシステムで協調動作

アクターモデルは、Future/Promiseよりも古い、1973年に発表された並列演算モデルです。自立した多数の小さなコンピュータ（アクター）が協調して動作するというモデルになっています。各アクターは、別のアクターから送られてくるメッセージを受け取るメールボックスを持ち、そのメッセージをもとに協調動作します。各アクターは自律しており、並行動作するものとして考えます。

```go
package main

import (
	"fmt"

	console "github.com/AsynkronIT/goconsole"
	"github.com/AsynkronIT/protoactor-go/actor"
)

// メッセージ
type hello struct{ Who string }

// アクター
type helloActor struct{}

func (state *helloActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *hello:
		fmt.Printf("Hello %v\n", msg.Who)
	}
}

func main() {
	props := actor.FromInstance(&helloActor{})
	pid := actor.Spawn(props)
	pid.Tell(&hello{Who: "Roger"})
	pid.Tell(&hello{Who: "World"})
	console.ReadLine()
}
```

```console
Hello Roger
Hello World
```
