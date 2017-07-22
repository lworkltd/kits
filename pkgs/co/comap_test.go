package co

import (
	"fmt"
	"testing"
)

func TestComap(t *testing.T) {
	var m Comap
	m = New()
	m.Set(1, 2)
	fmt.Println(m.Format("xxx%key,%valuexxx"))
	m.Set(1, 3)
	fmt.Println(m.Format("xxx%key,%valuexxx"))
	m.Set(2, 5)
	m.Set(3, &map[int]string{1: "2323"})
	m.Set(&map[int]string{1: "2323"}, &map[int]string{1: "2323"})
	fmt.Println(m.Pairs())
	m.Set(12, 7)
	fmt.Println(m.SortedPairs(func(p1 Pair, p2 Pair) bool {
		return false
	}))

	fmt.Println(m.Keys())
	fmt.Println(m.Values())
}
