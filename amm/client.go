package amm

import (
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type Client struct {
	RPC                     *rpc.Client
	EventAuthority          solana.PublicKey
	GlobalConfig            solana.PublicKey
	GlobalVolumeAccumulator solana.PublicKey
	FeeConfig               solana.PublicKey
	Commitment              rpc.CommitmentType
}

func NewClient(rpcClient *rpc.Client, commitment rpc.CommitmentType) *Client {
	return &Client{
		RPC:                     rpcClient,
		EventAuthority:          DeriveEventAuthority(),
		GlobalConfig:            DeriveGlobalConfig(),
		GlobalVolumeAccumulator: DeriveGlobalVolumeAccumulator(),
		FeeConfig:               DeriveFeesFeeConfig(),
		Commitment:              commitment,
	}
}
