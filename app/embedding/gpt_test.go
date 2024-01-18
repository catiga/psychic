package embedding

import (
	"fmt"
	"testing"
)

func TestEmb(t *testing.T) {
	g := &GPT{}

	v, err := g.Query("", "query questions", map[string]string{
		"user": "0",
		"char": "1",
	}, 10)

	fmt.Println(v, err)
}
