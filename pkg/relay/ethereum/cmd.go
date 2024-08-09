package ethereum

import (
	"fmt"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	chantypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/contract/iibcchannelupgradablemodule"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperledger-labs/yui-relayer/config"
	"github.com/spf13/cobra"
)

func ethereumCmd(ctx *config.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:     "ethereum",
		Aliases: []string{"eth"},
	}

	cmd.AddCommand(
		channelUpgradeCmd(ctx),
	)

	return &cmd
}

func channelUpgradeCmd(ctx *config.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:     "channel-upgrade",
		Aliases: []string{"upgrade"},
	}

	cmd.AddCommand(
		proposeUpgradeCmd(ctx),
		allowTransitionToFlushCompleteCmd(ctx),
	)

	return &cmd
}

func proposeUpgradeCmd(ctx *config.Context) *cobra.Command {
	const (
		flagAppAddress       = "app-address"
		flagOrdering         = "ordering"
		flagConnectionHops   = "connection-hops"
		flagVersion          = "version"
		flagTimeoutHeight    = "timeout-height"
		flagTimeoutTimestamp = "timeout-timestamp"
	)

	cmd := cobra.Command{
		Use:     "propose-init",
		Aliases: []string{"propose"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pathName := args[0]
			chainID := args[1]

			// get app address from flags
			var appAddr common.Address
			if appAddrHex, err := cmd.Flags().GetString(flagAppAddress); err != nil {
				return err
			} else {
				appAddr = common.HexToAddress(appAddrHex)
			}

			// get ordering from flags
			var ordering uint8
			if s, err := cmd.Flags().GetString(flagOrdering); err != nil {
				return err
			} else if n := chantypes.Order_value[s]; n == 0 {
				return fmt.Errorf("invalid ordering flag: %s", s)
			} else {
				ordering = uint8(n)
			}

			// get connection hops from flags
			connHops, err := cmd.Flags().GetStringSlice(flagConnectionHops)
			if err != nil {
				return err
			}

			// get version from flags
			version, err := cmd.Flags().GetString(flagVersion)
			if err != nil {
				return err
			}

			// get timeout height
			var timeoutHeight clienttypes.Height
			if s, err := cmd.Flags().GetString(flagTimeoutHeight); err != nil {
				return err
			} else if timeoutHeight, err = clienttypes.ParseHeight(s); err != nil {
				return err
			}

			// get timeout timestamp
			timeoutTimestamp, err := cmd.Flags().GetUint64(flagTimeoutTimestamp)
			if err != nil {
				return err
			}

			var ethChain *Chain
			if chains, _, _, err := ctx.Config.ChainsFromPath(pathName); err != nil {
				return err
			} else if chain, ok := chains[chainID]; !ok {
				return fmt.Errorf("chain not found: %s", chainID)
			} else if ethChain, ok = chain.Chain.(*Chain); !ok {
				return fmt.Errorf("chain is not ethereum: %T", chain.Chain)
			}

			return ethChain.ProposeUpgrade(
				cmd.Context(),
				appAddr,
				ethChain.pathEnd.PortID,
				ethChain.pathEnd.ChannelID,
				iibcchannelupgradablemodule.UpgradeFieldsData{
					Ordering:       ordering,
					ConnectionHops: connHops,
					Version:        version,
				},
				iibcchannelupgradablemodule.TimeoutData{
					Height:    iibcchannelupgradablemodule.HeightData(timeoutHeight),
					Timestamp: timeoutTimestamp,
				},
			)
		},
	}

	cmd.Flags().String(flagAppAddress, "", "IBC app module address")
	cmd.Flags().String(flagOrdering, "", "channel ordering applied for the new channel")
	cmd.Flags().StringSlice(flagConnectionHops, []string{}, "connection hops applied for the new channel")
	cmd.Flags().String(flagVersion, "", "channel version applied for the new channel")
	cmd.Flags().String(flagTimeoutHeight, "", "timeout height")
	cmd.Flags().Uint64(flagTimeoutTimestamp, 0, "timeout timestamp")

	return &cmd
}

func allowTransitionToFlushCompleteCmd(ctx *config.Context) *cobra.Command {
	const (
		flagAppAddress      = "app-address"
		flagUpgradeSequence = "upgrade-sequence"
	)

	cmd := cobra.Command{
		Use:     "allow-transition-to-flush-complete",
		Aliases: []string{"allow"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pathName := args[0]
			chainID := args[1]

			// get app address from flags
			var appAddr common.Address
			if appAddrHex, err := cmd.Flags().GetString(flagAppAddress); err != nil {
				return err
			} else {
				appAddr = common.HexToAddress(appAddrHex)
			}

			// get upgrade sequence from flags
			upgradeSequence, err := cmd.Flags().GetUint64(flagUpgradeSequence)
			if err != nil {
				return err
			}

			var ethChain *Chain
			if chains, _, _, err := ctx.Config.ChainsFromPath(pathName); err != nil {
				return err
			} else if chain, ok := chains[chainID]; !ok {
				return fmt.Errorf("chain not found: %s", chainID)
			} else if ethChain, ok = chain.Chain.(*Chain); !ok {
				return fmt.Errorf("chain is not ethereum: %T", chain.Chain)
			}

			return ethChain.AllowTransitionToFlushComplete(
				cmd.Context(),
				appAddr,
				ethChain.pathEnd.PortID,
				ethChain.pathEnd.ChannelID,
				upgradeSequence,
			)
		},
	}

	cmd.Flags().String(flagAppAddress, "", "IBC app module address")
	cmd.Flags().Uint64(flagUpgradeSequence, 0, "upgrade sequence")

	return &cmd
}
