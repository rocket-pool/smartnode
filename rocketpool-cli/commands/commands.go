package commands

import (
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/auction"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/minipool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/network"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/node"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/odao"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/pdao"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/queue"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/security"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/service"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/wallet"
	"github.com/urfave/cli/v2"
)

func RegisterCommands(app *cli.App) {
	auction.RegisterCommands(app, "auction", []string{"a"})
	minipool.RegisterCommands(app, "minipool", []string{"m"})
	network.RegisterCommands(app, "network", []string{"e"})
	node.RegisterCommands(app, "node", []string{"n"})
	odao.RegisterCommands(app, "odao", []string{"o"})
	pdao.RegisterCommands(app, "pdao", []string{"p"})
	queue.RegisterCommands(app, "queue", []string{"q"})
	security.RegisterCommands(app, "security", []string{"c"})
	service.RegisterCommands(app, "service", []string{"s"})
	wallet.RegisterCommands(app, "wallet", []string{"w"})
}
