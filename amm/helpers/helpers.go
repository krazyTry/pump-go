package helpers

import (
	"math/rand"
	"time"

	solana "github.com/solana-foundation/solana-go/v2"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

var protocolFeeRecipients = []solana.PublicKey{
	solana.MustPublicKeyFromBase58("62qc2CNXwrYqQScmEdiZFFAnJR262PxWEuNQtxfafNgV"),
	solana.MustPublicKeyFromBase58("7VtfL8fvgNfhz17qKRMjzQEXgbdpnHHHQRh54R9jP2RJ"),
	solana.MustPublicKeyFromBase58("7hTckgnGnLQR6sdH7YkqFTAA7VwTfYFaZ6EhEsU3saCX"),
	solana.MustPublicKeyFromBase58("9rPYyANsfQZw3DnDmKE3YCQF5E8oD89UXoHn9JFEhJUz"),
	solana.MustPublicKeyFromBase58("AVmoTthdrX6tKt4nDjco2D775W2YK3sDhxPcMmzUAmTY"),
	solana.MustPublicKeyFromBase58("FWsW1xNtWscwNmKv6wVsU1iTzRN6wmmk3MjxRP5tT7hz"),
	solana.MustPublicKeyFromBase58("G5UZAVbAf46s7cKWoyKu8kYTip9DGTpbLZ2qa9Aq69dP"),
	solana.MustPublicKeyFromBase58("JCRGumoE9Qi5BBgULTgdgTLjSgkCMSbF62ZZfGs84JeU"),
}

func GetStaticRandomProtocolFeeRecipient() solana.PublicKey {
	return protocolFeeRecipients[rnd.Intn(len(protocolFeeRecipients))]
}

var reservedFeeRecipients = []solana.PublicKey{
	solana.MustPublicKeyFromBase58("GesfTA3X2arioaHp8bbKdjG9vJtskViWACZoYvxp4twS"),
	solana.MustPublicKeyFromBase58("4budycTjhs9fD6xw62VBducVTNgMgJJ5BgtKq7mAZwn6"),
	solana.MustPublicKeyFromBase58("8SBKzEQU4nLSzcwF4a74F2iaUDQyTfjGndn6qUWBnrpR"),
	solana.MustPublicKeyFromBase58("4UQeTP1T39KZ9Sfxzo3WR5skgsaP6NZa87BAkuazLEKH"),
	solana.MustPublicKeyFromBase58("8sNeir4QsLsJdYpc9RZacohhK1Y5FLU3nC5LXgYB4aa6"),
	solana.MustPublicKeyFromBase58("Fh9HmeLNUMVCvejxCtCL2DbYaRyBFVJ5xrWkLnMH6fdk"),
	solana.MustPublicKeyFromBase58("463MEnMeGyJekNZFQSTUABBEbLnvMTALbT6ZmsxAbAdq"),
	solana.MustPublicKeyFromBase58("6AUH3WEHucYZyC61hqpqYUWVto5qA5hjHuNQ32GNnNxA"),
}

func GetStaticRandomReservedFeeRecipient() solana.PublicKey {
	return reservedFeeRecipients[rnd.Intn(len(reservedFeeRecipients))]
}

var buybackFeeRecipients = []solana.PublicKey{
	solana.MustPublicKeyFromBase58("5YxQFdt3Tr9zJLvkFccqXVUwhdTWJQc1fFg2YPbxvxeD"),
	solana.MustPublicKeyFromBase58("9M4giFFMxmFGXtc3feFzRai56WbBqehoSeRE5GK7gf7"),
	solana.MustPublicKeyFromBase58("GXPFM2caqTtQYC2cJ5yJRi9VDkpsYZXzYdwYpGnLmtDL"),
	solana.MustPublicKeyFromBase58("3BpXnfJaUTiwXnJNe7Ej1rcbzqTTQUvLShZaWazebsVR"),
	solana.MustPublicKeyFromBase58("5cjcW9wExnJJiqgLjq7DEG75Pm6JBgE1hNv4B2vHXUW6"),
	solana.MustPublicKeyFromBase58("EHAAiTxcdDwQ3U4bU6YcMsQGaekdzLS3B5SmYo46kJtL"),
	solana.MustPublicKeyFromBase58("5eHhjP8JaYkz83CWwvGU2uMUXefd3AazWGx4gpcuEEYD"),
	solana.MustPublicKeyFromBase58("A7hAgCzFw14fejgCp387JUJRMNyz4j89JKnhtKU8piqW"),
}

func GetStaticRandomBuybackFeeRecipient() solana.PublicKey {
	return buybackFeeRecipients[rnd.Intn(len(buybackFeeRecipients))]
}
