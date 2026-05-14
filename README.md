# Pump SDK library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/krazyTry/pump-go.svg)](https://pkg.go.dev/github.com/krazyTry/pump-go)
[![GitHub tag (latest SemVer pre-release)](https://img.shields.io/github/v/tag/krazyTry/pump-go?include_prereleases&label=release-tag)](https://github.com/krazyTry/pump-go/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/krazyTry/pump-go)](https://goreportcard.com/report/github.com/krazyTry/pump-go)
[![Open Source Helpers](https://www.codetriage.com/krazytry/pump-go/badges/users.svg)](https://www.codetriage.com/krazytry/pump-go)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/krazyTry/pump-go/blob/master/LICENSE.md)

Go SDK for interacting with the **Pump Protocol** on Solana.  
Currently supports **Bonding Curves v0.1.0** and **AMM v0.1.0**.

---

<div align="center">
    <img src="https://user-images.githubusercontent.com/15271561/128235229-1d2d9116-23bb-464e-b2cc-8fb6355e3b55.png" margin="auto" height="175"/>
</div>

## Features

 * **Bonding Curves – Create and manage token launch pools with bonding curves.
 * **AMM – Seamlessly migrate BC pools to AMM and interact with automated market makers.
 * **Liquidity Management – Add or remove liquidity, manage positions, and claim accrued fees.
 * **Token Swaps – Execute buy/sell operations with slippage protection and accurate quotations.
 * **Multi-Token Support – Fully compatible with SPL tokens and Token-2022.
 * **Real-Time Quotes – Retrieve precise swap quotations before transaction execution.
 * **Fee Management – Calculate and collect fees automatically.

## Install

Run `go get github.com/krazyTry/pump-go`

## Requirements 

Meteora SDK requires Go version `>=1.25.1`

## Documentation

https://pkg.go.dev/github.com/krazyTry/pump-go


## Usage

### Bonding Curve Example

Read tests in `tests/bonding_curve`


### AMM Example

Read tests in `tests/amm`


## Related Projects

 * [gagliardetto/solana-go](https://github.com/gagliardetto/solana-go) - Core Solana Go SDK
 * [gagliardetto/anchor-go](https://github.com/gagliardetto/anchor-go) - Anchor framework bindings for Go
 * [gagliardetto/binary](https://github.com/gagliardetto/binary) - Binary encoding/decoding utilities
 * [Pump Docs](https://github.com/pump-fun/pump-public-docs) - Protocol documentation

## FAQ

#### How can I quickly find my AMM?

```go
var ammService = amm.NewClient(rpcClient, rpc.CommitmentFinalized)

func GetPoolAmm(baseMint, creator solana.PublicKey) (*amm.AccountWithPool, error) {
	pools, err := ammService.FetchPoolByBaseMint(ctx, baseMint)
	if err != nil {
		t.Fatal("cpAmm.FetchPoolStatesByTokenAMint() fail", err)
	}

	var p *amm.AccountWithPool

	for _,pool :=range pools {
		if pool.Account.Creator != creator {
			continue
		}
		p = pool
		break
	}

	return p, nil
}
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

**Note:** This SDK provides Go bindings for the Pump Protocol on Solana and is **not affiliated with the official Pump team**.
