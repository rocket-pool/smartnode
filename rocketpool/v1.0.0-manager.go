package rocketpool

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-version"
)

// A wrapper that holds the updated contract information for this version
type LegacyVersionWrapper_v1_0_0 struct {
	rp              *RocketPool
	rpVersion       *version.Version
	contractNameMap map[string]string
	abiMap          map[string]string
}

// Creates a new wrapper for this version
func newLegacyVersionWrapper_v1_0_0(rp *RocketPool) *LegacyVersionWrapper_v1_0_0 {
	rpVersion, _ := version.NewSemver("1.0.0")
	return &LegacyVersionWrapper_v1_0_0{
		rp:        rp,
		rpVersion: rpVersion,
		contractNameMap: map[string]string{
			"rocketRewardsPool":      "rocketRewardsPool.v1",
			"rocketClaimNode":        "rocketClaimNode.v1",
			"rocketClaimTrustedNode": "rocketClaimTrustedNode.v1",
			"rocketMinipoolManager":  "rocketMinipoolManager.v1"},
		abiMap: map[string]string{
			"rocketRewardsPool":      "eJzVWN9vmzAQ/lcmnvPADxtD37po0ia1VZVmT1U1HfYRoRKobJMmqvq/zxCSQBOatMsYe0vs43zfd2ffZ9+/WEn2VGhlXdyXPzXKDNLp6gmtC4vnmZbA9ZdJzh9R3+lcwgx/lEYxcLRGVgbz0vCXbBpcCiFRKTOt136gHnh9GFlKg8brQkOUpIlemdksz55gBVGKuy/MykrLghuHZlAlswx0Id/OvI5eLDCfr+Z5YQDEkCoctfEIXKKwLswX1UwLHmzjrGHwFJJ5ks3GNe4DCEafdtrNStPnFkPLaWH+u9TfOQWDOGvEtzH4jC+dzPGAp4etweT2apo/YqbGJRTjeWuMC6yiaCbIXlKMHYfboU89QpkTeg4NYrQjwimNInBsLlzhskAg4bHHfRFyZCSmFG2fUofXKOos7uJYoFRJnpn18kJ31WwZf7AD1wYWdJTgIsHnnWVcZFyvF6qqDQzGOtlvgJJYEEHtroBnqA13XyGFrNotR8JuZeVwRs4WuuuEAfpuJ9cm9Crb1WZfmDBN4u80SD0sGBF4hEck+jCMcT43xlUtDwmOZ3uIGJwKR92CUkMDYRMvYEy8tyv2cjIsBNQGN2ZIjyEoI78Cpa9BDAwBD23PYduG0BWW6aSmMzX6+Kbj30CzJ9Rmr/sMNJrlt2WitHqfhijP00McVONnJYARQhwRiD4JyEoNc2Qz9seAj8INKCUfZqBTBJ3EwjTXkO6EwoD2BBo54vgOPULIXlhrRjanVdU9pocl0+gc7O472ZOUay8o3xXa3Sn6qVB+B3VSkvqrV4f5IYt85/z12g+jE5yZAxAliuE1NJdEkTBa/H/ldgxZNTyUWvVC4hIW9322llSsz9dCyvXNa0BFZkgJGMW/UGSnkXKDy4ExAuZm66Af9czILUpeytJhkWGj6zJOnVOvBRN8BilUldqBIRGUUebCP0jrsIiIffBiH+OeibhM0/x5eE8qNoS+Z3N3oE1230mH0jVOqtfFut663wU3qbncPEYOKBk+w4Cw4FiHPguPbd3xC7f3wLf6o7aQtVJcC3DZIu7kF/IGE23kJLLRE+6xa85ZkO870XmfFcj3ROGf8xeGPo/AL5+KfwPvAY9J",
			"rocketClaimNode":        "eJzNlMFuwjAMhl9l6rmHUtoOuKFphx0mIbYbQshJ3CpaSVDiwtDEuy+ljNKNwg7dxK2Nnfj77V+efXhSrQqy3mhWfhIaBfnrdoXeyONakQFOd1PN35BeSBvI8KlMSoGj53sKlmXiwpwmjIUwaK0LU/UOHA52c9+zBITPBQGTuaStiyqtVrAFlmN9w1W2ZAruHnSHVmYKqDDfIzv/hH5+pFmjsVIrd1EX1KatcP+DWkFdugqcR11L3NSZaaE4VYX2VKDIGzk2bBIH73GUikjEQRtwhvSoygaIy8xM6/wc8v68U+LeQGAAqWgS/wCC46S/jKC0uDz/WvJDDnI50dbKavQ3ITyMwoClSfLnwqe4ASPsBA2/btQwTtqsWoa6NWu/z1nEgv/qwHipC8dyUz3oB3EUJ9C1/f0rxl6gaq7Bo78PCQYzad0DjW79eqmeyG/KBRb0WJQM2/YTL8fVcc0I73tiGLoWzz8BB105fQ==",
			"rocketClaimTrustedNode": "eJzNlMFuwjAMhl9l6rmHUtoOuKFphx02IbYbQshN3CpaSVDiwNDEuy+ljNKNwg4McWtsx/7s/PXk0xNyYcl4g0n5SaglFG/rBXoDjylJGhjdjRV7R3olpSHHpzIoA4ae70mYl4EzfRgw5FyjMc5NVR7YGTZT3zMEhM+WIBWFoLXzSiUXsIa0wPqGq2xIW+YSOqMRuQSy+qdn4x/QT/c0S9RGKOkuKkttvVl37tUd1KUrx3HUpcBVHZlZyagqtKUCSd7AsWGTOPiIo4xHPA7agHOkR1kOgJ9mTpUqjiFv7Rcl7vQ4BpDxJvEvINi/9LcQXDJDyF8UPy2DuvOHAsR8pIwRlQJuov8wCoM0S5Jr9T/GFWhuRqjZedmGcdIm3NJ1Wel2uyyN0uDKgxjOlXVINzWKbhBHcQL/9E/4Z9Q+Q9lckXvR7wI05sKV0Y2h/XnhHkyh2TWkQSeNkn7b7mLlq124ZoT3Hd4P3aSnX+6rRSk=",
			"rocketMinipoolManager":  "eJztWU1v2zgQ/SuFzj5QEimSuWU/CvSQRbBdbA9BEAzJYSrElgyJcmIU/e+lbCuyHduSY7tQgd5scTic9zjDGQ7vvgVpNq1cGVzd1T8dFhmM/5tPMbgKdJ65ArT78G+un9B9dnkBj/ipFrKgMRgFGUxqwYdiXeDamALL0g+7pR5Yffh+PwpKBw5vKgcqHadu7kezPJvCHNQY2xl+5dIVlfYKg++jbwF4ofkkr7yZFsYljjatNviCJrjyMxYjGyDg1ZqVsZM0S6d5Pt5h3+hoZVlusEPRq8Ebmir/P2JJq8mlkzVNzXBN2UrgZmX3nwV6Ck0rizPMnP9bpo8ZuKqov5EXIhRVMlRgwUoZhZRpZgxHIoWx2qg4soJLHlIdMqoTFuqIMwGUag4ISH/zvsX7X+hdMp93Mx8TybUilsWJNkKEIVMipDJkhpOYmMiiZELSOGZcm0iFIDE0CgQkPEadyBWMFc+tITMsyjTP/Hp55fbFbA1AtOg2kYk9IThL8bmVtFWm3XKhRRyCx7jalk2gjFpDDSP7DH5E9+q0ebVgqsPyjZ3ZvStns977uCHKb+d+6z87eEqzxwGDSLjSEMXsAIiPqTcvLdEMGIYJERKNhzzp2mua4YAxaJSKS9jyp26zcmtLdDuMG3XPHaeTdNfU/TF4i4X3aleVR5Lndbh06UcN9++wd1rgGDyhX0/QUS6D8gQNz6n7agp4rmuOE9SYtCzz8Ww/Ief0rlgx1GhhcN5122xp42YdnrXK1Xf3uwKzHTxvojIWrUzYseQ9LIqJvvF13XEgvSli+lbI70SN1vI4lLYD9RuzHuri6mAR32L/x4sO+EwO/aGMOo4vTcHgk1MoKEVNwksT8SsUG0oIENLyS3Mx/OqRcUXRJurMTPRILf1O1vXTZWinq1AxISSSl/ai/308GXDDdiQrSZRwTYfsSG+JHJpLSQOEE9KVsNXc4TqH00o94bxdbzm+u0b5Y37bSA8IuLYJUJ4cn6abnlLPeGpY+PslLV1HnaqWrao3BCy+n/cOTxKlkgh/Fvr1RtYgCADOuBS8625zNgL6hMBWkO2KrvPd7VCxmOHxqeSd+L8099+6l2wwqy/2XdHwHjqmNcDj6Yi04omOL1efpZkucOJx96rSej9arAHcOtcl4ZEQ+mKADP5cQJwpJjW/3M1yY4d63ClOh6QIRSEScfEKBrNq8qE9iqd5ueixNOrM8stCeLt336ccKmF8uGWjF89HN+1jzIGo33oAbCbteAHsWyActScsFMhJrPd1hc0yke3EcgYnV4oSgLNXY+XBRHS63ZEPzIQaeqCX3laDDvXyJW0gZ78VTCWS+krw/gfIvWxo",
		},
	}
}

// Get the version for this manager
func (m *LegacyVersionWrapper_v1_0_0) GetVersion() *version.Version {
	return m.rpVersion
}

// Get the versioned name of the contract if it was upgraded as part of this deployment
func (m *LegacyVersionWrapper_v1_0_0) GetVersionedContractName(contractName string) (string, bool) {
	legacyName, exists := m.contractNameMap[contractName]
	return legacyName, exists
}

// Get the ABI for the provided contract
func (m *LegacyVersionWrapper_v1_0_0) GetEncodedABI(contractName string) string {
	return m.abiMap[contractName]
}

// Get the contract with the provided name for this version of Rocket Pool
func (m *LegacyVersionWrapper_v1_0_0) GetContract(contractName string, opts *bind.CallOpts) (*Contract, error) {
	return getLegacyContract(m.rp, contractName, m, opts)
}

func (m *LegacyVersionWrapper_v1_0_0) GetContractWithAddress(contractName string, address common.Address) (*Contract, error) {
	return getLegacyContractWithAddress(m.rp, contractName, address, m)
}
