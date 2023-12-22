package cmd

import (
	"github.com/hyperledger-labs/yui-relayer/config"
	"github.com/spf13/cobra"
)

func EthereumCmd(ctx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ethereum",
		Short: "manage ethereum configurations",
	}

	cmd.AddCommand(
		pendingCmd(ctx),
	)

	return cmd
}
