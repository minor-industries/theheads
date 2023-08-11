module github.com/cacktopus/theheads

go 1.19

replace (
	github.com/minor-industries/codelab => ../minor-industries/codelab
	github.com/minor-industries/grm => ../minor-industries/grm
	github.com/minor-industries/packager => ../minor-industries/packager
	github.com/minor-industries/platform => ../minor-industries/platform
	github.com/minor-industries/protobuf => ../minor-industries/protobuf
)

require (
	github.com/ajstarks/svgo v0.0.0-20211024235047-1546f124cd8b
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	github.com/gin-contrib/pprof v1.4.0
	github.com/gin-gonic/gin v1.9.1
	github.com/goburrow/modbus v0.1.0
	github.com/golang/protobuf v1.5.3
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/gvalkov/golang-evdev v0.0.0-20220815104727-7e27d6ce89b6
	github.com/hashicorp/serf v0.10.1
	github.com/jessevdk/go-flags v1.5.0
	github.com/larspensjo/Go-simplex-noise v0.0.0-20121005164837-bfdcb9fc4b93
	github.com/minor-industries/codelab v0.0.0-00010101000000-000000000000
	github.com/minor-industries/grm v0.0.1
	github.com/minor-industries/packager v0.0.0-00010101000000-000000000000
	github.com/minor-industries/platform v0.0.2
	github.com/minor-industries/protobuf v0.0.0-00010101000000-000000000000
	github.com/mitchellh/mapstructure v1.5.0
	github.com/montanaflynn/stats v0.7.1
	github.com/orcaman/concurrent-map/v2 v2.0.1
	github.com/ory/dockertest/v3 v3.9.1
	github.com/pelletier/go-toml/v2 v2.0.8
	github.com/pin/tftp v2.1.0+incompatible
	github.com/pixiv/go-libjpeg v0.0.0-20190822045933-3da21a74767d
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.16.0
	github.com/rpi-ws281x/rpi-ws281x-go v1.0.8
	github.com/rqlite/gorqlite v0.0.0-20230708021416-2acd02b70b79
	github.com/ryanuber/go-glob v1.0.0
	github.com/soheilhy/cmux v0.1.5
	github.com/sony/gobreaker v0.5.0
	github.com/stretchr/testify v1.8.4
	github.com/vrischmann/envconfig v1.3.0
	go.uber.org/atomic v1.11.0
	go.uber.org/zap v1.24.0
	gobot.io/x/gobot v1.16.0
	gocv.io/x/gocv v0.31.0
	golang.org/x/crypto v0.9.0
	gonum.org/v1/gonum v0.13.0
	gonum.org/v1/plot v0.13.0
	google.golang.org/grpc v1.57.0
	google.golang.org/protobuf v1.31.0
	periph.io/x/conn/v3 v3.7.0
	periph.io/x/devices/v3 v3.7.1
	periph.io/x/host/v3 v3.8.2
)

require (
	aead.dev/minisign v0.2.0 // indirect
	git.sr.ht/~sbinet/gg v0.4.1 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/armon/go-metrics v0.0.0-20180917152333-f0300d1749da // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bitfield/script v0.22.0 // indirect
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/cli v20.10.14+incompatible // indirect
	github.com/docker/docker v20.10.7+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-fonts/liberation v0.3.1 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-latex/latex v0.0.0-20230307184459-12ec69307ad9 // indirect
	github.com/go-pdf/fpdf v0.8.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.14.0 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/goburrow/serial v0.1.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.3 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/go-sockaddr v1.0.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/memberlist v0.5.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/itchyny/gojq v0.12.12 // indirect
	github.com/itchyny/timefmt-go v0.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/miekg/dns v1.1.41 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.1.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/sean-/seed v0.0.0-20170313163322-e2103e2c3529 // indirect
	github.com/sigurn/crc8 v0.0.0-20160107002456-e55481d6f45c // indirect
	github.com/sigurn/utils v0.0.0-20190728110027-e1fefb11a144 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/image v0.7.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.9.0 // indirect
	golang.org/x/term v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mvdan.cc/sh/v3 v3.6.0 // indirect
	periph.io/x/periph v3.6.2+incompatible // indirect
)
