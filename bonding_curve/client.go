package bonding_curve

import (
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type Client struct {
	RPC                     *rpc.Client
	EventAuthority          solana.PublicKey
	Global                  solana.PublicKey
	GlobalVolumeAccumulator solana.PublicKey
	FeeConfig               solana.PublicKey
	FeesEventAuthority      solana.PublicKey
	FeesFeeProgramGlobal    solana.PublicKey
	AmmEventAuthority       solana.PublicKey
	AmmGlobalConfig         solana.PublicKey
	Commitment              rpc.CommitmentType
}

func NewClient(rpcClient *rpc.Client, commitment rpc.CommitmentType) *Client {
	return &Client{
		RPC:                     rpcClient,
		EventAuthority:          DeriveEventAuthority(),
		Global:                  DeriveGlobal(),
		GlobalVolumeAccumulator: DeriveGlobalVolumeAccumulator(),
		FeeConfig:               DeriveFeesFeeConfig(),
		FeesEventAuthority:      DeriveFeesEventAuthority(),
		FeesFeeProgramGlobal:    DeriveFeesFeeProgramGlobal(),
		AmmEventAuthority:       DeriveAmmEventAuthority(),
		AmmGlobalConfig:         DeriveAmmGlobalConfig(),
		Commitment:              commitment,
	}
}
