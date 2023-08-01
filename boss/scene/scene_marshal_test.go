package scene

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"
	"testing"
)

//go:embed scene.toml
var sceneTOML []byte

func TestMarshal(t *testing.T) {
	sc := &Scene{}

	d := toml.NewDecoder(bytes.NewBuffer(sceneTOML))
	d.DisallowUnknownFields()

	err := d.Decode(sc)
	require.NoError(t, err)

	fmt.Println(sc)
}
