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
