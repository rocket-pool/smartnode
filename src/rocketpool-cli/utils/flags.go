package utils

import (
	"github.com/urfave/cli/v2"
)

const (
	PrintTxDataFlag string = "print-tx-data"
	SignTxOnlyFlag  string = "sign-tx-only"
	NoRestartFlag   string = "no-restart"
	MnemonicFlag    string = "mnemonic"
)

var (
	YesFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "yes",
		Aliases: []string{"y"},
		Usage:   "Automatically confirm all interactive questions",
	}
	RawFlag *cli.BoolFlag = &cli.BoolFlag{
		Name: "raw",
	}
	ComposeFileFlag *cli.StringSliceFlag = &cli.StringSliceFlag{
		Name:    "compose-file",
		Aliases: []string{"f"},
		Usage:   "Supplemental Docker compose files for custom containers to include when performing service commands such as 'start' and 'stop'; this flag may be defined multiple times",
	}
)

func InstantiateFlag[FlagType cli.Flag](prototype FlagType, description string) cli.Flag {
	switch typedProto := any(prototype).(type) {
	case *cli.BoolFlag:
		return &cli.BoolFlag{
			Name:    typedProto.Name,
			Aliases: typedProto.Aliases,
			Usage:   description,
		}
	case *cli.Uint64Flag:
		return &cli.Uint64Flag{
			Name:    typedProto.Name,
			Aliases: typedProto.Aliases,
			Usage:   description,
		}
	case *cli.StringFlag:
		return &cli.StringFlag{
			Name:    typedProto.Name,
			Aliases: typedProto.Aliases,
			Usage:   description,
		}
	case *cli.Float64Flag:
		return &cli.Float64Flag{
			Name:    typedProto.Name,
			Aliases: typedProto.Aliases,
			Usage:   description,
		}
	default:
		panic("unsupported flag type")
	}
}
