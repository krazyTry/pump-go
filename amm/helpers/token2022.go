package helpers

import (
	"encoding/binary"
	"errors"
	"fmt"

	solana "github.com/gagliardetto/solana-go"
)

const (
	// Token mint base size (Token-2020 compatible header)
	MintBaseSize = 82 // :contentReference[oaicite:2]{index=2}

	// Token-2022 extension types (from SPL Token JS docs) :contentReference[oaicite:3]{index=3}
	ExtUninitialized     uint16 = 0
	ExtTransferFeeConfig uint16 = 1
	ExtTransferHook      uint16 = 14
)

// Extensions holds raw TLV slices + decoded structs you care about.
type Extensions struct {
	Raw map[uint16][]byte

	TransferFeeConfig *TransferFeeConfig
	HasTransferHook   bool
}

type TransferFee struct {
	Epoch  uint64
	MaxFee uint64
	FeeBps uint16
}

type TransferFeeConfig struct {
	// Authorities are stored as COption<Pubkey> on-chain; nil means "None".
	TransferFeeConfigAuthority *solana.PublicKey
	WithdrawWithheldAuthority  *solana.PublicKey

	WithheldAmount uint64

	Older TransferFee
	Newer TransferFee
}

// FeeForEpoch picks older/newer based on current epoch.
// SPL Token JS docs describe older used if currentEpoch < newer.epoch, else newer. :contentReference[oaicite:4]{index=4}
func (c *TransferFeeConfig) FeeForEpoch(currentEpoch uint64) TransferFee {
	if currentEpoch < c.Newer.Epoch {
		return c.Older
	}
	return c.Newer
}

// parseToken2022Extensions parses TLV extensions from a Token-2022 *Mint* account data.
//
// data: account data bytes (base64 decoded)
// returns:
// - Extensions.Raw: map[extType]extData
// - Extensions.TransferFeeConfig decoded if present
func parseToken2022Extensions(data []byte) (*Extensions, error) {
	if len(data) < MintBaseSize {
		return nil, fmt.Errorf("data too short for mint base: got=%d want>=%d", len(data), MintBaseSize)
	}

	exts := &Extensions{
		Raw: make(map[uint16][]byte),
	}

	off := MintBaseSize
	for {
		// Need at least 4 bytes for TLV header: u16 type + u16 length
		if off+4 > len(data) {
			break
		}

		typ := binary.LittleEndian.Uint16(data[off : off+2])
		l := binary.LittleEndian.Uint16(data[off+2 : off+4])
		off += 4

		// Convention: trailing zero padding often appears; stop on (0,0).
		if typ == ExtUninitialized && l == 0 {
			break
		}

		if off+int(l) > len(data) {
			return nil, fmt.Errorf("invalid TLV length: type=%d len=%d off=%d total=%d", typ, l, off, len(data))
		}

		val := data[off : off+int(l)]
		off += int(l)

		// Store raw
		exts.Raw[typ] = val

		// Decode the ones we care about
		switch typ {
		case ExtTransferFeeConfig:
			cfg, err := parseTransferFeeConfig(val)
			if err != nil {
				return nil, fmt.Errorf("parse TransferFeeConfig failed: %w", err)
			}
			exts.TransferFeeConfig = cfg
		case ExtTransferHook:
			exts.HasTransferHook = true
		}

		// Optional: if remaining bytes are all zeros, you can break early.
		// (Leave it simple; loop will stop naturally when it can't read next header.)
	}

	return exts, nil
}

// --- internal decoders ---

func parseTransferFeeConfig(b []byte) (*TransferFeeConfig, error) {
	// Layout is fixed-size in practice, but can evolve; parse minimally and ignore any extra bytes.
	// Expected fields (conceptually):
	// - COption<Pubkey> transfer_fee_config_authority
	// - COption<Pubkey> withdraw_withheld_authority
	// - u64 withheld_amount
	// - TransferFee older
	// - TransferFee newer
	//
	// We decode:
	// - COption<Pubkey> is 4-byte tag (0 or 1) + 32 bytes when Some
	// - TransferFee: epoch u64 + maximum_fee u64 + transfer_fee_basis_points u16 (+ possible padding)
	off := 0

	auth1, n, err := readCOptionPubkey(b, off)
	if err != nil {
		return nil, err
	}
	off += n

	auth2, n, err := readCOptionPubkey(b, off)
	if err != nil {
		return nil, err
	}
	off += n

	if off+8 > len(b) {
		return nil, errors.New("transfer fee config: truncated withheld_amount")
	}
	withheld := binary.LittleEndian.Uint64(b[off : off+8])
	off += 8

	older, n, err := readTransferFee(b, off)
	if err != nil {
		return nil, err
	}
	off += n

	newer, n, err := readTransferFee(b, off)
	if err != nil {
		return nil, err
	}
	off += n

	return &TransferFeeConfig{
		TransferFeeConfigAuthority: auth1,
		WithdrawWithheldAuthority:  auth2,
		WithheldAmount:             withheld,
		Older:                      older,
		Newer:                      newer,
	}, nil
}

func readCOptionPubkey(b []byte, off int) (*solana.PublicKey, int, error) {
	if off+4 > len(b) {
		return nil, 0, errors.New("COption<Pubkey>: truncated tag")
	}
	tag := binary.LittleEndian.Uint32(b[off : off+4])
	off += 4

	switch tag {
	case 0:
		// None
		return nil, 4, nil
	case 1:
		if off+32 > len(b) {
			return nil, 0, errors.New("COption<Pubkey>: truncated pubkey")
		}
		pk := solana.PublicKeyFromBytes(b[off : off+32])
		return &pk, 4 + 32, nil
	default:
		return nil, 0, fmt.Errorf("COption<Pubkey>: invalid tag=%d", tag)
	}
}

func readTransferFee(b []byte, off int) (TransferFee, int, error) {
	// Minimum bytes: 8 + 8 + 2 = 18
	if off+18 > len(b) {
		return TransferFee{}, 0, errors.New("TransferFee: truncated")
	}
	epoch := binary.LittleEndian.Uint64(b[off : off+8])
	maxFee := binary.LittleEndian.Uint64(b[off+8 : off+16])
	bps := binary.LittleEndian.Uint16(b[off+16 : off+18])

	// Some implementations pad to 24 bytes (align); TLV length tells the true size.
	// Here we advance by 24 if enough bytes remain AND the surrounding structure expects padding.
	// Safer strategy: if there are at least 24 bytes left before next field, consume 24; otherwise consume 18.
	// But since we're parsing inside a known struct, we can prefer 24 when available.
	advance := 18
	if off+24 <= len(b) {
		advance = 24
	}

	return TransferFee{
		Epoch:  epoch,
		MaxFee: maxFee,
		FeeBps: bps,
	}, advance, nil
}
