package main

import (
	"net/http"
	"os"
)

func main() {
	req, err := http.NewRequest("GET", "http://ascii.jp", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("X-TEST", "ADD")
	req.Write(os.Stdout)
}
