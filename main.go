package main

import (
	"fmt"

	_ "github.com/lworkltd/kits/service/discovery"
	_ "github.com/lworkltd/kits/service/grpcinvoke"
	_ "github.com/lworkltd/kits/service/grpcsrv"
	_ "github.com/lworkltd/kits/service/invoke"
	_ "github.com/lworkltd/kits/service/monitor"
	_ "github.com/lworkltd/kits/service/profile"
	_ "github.com/lworkltd/kits/service/restful/code"
	_ "github.com/lworkltd/kits/service/restful/wrap"
	_ "github.com/lworkltd/kits/utils/co"
	_ "github.com/lworkltd/kits/utils/ipnet"
	_ "github.com/lworkltd/kits/utils/jsonize"
	_ "github.com/lworkltd/kits/utils/tags"
)

func main() {
	fmt.Println("nothing to be done")
}
