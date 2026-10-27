package bonding_curve

import (
	"errors"
	"fmt"

	"github.com/krazyTry/pump-go/bonding_curve/helpers"
	ammgen "github.com/krazyTry/pump-go/gen/amm"
	solana "github.com/solana-foundation/solana-go/v2"
)

// MultiHopInstructionHop is one canonical pool or bonding curve on a multi_hop_swap route.
// Venue is "pool" or "curve". A SOL quote may be the zero key or WSOL.
type MultiHopInstructionHop struct {
	Venue             string
	BaseMint          solana.PublicKey
	QuoteMint         solana.PublicKey
	BaseTokenProgram  solana.PublicKey
	QuoteTokenProgram solana.PublicKey
}

// MultiHopMint is one end of a route.
type MultiHopMint struct {
	Mint         solana.PublicKey
	TokenProgram solana.PublicKey
}

// MultiHopEnds is where a route starts, ends, and which mint pays the protocol fee.
type MultiHopEnds struct {
	Input    MultiHopMint
	Output   MultiHopMint
	Currency MultiHopMint
}

func normalizeHop(hop MultiHopInstructionHop) MultiHopInstructionHop {
	side := func(mint, tokenProgram solana.PublicKey) (solana.PublicKey, solana.PublicKey) {
		if IsLegacyQuoteMint(mint) {
			return solana.WrappedSol, solana.TokenProgramID
		}
		return mint, tokenProgram
	}
	base, baseProgram := side(hop.BaseMint, hop.BaseTokenProgram)
	quote, quoteProgram := side(hop.QuoteMint, hop.QuoteTokenProgram)
	return MultiHopInstructionHop{
		Venue:             hop.Venue,
		BaseMint:          base,
		QuoteMint:         quote,
		BaseTokenProgram:  baseProgram,
		QuoteTokenProgram: quoteProgram,
	}
}

// MultiHopRouteEnds returns the mints a route starts and ends in.
// side is "buy" or "sell". SOL comes back as WSOL.
func MultiHopRouteEnds(raw []MultiHopInstructionHop, side string) (MultiHopEnds, error) {
	if len(raw) == 0 {
		return MultiHopEnds{}, errors.New("multi_hop_swap: a route needs at least one hop")
	}
	if side != "buy" && side != "sell" {
		return MultiHopEnds{}, fmt.Errorf("multi_hop_swap: side %q must be buy or sell", side)
	}
	hops := make([]MultiHopInstructionHop, len(raw))
	for i, hop := range raw {
		hops[i] = normalizeHop(hop)
	}
	isBuy := side == "buy"
	for index, hop := range hops {
		if index+1 == len(hops) {
			break
		}
		next := hops[index+1]
		got, want := next.BaseMint, hop.QuoteMint
		if isBuy {
			got, want = next.QuoteMint, hop.BaseMint
		}
		if !got.Equals(want) {
			return MultiHopEnds{}, fmt.Errorf("multi_hop_swap: hop %d does not continue from hop %d", index+1, index)
		}
	}
	first, last := hops[0], hops[len(hops)-1]
	input := MultiHopMint{Mint: first.BaseMint, TokenProgram: first.BaseTokenProgram}
	output := MultiHopMint{Mint: last.QuoteMint, TokenProgram: last.QuoteTokenProgram}
	currency := output
	if isBuy {
		input = MultiHopMint{Mint: first.QuoteMint, TokenProgram: first.QuoteTokenProgram}
		output = MultiHopMint{Mint: last.BaseMint, TokenProgram: last.BaseTokenProgram}
		currency = input
	}
	return MultiHopEnds{Input: input, Output: output, Currency: currency}, nil
}

func multiHopHopAccounts(raw MultiHopInstructionHop) []*solana.AccountMeta {
	hop := normalizeHop(raw)
	venue := DeriveBondingCurve(hop.BaseMint)
	if hop.Venue == "pool" {
		venue = DeriveCanonicalPoolPDAWithQuote(DerivePoolAuthority(hop.BaseMint), hop.BaseMint, hop.QuoteMint)
	}
	return []*solana.AccountMeta{
		solana.NewAccountMeta(hop.BaseMint, false, false),
		solana.NewAccountMeta(hop.QuoteMint, false, false),
		solana.NewAccountMeta(venue, true, false),
		solana.NewAccountMeta(helpers.FindAssociatedTokenAddress(venue, hop.BaseMint, hop.BaseTokenProgram), true, false),
		solana.NewAccountMeta(helpers.FindAssociatedTokenAddress(venue, hop.QuoteMint, hop.QuoteTokenProgram), true, false),
	}
}

