module github.com/rocket-pool/smartnode

go 1.13

require (
	github.com/a8m/envsubst v1.3.0
	github.com/alessio/shellescape v1.4.1
	github.com/blang/semver/v4 v4.0.0
	github.com/btcsuite/btcd v0.23.1
	github.com/btcsuite/btcd/btcec/v2 v2.2.1 // indirect
	github.com/btcsuite/btcd/btcutil v1.1.2
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.18+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/ethereum/go-ethereum v1.10.25
	github.com/fatih/color v1.13.0
	github.com/ferranbt/fastssz v0.1.2
	github.com/gdamore/tcell/v2 v2.5.3
	github.com/glendc/go-external-ip v0.1.0
	github.com/google/uuid v1.3.0
	github.com/hashicorp/go-version v1.6.0
	github.com/herumi/bls-eth-go-binary v1.28.1 // indirect
	github.com/imdario/mergo v0.3.13
	github.com/klauspost/compress v1.15.11
	github.com/klauspost/cpuid/v2 v2.1.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/moby/term v0.0.0-20220808134915-39b0c02b01ae // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58
	github.com/prometheus/client_golang v1.13.0
	github.com/prysmaticlabs/go-bitfield v0.0.0-20210809151128-385d8c5e3fb7
	github.com/prysmaticlabs/prysm/v3 v3.1.1
	github.com/rivo/tview v0.0.0-20220916081518-2e69b7385a37
	github.com/rocket-pool/rocketpool-go v1.4.1
	github.com/sethvargo/go-password v0.2.0
	github.com/shirou/gopsutil/v3 v3.22.9
	github.com/tyler-smith/go-bip39 v1.1.0
	github.com/urfave/cli v1.22.10
	github.com/wealdtech/go-ens/v3 v3.5.5
	github.com/wealdtech/go-eth2-types/v2 v2.7.0
	github.com/wealdtech/go-eth2-util v1.7.0
	github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4 v1.3.0
	github.com/wealdtech/go-merkletree v1.0.1-0.20190605192610-2bb163c2ea2a
	github.com/web3-storage/go-w3s-client v0.0.6
	golang.org/x/crypto v0.0.0-20221005025214-4161e89ecf1b
	golang.org/x/sync v0.0.0-20220929204114-8fcdb60fdcc0
	golang.org/x/sys v0.0.0-20220928140112-f11e5e49a4ec // indirect
	golang.org/x/term v0.0.0-20220919170432-7a66f970e087
	google.golang.org/grpc v1.49.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools/v3 v3.3.0 // indirect
)

replace github.com/wealdtech/go-merkletree v1.0.1-0.20190605192610-2bb163c2ea2a => github.com/rocket-pool/go-merkletree v1.0.1-0.20220406020931-c262d9b976dd

replace github.com/web3-storage/go-w3s-client => github.com/rocket-pool/go-w3s-client v0.0.0-20221006052217-dbd9938d11d8

// replace github.com/rocket-pool/rocketpool-go => ../rocketpool-go
