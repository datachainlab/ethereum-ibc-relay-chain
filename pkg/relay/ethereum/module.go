package ethereum

import (
	"github.com/hyperledger-labs/yui-relayer/log"
)

const ModuleName = "ethereum.chain"

func GetModuleLogger() *log.RelayLogger {
	return log.GetLogger().
		WithModule(ModuleName)
}
