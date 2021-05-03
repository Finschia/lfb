module github.com/line/lfb

go 1.15

require (
	github.com/dgraph-io/ristretto v0.0.3 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/line/lfb-sdk v1.0.0-init.1.0.20210624050312-9bcea92166c3
	github.com/line/ostracon v0.34.9-0.20210610071151-a52812ac9add
	github.com/line/tm-db/v2 v2.0.0-init.1.0.20210413083915-5bb60e117524
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
)

replace (
	github.com/CosmWasm/wasmvm => github.com/line/wasmvm v0.14.0-0.4.0
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)