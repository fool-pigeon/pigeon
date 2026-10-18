// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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
