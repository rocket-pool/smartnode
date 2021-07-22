package trustednode

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	trustednodedao "github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/node"
	trustednodesettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
	minipoolutils "github.com/rocket-pool/rocketpool-go/tests/testutils/minipool"
	nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
)


func TestMemberDetails(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Disable min commission rate for unbonded pools
    if _, err := trustednodesettings.BootstrapMinipoolUnbondedMinFee(rp, uint64(0), ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get & check minimum member count
    if minMemberCount, err := trustednodedao.GetMinimumMemberCount(rp, nil); err != nil {
        t.Error(err)
    } else if minMemberCount == 0 {
        t.Error("Incorrect trusted node DAO minimum member count")
    }

    // Get & check initial member details
    if members, err := trustednodedao.GetMembers(rp, nil); err != nil {
        t.Error(err)
    } else if len(members) != 0 {
        t.Error("Incorrect initial trusted node DAO member count")
    }

    // Set proposal cooldown
    if _, err := trustednodesettings.BootstrapProposalCooldownTime(rp, 0, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Register nodes
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", trustedNodeAccount1.GetTransactor()); err != nil { t.Fatal(err) }
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", trustedNodeAccount2.GetTransactor()); err != nil { t.Fatal(err) }
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", trustedNodeAccount3.GetTransactor()); err != nil { t.Fatal(err) }
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Bootstrap trusted node DAO member
    memberId := "coolguy"
    memberEmail := "coolguy@rocketpool.net"
    if _, err := trustednodedao.BootstrapMember(rp, memberId, memberEmail, trustedNodeAccount1.Address, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if _, err := trustednodedao.BootstrapMember(rp, memberId, memberEmail, trustedNodeAccount2.Address, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if _, err := trustednodedao.BootstrapMember(rp, memberId, memberEmail, trustedNodeAccount3.Address, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get RPL bond amount
    rplBondAmount, err := trustednodesettings.GetRPLBond(rp, nil)
    if err != nil { t.Fatal(err) }

    // Mint trusted node RPL bond & join trusted node DAO
    if err := nodeutils.MintTrustedNodeBond(rp, ownerAccount, trustedNodeAccount1); err != nil { t.Fatal(err) }
    if err := nodeutils.MintTrustedNodeBond(rp, ownerAccount, trustedNodeAccount2); err != nil { t.Fatal(err) }
    if err := nodeutils.MintTrustedNodeBond(rp, ownerAccount, trustedNodeAccount3); err != nil { t.Fatal(err) }
    if _, err := trustednodedao.Join(rp, trustedNodeAccount1.GetTransactor()); err != nil {
        t.Fatal(err)
    }
    if _, err := trustednodedao.Join(rp, trustedNodeAccount2.GetTransactor()); err != nil {
        t.Fatal(err)
    }
    if _, err := trustednodedao.Join(rp, trustedNodeAccount3.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Submit a proposal
    if _, _, err := trustednodedao.ProposeMemberLeave(rp, "bye", trustedNodeAccount1.Address, trustedNodeAccount1.GetTransactor()); err != nil { t.Fatal(err) }

    // Create an unbonded minipool
    if _, err := minipoolutils.CreateMinipool(rp, ownerAccount, trustedNodeAccount1, big.NewInt(0)); err != nil { t.Fatal(err) }

    // Get & check updated member details
    if members, err := trustednodedao.GetMembers(rp, nil); err != nil {
        t.Error(err)
    } else if len(members) != 3 {
        t.Error("Incorrect updated trusted node DAO member count")
    } else {
        member := members[0]
        if !bytes.Equal(member.Address.Bytes(), trustedNodeAccount1.Address.Bytes()) {
            t.Errorf("Incorrect member address %s", member.Address.Hex())
        }
        if !member.Exists {
            t.Error("Incorrect member exists status")
        }
        if member.ID != memberId {
            t.Errorf("Incorrect member ID %s", member.ID)
        }
        if member.Url != memberEmail {
            t.Errorf("Incorrect member email %s", member.Url)
        }
        if member.JoinedTime == 0 {
            t.Errorf("Incorrect member joined time %d", member.JoinedTime)
        }
        if member.LastProposalTime == 0 {
            t.Errorf("Incorrect member last proposal time %d", member.LastProposalTime)
        }
        if member.RPLBondAmount.Cmp(rplBondAmount) != 0 {
            t.Errorf("Incorrect member RPL bond amount %s", member.RPLBondAmount.String())
        }
        if member.UnbondedValidatorCount != 1 {
            t.Errorf("Incorrect member unbonded validator count %d", member.UnbondedValidatorCount)
        }
    }

}


func TestUpgradeContract(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Upgrade contract
    contractName := "rocketDepositPool"
    contractNewAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
    contractNewAbi := "[{\"name\":\"foo\",\"type\":\"function\",\"inputs\":[],\"outputs\":[]}]"
    if _, err := trustednodedao.BootstrapUpgrade(rp, "upgradeContract", contractName, contractNewAbi, contractNewAddress, ownerAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated contract details
    if contractAddress, err := rp.GetAddress(contractName); err != nil {
        t.Error(err)
    } else if !bytes.Equal(contractAddress.Bytes(), contractNewAddress.Bytes()) {
        t.Errorf("Incorrect updated contract address %s", contractAddress.Hex())
    }
    if contractAbi, err := rp.GetABI(contractName); err != nil {
        t.Error(err)
    } else if _, ok := contractAbi.Methods["foo"]; !ok {
        t.Errorf("Incorrect updated contract ABI")
    }

}

