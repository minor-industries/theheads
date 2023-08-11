package main

import (
	"fmt"
	"github.com/minor-industries/grm"
	"path/filepath"
)

func main() {
	grm.Main(map[string]func(rule string){
		"protos": func(rule string) {
			protoFiles, err := filepath.Glob("protos/*.proto")
			if err != nil {
				panic(err)
			}

			args := []string{
				"/bin/protoc",
				"--proto_path=./protos",
				"-I/build/include",
				"--go_out=plugins=grpc,paths=source_relative:./gen/go/heads",
			}

			for _, file := range protoFiles {
				// this may run into trouble if there are two proto files with the same name in
				// different directories
				base := filepath.Base(file)
				opt := fmt.Sprintf("--go_opt=M%s=github.com/cacktopus/theheads/camera/gen/go/heads", base)
				args = append(args, opt)
			}

			args = append(args, protoFiles...)

			grm.RunDocker("heads-protoc", args...)
		},
	})
}
