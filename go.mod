module github.com/rocket-pool/smartnode

go 1.13

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.4.2-0.20180625184442-8e610b2b55bf
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/ethereum/go-ethereum v1.10.0
	github.com/fatih/color v1.7.0
	github.com/gogo/protobuf v1.3.1
	github.com/google/uuid v1.1.5
	github.com/gorilla/websocket v1.4.2
	github.com/imdario/mergo v0.3.9
	github.com/minio/highwayhash v1.0.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/prysmaticlabs/ethereumapis v0.0.0-20200729044127-8027cc96e2c0
	github.com/prysmaticlabs/go-ssz v0.0.0-20210121151755-f6208871c388
	github.com/rocket-pool/rocketpool-go v0.0.8
	github.com/tyler-smith/go-bip39 v1.0.1-0.20181017060643-dbb3b84ba2ef
	github.com/urfave/cli v1.22.4
	github.com/wealdtech/go-eth2-types/v2 v2.5.0
	github.com/wealdtech/go-eth2-util v1.6.0
	github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4 v1.1.1
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	golang.org/x/sys v0.0.0-20210305230114-8fe3ee5dd75b // indirect
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1
	google.golang.org/grpc v1.29.1
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/rocket-pool/rocketpool-go v0.0.8 => ../rocketpool-go
