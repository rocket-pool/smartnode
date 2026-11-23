package megapool

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

type MegapoolV2 interface {
	Megapool
}

// Megapool contract from delegate version 2, which adds support for EL-triggered forced exits
type megapoolV2 struct {
	megapoolV1
	// TODO keeping this while waiting for megapoolv2 details
}

const (
	megapoolV2EncodedAbi string = "eJzsWl9z2jgQ/yo3fmbuAXp96BsN9CYzaS9NQu8hk2EWayGaCMmVVvyZTr77jY3BFjbYYMO5IY9g7e5v/2pX0uMvD6SSy6myxvs0BmGw5XEZWDLep8dfHpcMF8hSnwi1BPGwDND75Fkuqf3XR6/lSZiGf8BUWUleyyN3wWvrcF7Ep5jD6Wmz4CtOIFBKfFaS3SGzPrJkPc5QkvfaegsK9nBE19LXCOYtq/imfXiLEgQtu0EgeBUdSdsSsGYgOANS+prloDulnj/WkrvG8Im8CFV7+NOivQxVuTFKzC5C1748mVs77RKadtrnUnTB6TI8uuDE5eQCNL1R/stFuPSe4DIUHUhR0adNbI/ucA6amSsBfHoG3aRi2K1Rv5ki1HUy1EjPdfILtCLlK9ED1f0f/NrjhjQf2R3bi+PJ/RKHO8MyEQpBIJZxo+21PGUpZv/U8gwB4VdLMOKC0zKKBRnAEkYipcnYSp+4kiXQpbfw4c760Gk7AKOe+IuVzBTjKwku4e6D8K0AwluUjMtJ7ANHUrGhwxRJKHNjZT+DKCcqcdgEbfefSnzCZNrNINfoM47zIovXE6wbb7276YxuOiZr/WcQAuUE+wtOjptysZWvKykR4f53GtYMA40+ENbJvjbbstXo3Exs8ay7acHqROmIiffI0/CfIHV94ik9ruLilAgrtFmunY7LSxdZfEr0A4TFYkhO3aizWDigwsPIpmCJR9Ym+q6/CLiG0JYPqza/EQa7AUObtrNp2KIJrom+/IbzDazVfcpPyzVOUVJTjPctnN0Y02gKeiXYLMpAWn+qEVJorSaZ6Hu4n7Imoao0jpwM1R2OrWzWvjMwqK8g4ASiSZBWEdUwYA0porU1m2mVruVYbWnkq2mgJMpyQgWYuLMKC3i8A26JbpXiEu4DaChp0KpxicvSwUwYBspwKg1ipJRIqM3WGXH0tZAKt25FylFx+T0eaA4ku9UY4TyUEhdBuNsNnKcB5UjXA87BhLh9j1KObPsAexdV6NePH1xxn0GA9Lec//FDGfKV2GwWhNRPmbNOQ9r69MddSETrU/h7UhomeANLZenP7USN5STcyQYCm1hVupLd2tELLt/Ly3t5eS8vv1V54atfbonJwhstCVOzT7BO95hu9blhpSm3Ju3Xqz6NHEj/cnpmGuYgrjQylMRBFIwskfxOeye0TrsCuOIOezhSku2/29uTM0NrsL9K8QcehmRx+mw5YjjLeHHNIFp4CId7PpFAVmNZJk7gxVW8BwR3SlG+G+LFMnXe4bi34rFrxfQorjbD+TpCRwL7gfKfj6lZQ40+StpBnzKTIj5e1n3vcXojrS6Lr+Xfc+Rl7JM5Mhr64Z2Pzjk5OsRF4kwO+sIliGQL+40cZYSKGkhDMA2KFf2m6CzBePBtro7eLsf9Z23Y0vwDWGbuJXLZn8e3CbR1W3lajxyNLNzRlaVbrdS4kSjTjdDV+qq5h8AEl65di0M1P0ZrbD2qKrgeQ9I7/R4FtzqFTFtwtHIJuhlqE34+u6GrPGzL7FdKsB4KnKzu+cld9do6nKHEeSWGhz8ZW4sbBBMNLP/BWIU3nRkNx1pN61Ft5/5wCjP16Tl8t+Mjn1WxUT4cdziIwjl3IKhZp4HBGyA0NAgYUL5a2dzKFmoQYgT+SwwyVjoRw9wIc1K+4m6QCIneD2wSZ09NyQRkTjAeWlP244ku7LHgXtANgURO9H9NgPrjMUZPUppmqU0glgN2IlsVG2AonctwcpelMotLHh5hcONqUjHaCwwxzC0cCShTZOjj4GVJMuvDqYrPsECbsEJ12o9P+a3O1iPFaOGOUcepgsMx4g2f8v2DxFhpv5aHhk//DQDs3Fxe"
)

// The decoded ABI for v2 megapools
var megapoolV2Abi *abi.ABI

// Create new megapool contract
func NewMegaPoolV2(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (MegapoolV2, error) {

	var contract *rocketpool.Contract
	var err error
	if megapoolV2Abi == nil {
		// Get contract
		contract, err = createMegapoolContractFromEncodedAbi(rp, address, megapoolV2EncodedAbi)
	} else {
		contract, err = createMegapoolContractFromAbi(rp, address, megapoolV2Abi)
	}
	if err != nil {
		return nil, err
	} else if megapoolV2Abi == nil {
		megapoolV2Abi = contract.ABI
	}

	// Create and return
	return &megapoolV2{
		megapoolV1: megapoolV1{
			Address:    address,
			Version:    2,
			Contract:   contract,
			RocketPool: rp,
		},
	}, nil
}
