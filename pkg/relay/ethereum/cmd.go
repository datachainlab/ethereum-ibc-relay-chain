package ethereum

import (
	"errors"
	"fmt"
	"strings"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	chantypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/contract/iibcchannelupgradablemodule"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hyperledger-labs/yui-relayer/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
		proposeAppVersionCmd(ctx),
	)

	return &cmd
}

func proposeUpgradeCmd(ctx *config.Context) *cobra.Command {
	const (
		flagOrdering         = "ordering"
		flagConnectionHops   = "connection-hops"
		flagVersion          = "version"
		flagTimeoutHeight    = "timeout-height"
		flagTimeoutTimestamp = "timeout-timestamp"
	)

	cmd := cobra.Command{
		Use:     "propose-upgrade",
		Aliases: []string{"propose"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pathName := args[0]
			chainID := args[1]

			// get ordering from flags
			ordering, err := getOrderFromFlags(cmd.Flags(), flagOrdering)
			if err != nil {
				return err
			} else if ordering == chantypes.NONE {
				return errors.New("NONE is unacceptable channel ordering")
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
				ethChain.pathEnd.PortID,
				ethChain.pathEnd.ChannelID,
				iibcchannelupgradablemodule.UpgradeFieldsData{
					Ordering:       uint8(ordering),
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

	cmd.Flags().String(flagOrdering, "", "channel ordering applied for the new channel")
	cmd.Flags().StringSlice(flagConnectionHops, []string{}, "connection hops applied for the new channel")
	cmd.Flags().String(flagVersion, "", "channel version applied for the new channel")
	cmd.Flags().String(flagTimeoutHeight, "", "timeout height")
	cmd.Flags().Uint64(flagTimeoutTimestamp, 0, "timeout timestamp")

	return &cmd
}

func allowTransitionToFlushCompleteCmd(ctx *config.Context) *cobra.Command {
	const (
		flagUpgradeSequence = "upgrade-sequence"
	)

	cmd := cobra.Command{
		Use:     "allow-transition-to-flush-complete",
		Aliases: []string{"allow"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pathName := args[0]
			chainID := args[1]

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
				ethChain.pathEnd.PortID,
				ethChain.pathEnd.ChannelID,
				upgradeSequence,
			)
		},
	}

	cmd.Flags().Uint64(flagUpgradeSequence, 0, "upgrade sequence")

	return &cmd
}

func proposeAppVersionCmd(ctx *config.Context) *cobra.Command {
	const (
		flagVersion         = "version"
		flagImplementation  = "implementation"
		flagInitialCalldata = "initial-calldata"
	)

	cmd := cobra.Command{
		Use:  "propose-app-version",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pathName := args[0]
			chainID := args[1]

			// get app version from flags
			version, err := cmd.Flags().GetString(flagVersion)
			if err != nil {
				return err
			}

			// get the new implementation address from flags
			var implementation common.Address
			if bz, err := cmd.Flags().GetBytesHex(flagImplementation); err != nil {
				return err
			} else {
				implementation = common.BytesToAddress(bz)
			}

			// get the initial calldata from flags
			initialCalldata, err := cmd.Flags().GetBytesHex(flagInitialCalldata)
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

			return ethChain.ProposeAppVersion(
				cmd.Context(),
				ethChain.pathEnd.PortID,
				ethChain.pathEnd.ChannelID,
				version,
				implementation,
				initialCalldata,
			)
		},
	}

	cmd.Flags().String(flagVersion, "", "app version")
	cmd.Flags().BytesHex(flagImplementation, nil, "new implementation")
	cmd.Flags().BytesHex(flagInitialCalldata, nil, "initial calldata")

	return &cmd
}

func getOrderFromFlags(flags *pflag.FlagSet, flagName string) (chantypes.Order, error) {
	s, err := flags.GetString(flagName)
	if err != nil {
		return 0, err
	}

	s = "ORDER_" + strings.ToUpper(s)
	value, ok := chantypes.Order_value[s]
	if !ok {
		return 0, fmt.Errorf("invalid channel order specified: %s", s)
	}

	return chantypes.Order(value), nil
}
