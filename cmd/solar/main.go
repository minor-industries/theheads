package main

import (
	"fmt"
	"github.com/minor-industries/theheads/solar"
	"os"
)

func main() {
	err := solar.Run()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s", err.Error())
		os.Exit(-1)
	}
}
