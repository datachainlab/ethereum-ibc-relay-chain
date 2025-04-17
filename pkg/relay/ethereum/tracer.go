package ethereum

import "go.opentelemetry.io/otel"

var (
	tracer = otel.Tracer("github.com/datachainlab/ethereum-ibc-relay-chain/pkg/relay/ethereum")
)
