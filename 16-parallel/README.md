# 16 Parallel

# 複数のしごとを同時に行うとは?

複数の仕事を行うことを表す言葉には、並行と並列の２つがありますが、これらには次のような区別があります。

## 並行処理（Concurrent）

並行（Concurrent）: CPU数・コア数の限界を超えて複数のしごとを同時に行うこと

身近にある並行処理は、パソコンでブラウザを見ながら、スライド作成が出来るとかです。シングルコアで並行処理をする場合、トータルでのスループットは変わりません。スループットが変わらないのに並行処理が必要なのは、とりかかっている１つの仕事でプログラム全体がブロックされてしまうのを防ぐためです。

## 並列処理（Parallel）

並列(Parallel): 複数のCPU、コアを効率良く扱って計算速度を挙げることです。

並列は、CPUのコアが複数あるコンピュータで、効率よく計算処理を行う時に必要な概念です。例えば8コアのCPUが8つ同時に100%稼働すると、トータルのスループットが8倍になります。

## どちらが大切？

タスクによって並列と並行を両方とも考慮することが、初めて効率化を最大に出来ます。CPUにおける処理時間が大きい場合（ユーザー時間が支配的な場合）は並列、I/O待ちなどでCPUが暇をしている時は並行で処理するというのが基本です。

ユーザー時間：アプリケーションプログラム自体が直接ＣＰＵを使っている時間。計算時間など。

## Go言語の並列処理のための道具

Go言語には並列処理を簡単に書くための道具が備わっています。

- goroutine
- チャネル
- select

### goroutine

```go
package main

import (
	"fmt"
	"time"
)

func sub() {
	fmt.Println("sub() is running")
	time.Sleep(time.Second)
	fmt.Println("sub() is finished")
}

func main() {
	fmt.Println("start sub()")
	go sub()
	time.Sleep(2 * time.Second)
	fmt.Println("finish main()")
}
```

上記の例では、関数をgoを付けて呼び出していますが、Go言語では無名関数（クロージャ）が作れるので、次のように無名関数の作成とgoroutine化を同時に行うことが出来ます。

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("start sub()")
	// インラインで無名関数を作ってその場でgoroutineで実行
	go func() {
		fmt.Println("sub() is running")
		time.Sleep(time.Second)
		fmt.Println("sub() is finished")
	}()
	time.Sleep(2 * time.Second)
}
```

#### gorutineと情報共有

```go
package main

import (
	"fmt"
	"time"
)

func sub1(c int) {
	fmt.Println("share by args", c*c)
}

func main() {
	// 引数渡し
	go sub1(10)

	// クロージャのキャプチャ渡し
	c := 20
	go func() {
		fmt.Println("share by args", c*c)
	}()
	time.Sleep(time.Second)
}
```

クロージャのキャプチャ渡しの場合、内部的には、無名関数に暗黙の引数が追加され、その暗黙の引数にデータや参照が渡されgoroutineとして扱われます。つまり、関数実行して引数で渡すのと同じことです。

関数に引数として渡す方法と、クロージャのローカル変数にキャプチャして渡す方法との間で、１つ違いがあるとするなら、次のようなforループ内でgoroutineを起動する場合です。

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	tasks := []string{
		"cmake ...",
		"cmake . --build Release",
		"cpack",
	}
	for _, task := range tasks {
		go func() {
			// goroutineが起動するときにはループが回りきって
			// 全部のtaskが最後のタスクになってしまう
      // goroutineの起動はループに比べると遅いため
			fmt.Println(task)
		}()
	}
	time.Sleep(time.Second)
}
```

```console
cpack
cpack
cpack
```

