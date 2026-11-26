package helpers

import (
	"math/rand"
	"time"

	solana "github.com/solana-foundation/solana-go/v2"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

var protocolFeeRecipients = []solana.PublicKey{
	solana.MustPublicKeyFromBase58("3Mqk6qHEj7r3TFNFYjCP85a5venbp2mefYQTSKVVchBC"),
	solana.MustPublicKeyFromBase58("A3cm36KqisUXVTSRCK6VQ7kvXLene4gsbQedhcd5MujQ"),
	solana.MustPublicKeyFromBase58("BWNNnhN6id5j71h5wfkzmXjnV2fgF9zA61FqGJJjnakw"),
	solana.MustPublicKeyFromBase58("DrcDxWxLi3i7Yd7cTyCyd44DE7pwS9BvFxSsPx1c7XTv"),
	solana.MustPublicKeyFromBase58("EAN7utdjyxhKVaqA69DRyxW5ffBhnidRZfifmfhwUgPm"),
	solana.MustPublicKeyFromBase58("Eh7hTLi6iMbT55jAEYVgtN5QaSb6z9Ytzwxs2LGG5qg4"),
	solana.MustPublicKeyFromBase58("Fso3Ggq2Um1pQT6Uzq8wDicgRh62vANYZUT8uqJd2fXh"),
}

func GetStaticRandomProtocolFeeRecipient() solana.PublicKey {
	return protocolFeeRecipients[rnd.Intn(len(protocolFeeRecipients))]
}

var reservedFeeRecipients = []solana.PublicKey{
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
