module github.com/NpoolPlatform/cloud-hashing-apis

go 1.16

require (
	entgo.io/ent v0.9.1
	github.com/NpoolPlatform/application-management v0.0.0-20211211055746-45df92c06c5f
	github.com/NpoolPlatform/cloud-hashing-billing v0.0.0-20211120094336-58e1a1ffa8be
	github.com/NpoolPlatform/cloud-hashing-goods v0.0.0-20211210123412-481038d89121
	github.com/NpoolPlatform/cloud-hashing-inspire v0.0.0-20211202123501-eece2e1c91af
	github.com/NpoolPlatform/cloud-hashing-order v0.0.0-20211229094147-8745991fbea7
	github.com/NpoolPlatform/go-service-framework v0.0.0-20211222114515-4928e6cf3f1f
	github.com/NpoolPlatform/message v0.0.0-20211210024747-4c069e246981
	github.com/NpoolPlatform/sphinx-coininfo v0.0.0-20211208035009-5ad2768d2290
	github.com/NpoolPlatform/sphinx-proxy v0.0.0-20211210102925-d9b8abe11021
	github.com/NpoolPlatform/sphinx-service v0.0.0-20211124061425-6ef21f0b762f // indirect
	github.com/NpoolPlatform/user-management v0.0.0-20211206121520-304b4b6e1680
	github.com/go-resty/resty/v2 v2.7.0
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa
	google.golang.org/grpc v1.42.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.27.1
)

replace google.golang.org/grpc => github.com/grpc/grpc-go v1.41.0
