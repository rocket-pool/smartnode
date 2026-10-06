package config

import (
	"reflect"
	"strings"
	"testing"

	"github.com/rivo/tview"
	serviceconfig "github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func TestWizardRefreshesNetworkClientChoices(t *testing.T) {
	cfg, err := serviceconfig.NewRocketPoolConfig(t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	md := &MainDisplay{Config: cfg, app: tview.NewApplication(), pages: tview.NewPages(), navHeader: tview.NewTextView(), isNew: true}
	wiz := newWizard(md)
	for _, network := range []string{"plataberget", "testnet", "plataberget"} {
		// Use the network wizard action to exercise a change after modals exist.
		for i, option := range cfg.Smartnode.Network.Options {
			if string(option.Value.(cfgtypes.Network)) == network {
				wiz.networkModal.modal.done(i, option.Name)
				break
			}
		}
		for _, tc := range []struct {
			step        *choiceWizardStep
			plataberget []string
		}{
			{wiz.executionLocalModal, []string{"Random (Recommended)", "Nethermind"}},
			{wiz.consensusLocalModal, []string{"Random (Recommended)", "Lodestar", "Nimbus", "Teku"}},
			{wiz.consensusExternalSelectModal, []string{"Lodestar", "Nimbus", "Teku"}},
			{wiz.nativeCcModal, []string{"Lodestar", "Nimbus", "Teku"}},
		} {
			tc.step.show()
			var labels []string
			for _, form := range tc.step.modal.forms {
				for i := 0; i < form.GetButtonCount(); i++ {
					labels = append(labels, strings.TrimSpace(form.GetButton(i).GetLabel()))
				}
			}
			if network == "plataberget" && !reflect.DeepEqual(labels, tc.plataberget) {
				t.Fatalf("stale wizard choices: %v, want %v", labels, tc.plataberget)
			}
			if network == "testnet" && len(labels) < 5 {
				t.Fatalf("Hoodi choices not restored: %v", labels)
			}
		}
	}
}

func TestSettingsRefreshesNetworkClientChoices(t *testing.T) {
	cfg, err := serviceconfig.NewRocketPoolConfig(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	layout := newStandardLayout()
	layout.createForm(&cfg.Smartnode.Network, "Native client")
	items := createParameterizedFormItems([]*cfgtypes.Parameter{&cfg.Native.ConsensusClient}, layout)
	layout.mapParameterizedFormItems(items...)
	layout.form.AddFormItem(items[0].item)
	dropDown := items[0].item.(*DropDown)
	for _, network := range []cfgtypes.Network{"plataberget", "testnet", "plataberget"} {
		cfg.ChangeNetwork(network)
		layout.refresh()
		for i, option := range cfg.Native.ConsensusClient.Options {
			if dropDown.options[i].Text != option.Name {
				t.Fatal("settings dropdown has stale options")
			}
			dropDown.SetCurrentOption(i)
			if cfg.Native.ConsensusClient.Value != option.Value {
				t.Fatalf("selected %s but saved %v", option.Name, cfg.Native.ConsensusClient.Value)
			}
			if strings.TrimSpace(layout.descriptionBox.GetText(false)) != option.Description {
				t.Fatalf("%s/%s: description %q, want %q", network, option.Name, layout.descriptionBox.GetText(false), option.Description)
			}
		}
	}
}

func TestSettingsPageRebuildsAfterNetworkChange(t *testing.T) {
	cfg, err := serviceconfig.NewRocketPoolConfig(t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	md := NewMainDisplay(tview.NewApplication(), nil, cfg, false, false, false)
	cfg.ChangeNetwork("plataberget")
	md.settingsHome.refresh()
	ec := md.settingsHome.ecPage
	if len(ec.ecDropdown.item.(*DropDown).options) != 1 {
		t.Fatal("execution settings retained unsupported clients")
	}
	// The parameter section must switch from Geth to Nethermind as well.
	foundNethermind := false
	for i := 0; i < ec.layout.form.GetFormItemCount(); i++ {
		for _, item := range ec.nethermindItems {
			if ec.layout.form.GetFormItem(i) == item.item {
				foundNethermind = true
			}
		}
	}
	if !foundNethermind {
		t.Fatal("execution settings still show the previous client's parameters")
	}
}