// MultiHopSwapInstruction builds pump-amm multi_hop_swap.
// buybackFeeRecipient is the wallet; the zero key uses the first listed buyback wallet.
func (s *Client) MultiHopSwapInstruction(user solana.PublicKey, hops []MultiHopInstructionHop, side string, amountIn, minAmountOut uint64, buybackFeeRecipient solana.PublicKey) (solana.Instruction, error) {
	ends, err := MultiHopRouteEnds(hops, side)
	if err != nil {
		return nil, err
	}
	if buybackFeeRecipient.Equals(solana.PublicKey{}) {
		buybackFeeRecipient = helpers.GetBuybackFeeRecipientForTokenQuote()
	}
	ix, err := ammgen.NewMultiHopSwapInstruction(
		amountIn,
		minAmountOut,
		user,
		helpers.FindAssociatedTokenAddress(user, ends.Input.Mint, ends.Input.TokenProgram),
		helpers.FindAssociatedTokenAddress(user, ends.Output.Mint, ends.Output.TokenProgram),
		s.AmmGlobalConfig,
		DeriveAmmFeeConfig(),
		DeriveAmmUserVolumeAccumulator(user),
		helpers.FindAssociatedTokenAddress(buybackFeeRecipient, ends.Currency.Mint, ends.Currency.TokenProgram),
		solana.TokenProgramID,
		solana.Token2022ProgramID,
		solana.SystemProgramID,
		s.AmmEventAuthority,
		AmmProgramID,
		ProgramID,
		s.Global,
		s.FeeConfig,
		s.EventAuthority,
	)
	if err != nil {
		return nil, err
	}
	gi := ix.(*solana.GenericInstruction)
	for _, hop := range hops {
		gi.AccountValues = append(gi.AccountValues, multiHopHopAccounts(hop)...)
	}
	return ix, nil
}

// MultiHopSwapInstructions adds the user's ATA creates, a WSOL wrap when the route starts at a pool, and WSOL closes.
func (s *Client) MultiHopSwapInstructions(user solana.PublicKey, hops []MultiHopInstructionHop, side string, amountIn, minAmountOut uint64, buybackFeeRecipient solana.PublicKey) ([]solana.Instruction, error) {
	ends, err := MultiHopRouteEnds(hops, side)
	if err != nil {
		return nil, err
	}
	if buybackFeeRecipient.Equals(solana.PublicKey{}) {
		buybackFeeRecipient = helpers.GetBuybackFeeRecipientForTokenQuote()
	}
	var out []solana.Instruction
	create := func(end MultiHopMint) {
		ata := helpers.FindAssociatedTokenAddress(user, end.Mint, end.TokenProgram)
		out = append(out, helpers.CreateAssociatedTokenAccountInstruction(user, ata, user, end.Mint, end.TokenProgram))
	}
	solIn := ends.Input.Mint.Equals(solana.WrappedSol)
	if solIn {
		create(ends.Input)
	}
	if solIn && len(hops) > 0 && normalizeHop(hops[0]).Venue == "pool" {
		ata := helpers.FindAssociatedTokenAddress(user, ends.Input.Mint, ends.Input.TokenProgram)
		wrap, err := helpers.WrapSOLInstruction(user, ata, amountIn)
		if err != nil {
			return nil, err
		}
		out = append(out, wrap...)
	}
	create(ends.Output)
	buybackATA := helpers.FindAssociatedTokenAddress(buybackFeeRecipient, ends.Currency.Mint, ends.Currency.TokenProgram)
	out = append(out, helpers.CreateAssociatedTokenAccountInstruction(user, buybackATA, buybackFeeRecipient, ends.Currency.Mint, ends.Currency.TokenProgram))
	ix, err := s.MultiHopSwapInstruction(user, hops, side, amountIn, minAmountOut, buybackFeeRecipient)
	if err != nil {
		return nil, err
	}
	out = append(out, ix)
	close := func(end MultiHopMint) {
		if !end.Mint.Equals(solana.WrappedSol) {
			return
		}
		unwrap, err := helpers.UnwrapSOLInstruction(user, user, true)
		if err == nil && unwrap != nil {
			out = append(out, unwrap)
		}
	}
	close(ends.Input)
	close(ends.Output)
	return out, nil
}

// DeriveAmmFeeConfig is the fee_config PDA seeded with the AMM program.
func DeriveAmmFeeConfig() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("fee_config"), AmmProgramID.Bytes()}, FeeProgramID)
	return key
}

// DeriveAmmUserVolumeAccumulator is the AMM user volume accumulator.
func DeriveAmmUserVolumeAccumulator(user solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("user_volume_accumulator"), user.Bytes()}, AmmProgramID)
	return key
}
