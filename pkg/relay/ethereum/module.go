package ethereum

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/hyperledger-labs/yui-relayer/config"
	"github.com/hyperledger-labs/yui-relayer/log"
	"github.com/spf13/cobra"
)

type Module struct{}

var _ config.ModuleI = (*Module)(nil)

const ModuleName = "ethereum.chain"

// Name returns the name of the module
func (Module) Name() string {
	return ModuleName
}

// RegisterInterfaces register the module interfaces to protobuf Any.
func (Module) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	RegisterInterfaces(registry)
}

// GetCmd returns the command
func (Module) GetCmd(ctx *config.Context) *cobra.Command {
	return ethereumCmd(ctx)
}

func GetModuleLogger() *log.RelayLogger {
	return log.GetLogger().
		WithModule(ModuleName)
}
