package amm

import (
	solana "github.com/gagliardetto/solana-go"
	ammgen "github.com/krazyTry/pump-go/gen/amm"
)

type Pool = ammgen.Pool

type Fees = ammgen.Fees

type FeeTier = ammgen.FeeTier

type FeeConfig = ammgen.FeeConfig

type GlobalConfig = ammgen.GlobalConfig

type GlobalVolumeAccumulator = ammgen.GlobalVolumeAccumulator

type UserVolumeAccumulator = ammgen.UserVolumeAccumulator

type AccountWithPool struct {
	PublicKey solana.PublicKey
	Account   *ammgen.Pool
}

type DepositResult struct {
	Token1    uint64
	LpToken   uint64
	MaxToken0 uint64
	MaxToken1 uint64
}
type DepositLpTokenResult struct {
	MaxBase  uint64
	MaxQuote uint64
}
type WithdrawResult struct {
	Base     uint64
	Quote    uint64
	MinBase  uint64
	MinQuote uint64
}
type BuyBaseInputResult struct {
	InternalQuoteAmount uint64
	UIQuote             uint64
	MaxQuote            uint64
}
type BuyQuoteInputResult struct {
	Base                     uint64
	InternalQuoteWithoutFees uint64
	MaxQuote                 uint64
}
type SellBaseInputResult struct {
	UIQuote                uint64
	MinQuote               uint64
	InternalQuoteAmountOut uint64
}
type SellQuoteInputResult struct {
	InternalRawQuote uint64
	Base             uint64
	MinQuote         uint64
}
