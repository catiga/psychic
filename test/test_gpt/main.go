package main

import (
	"eli/app/embedding"
	"fmt"
)

func main() {
	g := &embedding.GPT{}

	v, err := g.Query("", "马丁靴是谁发明的，发明出来是干什么的", map[string]string{
		"user": "0",
		"char": "1",
	}, 10)

	fmt.Println(v, err)
}
