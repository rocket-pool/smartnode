package state

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	v110rc1_rewards "github.com/rocket-pool/rocketpool-go/legacy/v1.1.0-rc1/rewards"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// TODO: temp until rocketpool-go supports RocketStorage contract address lookups per block
func GetClaimIntervalTime(cfg *config.RocketPoolConfig, index uint64, rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	switch cfg.Smartnode.Network.Value.(cfgtypes.Network) {
	case cfgtypes.Network_Prater:
		if index < 2 {
			contractAddress := cfg.Smartnode.GetPreviousRewardsPoolAddresses()[0]
			return v110rc1_rewards.GetClaimIntervalTime(rp, opts, &contractAddress)
		}
	}

	return rewards.GetClaimIntervalTime(rp, opts)
}

// TODO: temp until rocketpool-go supports RocketStorage contract address lookups per block
func GetNodeOperatorRewardsPercent(cfg *config.RocketPoolConfig, index uint64, rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	switch cfg.Smartnode.Network.Value.(cfgtypes.Network) {
	case cfgtypes.Network_Prater:
		if index < 2 {
			contractAddress := cfg.Smartnode.GetPreviousRewardsPoolAddresses()[0]
			return v110rc1_rewards.GetNodeOperatorRewardsPercent(rp, opts, &contractAddress)
		}
	}

	return rewards.GetNodeOperatorRewardsPercent(rp, opts)
}

// TODO: temp until rocketpool-go supports RocketStorage contract address lookups per block
func GetTrustedNodeOperatorRewardsPercent(cfg *config.RocketPoolConfig, index uint64, rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	switch cfg.Smartnode.Network.Value.(cfgtypes.Network) {
	case cfgtypes.Network_Prater:
		if index < 2 {
			contractAddress := cfg.Smartnode.GetPreviousRewardsPoolAddresses()[0]
			return v110rc1_rewards.GetTrustedNodeOperatorRewardsPercent(rp, opts, &contractAddress)
		}
	}

	return rewards.GetTrustedNodeOperatorRewardsPercent(rp, opts)
}

// TODO: temp until rocketpool-go supports RocketStorage contract address lookups per block
func GetProtocolDaoRewardsPercent(cfg *config.RocketPoolConfig, index uint64, rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	switch cfg.Smartnode.Network.Value.(cfgtypes.Network) {
	case cfgtypes.Network_Prater:
		if index < 2 {
			contractAddress := cfg.Smartnode.GetPreviousRewardsPoolAddresses()[0]
			return v110rc1_rewards.GetProtocolDaoRewardsPercent(rp, opts, &contractAddress)
		}
	}

	return rewards.GetProtocolDaoRewardsPercent(rp, opts)
}

// TODO: temp until rocketpool-go supports RocketStorage contract address lookups per block
func GetPendingRPLRewards(cfg *config.RocketPoolConfig, index uint64, rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	switch cfg.Smartnode.Network.Value.(cfgtypes.Network) {
	case cfgtypes.Network_Prater:
		if index < 2 {
			contractAddress := cfg.Smartnode.GetPreviousRewardsPoolAddresses()[0]
			return v110rc1_rewards.GetPendingRPLRewards(rp, opts, &contractAddress)
		}
	}

	return rewards.GetPendingRPLRewards(rp, opts)
}
