package cmd

import (
	"fmt"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum"
	"github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum/cmd/pending"
	"github.com/hyperledger-labs/yui-relayer/config"
	"github.com/spf13/cobra"
)

func pendingCmd(ctx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending",
		Short: "Manage ethereum pending transactions",
	}

	cmd.AddCommand(
		showPendingTxCmd(ctx),
		replacePendingTxCmd(ctx),
	)

	return cmd
}

func showPendingTxCmd(ctx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [chain-id]",
		Aliases: []string{"list"},
		Short:   "Show minimum nonce pending transactions sent by relayer",
		RunE: func(cmd *cobra.Command, args []string) error {
			chain, err := ctx.Config.GetChain(args[0])
			if err != nil {
				return err
			}
			ethChain := chain.Chain.(*ethereum.Chain)
			logic := pending.NewLogic(ethChain)
			tx, err := logic.ShowPendingTx(cmd.Context())
			if err != nil {
				return err
			}
			json, err := tx.MarshalJSON()
			if err != nil {
				return err
			}
			fmt.Println(string(json))
			return nil
		},
	}
	return cmd
}

func replacePendingTxCmd(ctx *config.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "replace [chain-id]",
		Aliases: []string{"replace"},
		Short:   "Replace minimum nonce pending transaction sent by relayer",
		RunE: func(cmd *cobra.Command, args []string) error {
			chain, err := ctx.Config.GetChain(args[0])
			if err != nil {
				return err
			}
			ethChain := chain.Chain.(*ethereum.Chain)
			logic := pending.NewLogic(ethChain)
			tx, err := logic.ShowPendingTx(cmd.Context())
			if err != nil {
				return err
			}
			ethereum.GetModuleLogger().Info("Pending transaction found", "txHash", tx.Hash())

			return logic.ReplacePendingTx(cmd.Context(), tx.Hash())
		},
	}
	return cmd
}
