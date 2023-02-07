module github.com/rocket-pool/smartnode

go 1.13

require (
	github.com/a8m/envsubst v1.3.0
	github.com/alessio/shellescape v1.4.1
	github.com/blang/semver/v4 v4.0.0
	github.com/btcsuite/btcd v0.23.4
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/btcsuite/btcd/btcutil v1.1.3
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v23.0.0+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/dustin/go-humanize v1.0.1
	github.com/ethereum/go-ethereum v1.10.26
	github.com/fatih/color v1.14.1
	github.com/ferranbt/fastssz v0.1.2
	github.com/gdamore/tcell/v2 v2.5.4
	github.com/glendc/go-external-ip v0.1.0
	github.com/google/uuid v1.3.0
	github.com/hashicorp/go-version v1.6.0
	github.com/imdario/mergo v0.3.13
	github.com/klauspost/compress v1.15.15
	github.com/klauspost/cpuid/v2 v2.2.3
	github.com/mitchellh/go-homedir v1.1.0
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58
	github.com/prometheus/client_golang v1.14.0
	github.com/prysmaticlabs/go-bitfield v0.0.0-20210809151128-385d8c5e3fb7
	github.com/prysmaticlabs/prysm/v3 v3.2.0
	github.com/rivo/tview v0.0.0-20230203122838-f0550c7918da
	github.com/rocket-pool/rocketpool-go v1.10.1-0.20230207001250-dd6b472c9971
	github.com/sethvargo/go-password v0.2.0
	github.com/shirou/gopsutil/v3 v3.23.1
	github.com/tyler-smith/go-bip39 v1.1.0
	github.com/urfave/cli v1.22.12
	github.com/wealdtech/go-ens/v3 v3.5.5
	github.com/wealdtech/go-eth2-types/v2 v2.8.1-0.20230131115251-b93cf60cee26
	github.com/wealdtech/go-eth2-util v1.8.0
	github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4 v1.3.0
	github.com/wealdtech/go-merkletree v1.0.1-0.20190605192610-2bb163c2ea2a
	github.com/web3-storage/go-w3s-client v0.0.7
	golang.org/x/crypto v0.5.0
	golang.org/x/sync v0.1.0
	golang.org/x/term v0.4.0
	google.golang.org/grpc v1.52.3 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools/v3 v3.4.0 // indirect
)

replace github.com/wealdtech/go-merkletree v1.0.1-0.20190605192610-2bb163c2ea2a => github.com/rocket-pool/go-merkletree v1.0.1-0.20220406020931-c262d9b976dd

replace github.com/web3-storage/go-w3s-client => github.com/rocket-pool/go-w3s-client v0.0.0-20221006052217-dbd9938d11d8

// replace github.com/wealdtech/go-eth2-types/v2 => github.com/rocket-pool/go-eth2-types/v2 v2.0.0-20230130220714-d88838162252

// replace github.com/rocket-pool/rocketpool-go => ../rocketpool-go
