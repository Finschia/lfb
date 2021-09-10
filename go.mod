module github.com/line/lfb

go 1.15

require (
	github.com/gorilla/mux v1.8.0
	github.com/line/lfb-sdk v1.0.0-init.1.0.20210910004202-84db942b6643
	github.com/line/ostracon v0.34.9-0.20210906083237-658e85d9b160
	github.com/line/tm-db/v2 v2.0.0-init.1.0.20210824011847-fcfa67dd3c70
	github.com/prometheus/client_golang v1.11.0
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
)

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
