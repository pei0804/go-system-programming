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

```go
package main

import (
    "bytes"
    "encoding/binary"
    "fmt"
)

func main() {
    // 32ビットのビッグエンディアンのデータ（10000）
    data := []byte{0x0, 0x0, 0x27, 0x10}
    var i int32
    // エンディアンの変換
    binary.Read(bytes.NewReader(data), binary.BigEndian, &i)
    fmt.Printf("data: %d\n", i)
}
```

## PNGファイルを分析してみる

PNGファイルはバイナリフォーマットです。先頭の8バイトがシグニチャ（固定のバイト列）となっています。それ以降はチャンクのブロックで構成されています。  

|長さ|種類|データ|CRC（誤り検知記号）|
|:---|:---|:-----|:------------------|
|4バイト|4バイト|長さで指定されたバイト数|4バイト|

書くチャンクとその長さを列挙する。  
readChunks()関数でチャンクごとにio.SectionReaderを作って配列に書くにうしています。それを表示する関数dumpChunk()で表示しています。

- [PNGフォーマット](http://www.setsuki.com/hsp/ext/png.htm)
- [IHDRヘッダ](http://www.setsuki.com/hsp/ext/chunk/IHDR.htm)

```go
package main

import (
    "encoding/binary"
    "fmt"
    "io"
    "os"
)

func dumpChunk(chunk io.Reader) {
    var length int32
    binary.Read(chunk, binary.BigEndian, &length)
    buffer := make([]byte, 4)
    chunk.Read(buffer)
    fmt.Printf("chunk '%v' (%d bytes)\n", string(buffer), length)
}

func readChunks(file *os.File) []io.Reader {
    // チャンクを格納する配列
    var chunks []io.Reader

    // 最初の8バイトを飛ばす
    file.Seek(8, 0)
    var offset int64 = 8

    for {
        var length int32
        err := binary.Read(file, binary.BigEndian, &length)
        if err == io.EOF {
            break
        }
        chunks = append(chunks, io.NewSectionReader(file, offset, int64(length)+12))
        // 次のチャンクの先頭に移動
        // 現在位置は長さを読み終わった箇所なので
        // チャンク名(4バイト) + データ長 + CRC(4バイト)先に移動
        offset, _ = file.Seek(int64(length+8), 1)
    }
    return chunks
}

func main() {
    file, err := os.Open("Lenna.png")
    if err != nil {
        panic(err)
    }
    chunks := readChunks(file)
    for _, chunk := range chunks {
        dumpChunk(chunk)
    }
}
```

```console
chunk 'IHDR' (13 bytes)
chunk 'sRGB' (1 bytes)
chunk 'IDAT' (473761 bytes)
chunk 'IEND' (0 bytes)
```

#### PNG画像に秘密のテキストを入れる

PNGにはテキストを追加するためのtEXtというチャンクが存在しています。そこにASCII PROGRAMING++というてきすとを入れてみます。

```go
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

func textChunk(text string) io.Reader {
	byteData := []byte(text)
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, int32(len(byteData)))
	buffer.WriteString("tEXt")
	buffer.Write(byteData)
	// CRCを計算して追加
	crc := crc32.NewIEEE()
	io.WriteString(crc, "tEXt")
	binary.Write(&buffer, binary.BigEndian, crc.Sum32())
	return &buffer
}

func readChunks(file *os.File) []io.Reader {
	// チャンクを格納する配列
	var chunks []io.Reader

	// 最初の8バイトを飛ばす（シグネチャ）
	file.Seek(8, 0)
	var offset int64 = 8

	for {
		var length int32
		// 長さを取得する
		err := binary.Read(file, binary.BigEndian, &length)
		if err == io.EOF {
			break
		}
		chunks = append(chunks, io.NewSectionReader(file, offset, int64(length)+12))
		// 次のチャンクの先頭に移動
		// 現在位置は長さを読み終わった箇所などで、
		// チャンク名（4バイト） + データ長 + CRC(４バイト) 先に移動
		// length + 8だと、8バイト足りないように感じるが、
		//元々既に8バイト飛ばしているので加算する形になる
		offset, _ = file.Seek(int64(length+8), 1)
	}
	return chunks
}

func dumpChunk(chunk io.Reader) {
	var length int32
	binary.Read(chunk, binary.BigEndian, &length)
	buffer := make([]byte, 4)
	chunk.Read(buffer)
	fmt.Printf("chunk '%v' (%d bytes)\n", string(buffer), length)
	if bytes.Equal(buffer, []byte("tEXt")) {
		rawText := make([]byte, length)
		chunk.Read(rawText)
		fmt.Println(string(rawText))
	}
}

func main() {
	file, err := os.Open("Lenna.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	newFile, err := os.Create("Lenna2.png")
	if err != nil {
		panic(err)
	}
	defer newFile.Close()
	chunks := readChunks(file)
	// シグニチャ書き込み
	io.WriteString(newFile, "\x89PNG\r\n\x1a\n")
	// 先頭に必要なIHDRチャンクを書き込み
	io.Copy(newFile, chunks[0])
	// テキストチャンクを追加
	io.Copy(newFile, textChunk("ASCII PROGRAMMING++"))
	// 残りのチャンクを追加
	for _, chunk := range chunks[1:] {
		io.Copy(newFile, chunk)
	}
	chunks = readChunks(newFile)
	for _, chunk := range chunks {
		dumpChunk(chunk)
	}
}
```

## テキストを解析する

## 改行・単語で区切る

全部読み込んでから分割する方法もありますが、bufio.Readerを使うことで、任意の文字で分割することが出来ます。

```go
package main

import (
    "bufio"
    "fmt"
    "strings"
)

var source = `1行目
2行目
3行目`

func main() {
    reader := bufio.NewReader(strings.NewReader(source))
    for {
        line, err := reader.ReadString('\n')
        fmt.Printf("%#v\n", line)
        break
        }
    }
}
```

終端を気にせず短く書く

```go
func main() {
    scanner := bufio.NewScanner(strings.NewReader(source))
    for scanner.Scan() {
        fmt.Printf("%#v\n", scanner.Text())
    }
}
```

注意点として、bufio.Readerの結果の行の末尾に改行コードが残っています。  
もし改行コード以外で区切りたい場合は、

```go
// 分割処理を単語区切りに設定
scanner.Split(bufio.ScanWords)
```

## データ型を指定して解析

io.Readerから読み込んだデータは、単なるバイト列か文字列としてしか扱っていません。それらを整数や浮動小数点数などに変換するには、fmt.Fscanを使います。  
※fmt.Fscaはスペース区切りが前提です。改行区切りの場合は、fmt.Fscanlが使えます。

```go
package main

import (
	"fmt"
	"strings"
)

var source = "123 1.234 1.0e4 test"

func main() {
	reader := strings.NewReader(source)
	var i int
	var f, g float64
	var s string
	fmt.Fscan(reader, &i, &f, &g, &s)
	fmt.Printf("i=%#v f=%#v g=%#v s=%#v\n", i, f, g, s)
}
```

任意のフォーマット

```go
fmt.Fscanf(reader, "%v, %v, %v, %v", &i, &f, &g, &s)
```

## その他の決まったフォーマットの文字列の解析

encodingパッケージの傘下にある機能を使えば様々なテキストを扱えます。

CSVファイルのパース

```go
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

var csvSource = `13101,"100  ","1000003","ﾄｳｷｮｳﾄ","ﾁﾖﾀﾞｸ","ﾋﾄﾂﾊﾞｼ(1ﾁｮｳﾒ)","東京都","千代田区","一ツ橋（１丁目）",1,0,1,0,0,0
13101,"101  ","1010003","ﾄｳｷｮｳﾄ","ﾁﾖﾀﾞｸ","ﾋﾄﾂﾊﾞｼ(2ﾁｮｳﾒ)","東京都","千代田区","一ツ橋（２丁目）",1,0,1,0,0,0
13101,"100  ","1000012","ﾄｳｷｮｳﾄ","ﾁﾖﾀﾞｸ","ﾋﾋﾞﾔｺｳｴﾝ","東京都","千代田区","日比谷公園",0,0,0,0,0,0
13101,"102  ","1020093","ﾄｳｷｮｳﾄ","ﾁﾖﾀﾞｸ","ﾋﾗｶﾜﾁｮｳ","東京都","千代田区","平河町",0,0,1,0,0,0
13101,"102  ","1020071","ﾄｳｷｮｳﾄ","ﾁﾖﾀﾞｸ","ﾌｼﾞﾐ","東京都","千代田区","富士見",0,0,1,0,0,0
`

func main() {
	reader := strings.NewReader(csvSource)
	csvReader := csv.NewReader(reader)
	for {
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		fmt.Println(line[2], line[6:9])
	}
}
```

## ストリームを自由に操るio.Reader/io.Writer

C++などにあるインターフェースを使った入出力の機構を、「ストリーム」と呼んでいます。  
Go言語ではストリームとは言いませんが、io.Readerとio.Writerをでーたが流れるパイプとして使うことが出来ます。  

- io.MultiReader
- io.TeeReader
- io.LimitReader
- io.SectionReader
- io.Pipe(io.PipeReader io.PipeWriter)

### io.MultiReader

引数で渡されたio.Readerの全ての入力が繋がっているかのように動作します。

```go
package main

import (
	"bytes"
	"io"
	"os"
)

func main() {
	header := bytes.NewBufferString("----- HEADER -----\n")
	content := bytes.NewBufferString("Example of io.MultiReader\n")
	footer := bytes.NewBufferString("----- FOOTER -----\n")

	reader := io.MultiReader(header, content, footer)
	io.Copy(os.Stdout, reader)
}
```

```console
----- HEADER -----
Example of io.MultiReader
----- FOOTER -----

[Process exited 0]
```

### io.TeeReader

読み込まれた内容を別のio.Writerに書き出します。  

```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

func main() {
	var buffer bytes.Buffer
	reader := bytes.NewBufferString("Example of io.TeeReader\n")
	teeReader := io.TeeReader(reader, &buffer)

	// データを読み捨てる
	_, _ = ioutil.ReadAll(teeReader)

	// けど、バッファに残ってる
	fmt.Println(buffer.String())
}
```
