package main

import (
	"github.com/minor-industries/theheads/common/discovery"
	"github.com/minor-industries/theheads/web"
)

func main() {
	web.Run(discovery.NewSerf("127.0.0.1:7373"))
}
