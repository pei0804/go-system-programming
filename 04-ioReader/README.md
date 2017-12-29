# 04 io.Reader

# バイナリ解析用のio.Reader関連機能

## 必要な部分を切り出す io.LimitReader / io.SectionReader

ファイルの先頭にはヘッダー領域があって、そこだけ解析するルーチンに処理を渡したい時には、io.LimitReaderを使うと、データがたくさん入っていても、先頭の部分までしか読み込めないようにブロックしてくれます。

```go
// 先頭16バイトしか読み込まないようにする
lReader := io.LimitReader(reader, 16)
```

長さだけではなく、スタートイチも固定したいことがあります。PNGファイルやOpenTypeフォントなど、バイナリファイル内がいくつかのチャンクに分かれている場合は、チャンクごとにReaderをわけて読み込むことで、別々のチャンクを読み込むコード間の独立性が高まり、全体としてメンテナンスしやすいコードになります。そんな時に便利なのが、io.SectionReaderです。
※io.SectionReaderはio.Readerが使えず、io.ReaderAtが使えます。os.File型はio.ReaderAtを満たしますが、それ以外のio.Readerを満たす型からio.SectionReaderでは直接読み込みできません。

```go
package main

import (
    "io"
    "os"
    "strings"
)

func main() {
    reader := strings.NewReader("Example of io.SectionReader\n")
    sectionReader := io.NewSectionReader(reader, 14, 7)
    io.Copy(os.Stdout, sectionReader)
}
```

## エンディアン解析

バイナリ解析ではエンディアン変換が必要になります。現在主流のCPUはリトルエディアンです。（サーバーや組み込み機器でビックエディアンのものもあります）  
リトルエディアンは、10000という数値（16進数表記で0x2710）があった時に、小さい桁からメモリに格納されます。（Goげんごで書くと、[]byte{0x10, 0x27, 0x0, 0x0}と表現）  
しかし、ネットワーク上で転送されるデータの多くは、大きい桁からメモリに格納されるビックエディアン（ネットワークバイトオーダーともいう）です。そのため多くの環境で、ネットワークで受け取ったデータをリトルエディアンに修正する必要があります。

任意のエディアンの数値を、現在の実行環境のエディアンの数値に修正するには、encoding/binaryパッケージを使います。このパッケージのbinary.Read()メソッドに、io.Readerとデータのエディアン、それに変換結果を格納する変数のポインタを渡せばエディアンが修正されたデータが得られます。


