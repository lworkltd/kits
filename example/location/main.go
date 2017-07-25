package main

import "github.com/lvhuat/kits/example/location/conf"
import "github.com/lvhuat/kits/example/location/api/server"

func main() {
	if err := conf.Parse(); err != nil {
		panic(err)
	}
	server.Setup(conf.GetService())
}
