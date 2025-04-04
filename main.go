package main

import (
	"fmt"
	"github.com/BoweFlex/vincent-vimgo/rope"
)

func main() {
	testString := "Hi This is geeksforgeeks. "
	testRope := rope.NewRope(nil, nil, testString)
	string2 := "I like to move it, move it. "
	rope2 := rope.NewRope(nil, nil, string2)
	fmt.Println(testRope.GetString())
	fmt.Println(rope2.GetString())
	combined := rope.ConcatRopes(&testRope, &rope2)
	fmt.Println(combined.GetString())
}
