package bonding_curve

import (
	"fmt"
	"math/big"

	solana "github.com/solana-foundation/solana-go/v2"
)

// UnsupportedQuoteMintError mirrors create_v2 rejecting a mint that is not SOL,
// not whitelisted on Global and not listed in QuoteControl.
type UnsupportedQuoteMintError struct {
	QuoteMint solana.PublicKey
}

func (e *UnsupportedQuoteMintError) Error() string {
	return fmt.Sprintf("unsupported quote mint %s: not SOL, not whitelisted on Global and not listed in QuoteControl", e.QuoteMint)
}

// CurveDepthExceededError mirrors pump error 6105.
type CurveDepthExceededError struct {
	Depth         int
	MaxCurveDepth int
}

func (e *CurveDepthExceededError) Error() string {
	return fmt.Sprintf("curve depth %d exceeds the max curve depth %d", e.Depth, e.MaxCurveDepth)
}

// QuoteBondingCurveNotEligibleError mirrors pump error 6100.
type QuoteBondingCurveNotEligibleError struct {
	QuoteMint solana.PublicKey
}

func (e *QuoteBondingCurveNotEligibleError) Error() string {
	return fmt.Sprintf("pump coin %s is not eligible as a quote: mayhem mode, or a chain root whose own quote is not listed", e.QuoteMint)
}

// QuoteCurveAwaitingMigrationError mirrors pump error 6107.
type QuoteCurveAwaitingMigrationError struct {
	QuoteMint solana.PublicKey
}

func (e *QuoteCurveAwaitingMigrationError) Error() string {
	return fmt.Sprintf("pump coin %s has completed its curve but is not migrated yet", e.QuoteMint)
}

// QuoteReservesOutOfRangeError mirrors pump error 6104.
type QuoteReservesOutOfRangeError struct {
	QuoteMint solana.PublicKey
	Derived   *big.Int
}

func (e *QuoteReservesOutOfRangeError) Error() string {
	derived := "0"
	if e.Derived != nil {
		derived = e.Derived.String()
	}
	return fmt.Sprintf("reserves %s derived from pump coin %s are out of range", derived, e.QuoteMint)
}

// CashbackDeprecatedError mirrors pump error 6082.
type CashbackDeprecatedError struct {
	Mint solana.PublicKey
}

func (e *CashbackDeprecatedError) Error() string {
	return fmt.Sprintf("cashback coins can no longer be created (mint %s): is_cashback_enabled is deprecated", e.Mint)
}

// HolderRewardDisabledError mirrors pump error 6084.
type HolderRewardDisabledError struct{}

func (e *HolderRewardDisabledError) Error() string {
	return "holder-reward coins are disabled (Global.isHolderRewardEnabled is false)"
}

// CreatorFeeNotConfigurableError mirrors pump error 6077.
type CreatorFeeNotConfigurableError struct{}

func (e *CreatorFeeNotConfigurableError) Error() string {
	return "configurable creator fees are disabled (Global.creatorFeeConfigurable is false)"
}

// CreatorFeeBpsOutOfRangeError mirrors pump error 6078.
type CreatorFeeBpsOutOfRangeError struct {
	CreatorFeeBps uint64
	Max           uint64
}

func (e *CreatorFeeBpsOutOfRangeError) Error() string {
	return fmt.Sprintf("creator fee of %d bps is outside 1..=%d (Global.maxConfigurableCreatorFeeBps)", e.CreatorFeeBps, e.Max)
}

// CashbackCoinNotSupportedError mirrors pump error 6094.
type CashbackCoinNotSupportedError struct {
	Mint solana.PublicKey
}

func (e *CashbackCoinNotSupportedError) Error() string {
	return fmt.Sprintf("v3 trades do not support cashback coin %s; use the v2 instructions", e.Mint)
}

// QuotePoolNotFoundError mirrors pump error 6102: a migrated quote coin has no canonical pool.
type QuotePoolNotFoundError struct {
	QuoteMint solana.PublicKey
	Pool      solana.PublicKey
}

func (e *QuotePoolNotFoundError) Error() string {
	return fmt.Sprintf("pump coin %s migrated but canonical pool %s was not found", e.QuoteMint, e.Pool)
}
