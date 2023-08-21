package main

import (
	"github.com/minor-industries/grm"
	"github.com/minor-industries/packager/pkg/packager"
	"os"
	"strings"
)

func Packager(
	rule string,
	buildCallback func(request *packager.BuildRequest) error,
) {

	parts := strings.Split(rule, "-")
	if len(parts) < 2 {
		panic("invalid rule")
	}
	arch := parts[len(parts)-1]
	pkgName := strings.Join(parts[:len(parts)-1], "-")

	if err := packager.Run(pkgName, &packager.Opts{
		Minor:        true,
		AllowDirty:   false,
		New:          false,
		Arch:         arch,
		SharedFolder: os.ExpandEnv("$HOME/shared"),
	}, buildCallback); err != nil {
		panic(err)
	}

}

func NewDocker(rule string) {
	Packager(rule, func(request *packager.BuildRequest) error {
		grm.DockerWithCustomVersion(request.Version)(rule)
		return nil
	})
}

var rules = map[string]func(rule string){
	"camera-arm64": NewDocker,
}

func main() {
	grm.Main(rules)
}
