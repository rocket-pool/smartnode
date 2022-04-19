module github.com/rocket-pool/smartnode

go 1.13

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/a8m/envsubst v1.2.0
	github.com/alessio/shellescape v1.4.1
	github.com/blang/semver/v4 v4.0.0
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.4.2-0.20180625184442-8e610b2b55bf
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/ethereum/go-ethereum v1.10.16
	github.com/fatih/color v1.13.0
	github.com/ferranbt/fastssz v0.0.0-20220103083642-bc5fefefa28b
	github.com/gdamore/tcell/v2 v2.4.1-0.20210905002822-f057f0a857a1
	github.com/glendc/go-external-ip v0.0.0-20200601212049-c872357d968e
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.2
	github.com/herumi/bls-eth-go-binary v0.0.0-20211108015406-b5186ba08dc7 // indirect
	github.com/imdario/mergo v0.3.12
	github.com/mitchellh/go-homedir v1.1.0
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58
	github.com/prometheus/client_golang v1.11.0
	github.com/prysmaticlabs/prysm/v2 v2.0.1
	github.com/rivo/tview v0.0.0-20220106183741-90d72bc664f5
	github.com/rocket-pool/rocketpool-go v1.10.1-0.20220419073641-28165de889f1
	github.com/sethvargo/go-password v0.2.0
	github.com/shirou/gopsutil/v3 v3.21.11
	github.com/tyler-smith/go-bip39 v1.1.0
	github.com/urfave/cli v1.22.5
	github.com/wealdtech/go-eth2-types/v2 v2.6.0
	github.com/wealdtech/go-eth2-util v1.7.0
	github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4 v1.2.0
	github.com/wealdtech/go-merkletree v1.0.1-0.20190605192610-2bb163c2ea2a
	golang.org/x/crypto v0.0.0-20211115234514-b4de73f9ece8
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211
	golang.org/x/tools v0.1.9 // indirect
	google.golang.org/grpc v1.42.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/wealdtech/go-merkletree v1.0.1-0.20190605192610-2bb163c2ea2a => github.com/rocket-pool/go-merkletree v1.0.1-0.20220406020931-c262d9b976dd

//replace github.com/rocket-pool/rocketpool-go => ../rocketpool-go
