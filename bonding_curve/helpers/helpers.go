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
	solana.MustPublicKeyFromBase58("CebN5WGQ4jvEPvsVU4EoHEpgzq1VV7AbicfhtW4xC9iM"),
	solana.MustPublicKeyFromBase58("FWsW1xNtWscwNmKv6wVsU1iTzRN6wmmk3MjxRP5tT7hz"),
	solana.MustPublicKeyFromBase58("G5UZAVbAf46s7cKWoyKu8kYTip9DGTpbLZ2qa9Aq69dP"),
}

func GetStaticRandomProtocolFeeRecipient() solana.PublicKey {
	return protocolFeeRecipients[rnd.Intn(len(protocolFeeRecipients))]
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
