package request_test

import (
	"fmt"
	"testing"
)

func TestImpl(t *testing.T) {
	var r interface{}
	fmt.Println(r)
	change(r, 1)
	fmt.Println(r)
	changeP(&r, 4)
	fmt.Println(r)
}
func change(i interface{}, v int) {
	i = v
}
func changeP(i *interface{}, v int) {
	*i = v
}
