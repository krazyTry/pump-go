package bonding_curve

import (
	solana "github.com/gagliardetto/solana-go"
)

// DonationRelay

func DeriveDonationRelayEventAuthority() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("__event_authority")}, DonationRelayProgramID)
	return key
}

func DeriveDonationRelayEpochTrackerPDA(configID, quoteMint solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("epoch_tracker_v1"), configID.Bytes(), quoteMint.Bytes()}, DonationRelayProgramID)
	return key
}

func DeriveDonationRelayDebouncerPDA(configID, quoteMint solana.PublicKey) solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("debouncer_v1"), configID.Bytes(), quoteMint.Bytes()}, DonationRelayProgramID)
	return key
}

func DeriveDonationRelayMintWhitelistPDA() solana.PublicKey {
	key, _, _ := solana.FindProgramAddress([][]byte{[]byte("mint_whitelist_v1")}, DonationRelayProgramID)
	return key
}