goroutineの起動はOSのネイティブスレッドよりは高速ですが、それでもコスト０ではありません。ループの変数は使いまわされてしまいますし、単順なループに比べてgoroutineの起動が遅いため、クロージャを使ってキャプチャするとループが回るたびにプログラマーが意図したのとは別のデータを参照してしまいます。その場合は関数の引数経由にして明示的に値コピーが行われるようにします。

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	tasks := []string{
		"cmake ...",
		"cmake . --build Release",
		"cpack",
	}
	for _, task := range tasks {
		go func(task string) {
			// goroutineが起動するときにはループが回りきって
			// 全部のtaskが最後のタスクになってしまう
			// goroutineの起動はループに比べると遅いため
			fmt.Println(task)
		}(task)
	}
	time.Sleep(time.Second)
}
```

子供のgoroutineから親へは、ｈ奇数やクロージャで渡したデータ構造（配列やマップ、チャネルなど）に書き込む、あるいはクロージャでキャプチャした変数（キャプチャはポインタを引数に渡した扱いになる）に書き込むことになります。マップ要素へのアクセスはアトミックではないため注意が必要です。同時に書き込むと予期せぬ上書きが発生する可能生があるため、何らかの形で同時上書きを防ぐ必要があります。

一番単純な方法は、書き込み先を共有しないことです。例えば、10個のgorutineを同時に実行するとき、最初から10個文の結果を保存する配列を用意しておいて、それぞれのgorutineから別の領域に書き込むようにする方法があります。それ以外にもチャネルなどsyncもあります。

### チャネル

チャネルの状態とふるまい

|状態|バッファなしチャネル|バッファ付きチャネル|閉じたチャネル|
|:-----|:-------------------|:-------------------|:-------------| 
|作り方|make(chan 型)|make(chan 型, 個数)|close(既存のチャネル)|
|チャネル<-値で送信|受け取り側で受信操作をするまで停止|バッファがあれば即終了。無ければ左と同じ|パニック|
|変数 := <- チャネルで受信|送信側がデータを入れるまで停止|送信側がデータを入れるまで停止|デフォルト値を返す|
|変数, ok := <- チャネルで受信|同上+okにtrueが入る|同上+okにtrueが入る|同上+okにfalseが入る|
|for 変数 := range チャネル 受信|チャネルに値が入る度にループが回る|チャネルに値が入る度にループが回る|ループから抜ける|

```go
// バッファなし
tasks := make(chan string)
// バッファつき
tasks := make(chan string, 10)
```

チャネルへデータを送信したり、チャネルからデータを受信するには、下記のように、<-演算子を使います。

```go
// データを送信
tasks <- "cmake .."
tasks <- "cmake . --build Debug"

// データを受取
task := <-tasks
// データ受取＆クローズ判定
task, ok := <-tasks
// データを読み捨てる
<-wait
```

読み込みは基本的に送信側が送信するまでブロックします。配列を使う場合のおうに、len()を使ってチャネルに入ったデータ数を確認し、データが入っている時だけ読み込むというコードにすれば、ブロックさせないことは可能です。しかし、その方法だと読み込み側が並列になった時にスケールしません。そのためselectを使う方が良いでしょう。

またチャネルはforループにたいして使うことも出来ます。

```go
for task := range tasks {
      // タスクチャネルにデータが投入される限りループが続く
}
```

先程のタイマーを使った待ちをチャネルを使って書き換えました。

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("start sub()")
	// 終了を受け取るためのチャネル
	done := make(chan bool)
	go func() {
		fmt.Println("sub()is finished")
		time.Sleep(time.Second)
		// 終了を通知
		done <- true
	}()
	<-done
	fmt.Println("all tasks are finished")
}
```

```console
start sub()
sub()is finished
all tasks are finished
```

Go言語のバージョン1.7から入ったcontextパッケージによるコンテキストを使った方法です。深いネストの中、あるいは派生ジョブとかが複雑なロジックの中でも、正しく終了やキャンセル、タイムアウトが実装出来るようになっています。  
context.WithCancel()以外には、終了時間を設定したりタイムアウトの期限を設定できるcontext.WithDeadline()やcontext.WithTimeout()もあります。

```
package main

import (
	"context"
	"fmt"
)

func main() {
	fmt.Println("start sub()")
	// 終了を受け取るための終了関数付きコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		fmt.Println("sub()is finished")
		// 終了を通知
		cancel()
	}()
	<-ctx.Done()
	fmt.Println("all tasks are finished")
}
```

### select文

特に問題が起きにくい

- 多対1 -> 書き込み
- 1対多 -> 読み込み

終了フラグを先に読み込むと止まる可能生がある

- 多対1 -> 読み込み
- 1対多 -> 書き込み

ブロックしうる複数のチャネルを同時に並列で読み込み、最初の読み込めたものを処理するにはselect文が有効です。

Go言語のselect文の基本的な使い方は、下記コードのようになります。selectは一度トリガーすると終わってしまうため、forループでくくって使われることがほとんどです。case文には必要な数だけチャネルの読み込みコードを列挙します。変数を書くと読み込んだ値も取得出来ます。こちらの構文を使うとどれかのチャネルが応答があるまでブロックし続けます。

```go
for {
  select {
    case data := <-reader:
    // 読み込んだデータを利用
    case <-exit:
    // ループを抜ける
    break
  }
}
```

下記のようにdefault節を書くと、何も読み込めなかった時にその節が実行されます。こちらの構文の場合はブロックせずに終了します。チャネルにデータが入るまでポーリングでループを回したい場合に使えます。

## 並列・並行処理の手法パターン

|手法                           |マルチプロセス            |イベント駆動            |マルチスレッド                  |
|:------------------------------|:-------------------------|:-----------------------|:-------------------------------|
|特徴                           | スクリプト言語でも使える |I/O待ちが思い時に最適|性能が高い|
|複数のタスクを同時に行う(並行) |◯                         |◯                       | ◯                              |
|複数のコアを使う(並列)         |◯                        |×                       | ◯                           | 
|起動コスト                     |×                         |◯                       | △                           | 
|情報共有コスト                 |×                         |◯                       | ◯                           | 
|メモリ安全性                   |◯                         |△                       | ×                           | 

