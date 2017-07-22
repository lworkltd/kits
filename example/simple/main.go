package main

import "github.com/lvhuat/kits/example/simple/conf"

func main() {
	if err := conf.Parse(); err != nil {
		panic(err)
	}
}
