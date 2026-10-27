import { createRequire } from "module";
import { pathToFileURL } from "url";

const require = createRequire("/code/go/pump-bc-sdk/package.json");
const { PublicKey } = require("@solana/web3.js");
const BN = require("bn.js");
const sdk = require("/code/go/pump-bc-sdk/dist/index.js");

const {
  getBuyTokenAmountFromSolAmount,
  getBuySolAmountFromTokenAmount,
  getSellSolAmountFromTokenAmount,
  getBuyV3QuoteAmountFromTokenAmount,
  getBuyV3TokenAmountFromQuoteAmount,
  newBondingCurve,
  initialVirtualQuoteReservesFor,
  pumpQuoteReserves,
  bondingCurveMarketCap,
  computeFeesBps,
  MAX_CURVE_DEPTH,
} = sdk;

const zero = PublicKey.default;
const wsol = new PublicKey("So11111111111111111111111111111111111111112");
const usdc = new PublicKey("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v");
const exotic = new PublicKey("pumpCmXqMfrsAkQ5r49WcJnRayYRqmXz6ae8H7H9Dfn");
const native2022 = new PublicKey("9pan9bMn5HatX4EJdBwg9VgCa7Uz5HL8N1m5D3NdXejP");
const creator = new PublicKey("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA");
const quoteControlMint = new PublicKey("4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU");

const global = {
  initialVirtualTokenReserves: new BN("1073000000000000"),
  initialVirtualSolReserves: new BN("30000000000"),
  initialRealTokenReserves: new BN("793100000000000"),
  tokenTotalSupply: new BN("1000000000000000"),
  feeBasisPoints: new BN(100),
  creatorFeeBasisPoints: new BN(5),
  poolMigrationFee: new BN("15000000"),
  initialVirtualQuoteReserves: new BN("4292000000"),
  whitelistedQuoteMints: [usdc],
  creatorFeeConfigurable: true,
  maxConfigurableCreatorFeeBps: new BN(500),
  maxCurveDepth: 3,
  mayhemModeEnabled: false,
  buybackBasisPoints: new BN(1000),
  isHolderRewardEnabled: false,
};

const feeConfig = {
  feeTiers: [
    { marketCapLamportsThreshold: new BN(0), fees: { lpFeeBps: new BN(0), protocolFeeBps: new BN(95), creatorFeeBps: new BN(30) } },
    { marketCapLamportsThreshold: new BN("10000000000"), fees: { lpFeeBps: new BN(0), protocolFeeBps: new BN(50), creatorFeeBps: new BN(20) } },
  ],
  stableFeeTiers: [
    { marketCapLamportsThreshold: new BN(0), fees: { lpFeeBps: new BN(0), protocolFeeBps: new BN(10), creatorFeeBps: new BN(5) } },
    { marketCapLamportsThreshold: new BN("10000000000"), fees: { lpFeeBps: new BN(0), protocolFeeBps: new BN(5), creatorFeeBps: new BN(2) } },
  ],
  flatFees: { lpFeeBps: new BN(0), protocolFeeBps: new BN(30), creatorFeeBps: new BN(10) },
  exoticFlatFees: { lpFeeBps: new BN(0), protocolFeeBps: new BN(80), creatorFeeBps: new BN(40) },
};

function curve(over) {
  return {
    virtualTokenReserves: new BN("1073000000000000"),
    virtualQuoteReserves: new BN("30000000000"),
    realTokenReserves: new BN("793100000000000"),
    realQuoteReserves: new BN(0),
    tokenTotalSupply: new BN("1000000000000000"),
    complete: false,
    creator,
    isMayhemMode: false,
    isCashbackCoin: false,
    quoteMint: zero,
    creatorFeeBps: new BN(0),
    canEditCreatorFee: false,
    isHolderReward: false,
    creatorFee: new BN(0),
    protocolFees: new BN(0),
    depth: 0,
    initialVirtualQuoteReserves: new BN("30000000000"),
    postCompleteBaseOut: new BN(0),
    postCompleteQuoteIn: new BN(0),
    ...over,
  };
}

const out = {};
const solCurve = curve();
const supply = new BN("1000000000000000");
out.buyTokens1Sol = getBuyTokenAmountFromSolAmount({
  global, feeConfig, mintSupply: supply, bondingCurve: solCurve, amount: new BN("1000000000"), quoteMint: zero,
}).toString();
out.buySolForTokens = getBuySolAmountFromTokenAmount({
  global, feeConfig, mintSupply: supply, bondingCurve: solCurve, amount: new BN("1000000000000"), quoteMint: zero,
}).toString();
out.sellSol = getSellSolAmountFromTokenAmount({
  global, feeConfig, mintSupply: supply, bondingCurve: solCurve, amount: new BN("1000000000000"),
}).toString();

out.buyTokensLowSupply = getBuyTokenAmountFromSolAmount({
  global, feeConfig, mintSupply: new BN("100000000000000"), bondingCurve: solCurve, amount: new BN("1000000000"), quoteMint: zero,
}).toString();
out.buySolLowSupply = getBuySolAmountFromTokenAmount({
  global, feeConfig, mintSupply: new BN("100000000000000"), bondingCurve: solCurve, amount: new BN("1000000000000"), quoteMint: zero,
}).toString();

const usdcCurve = curve({
  virtualQuoteReserves: new BN("5000000000"),
  quoteMint: usdc,
  initialVirtualQuoteReserves: new BN("5000000000"),
  creatorFeeBps: new BN(250),
});
out.buyUsdcTokens = getBuyTokenAmountFromSolAmount({
  global, feeConfig, mintSupply: supply, bondingCurve: usdcCurve, amount: new BN("1000000"), quoteMint: usdc,
}).toString();
out.buyUsdcCost = getBuySolAmountFromTokenAmount({
  global, feeConfig, mintSupply: supply, bondingCurve: usdcCurve, amount: new BN("1000000000000"), quoteMint: usdc,
}).toString();

const exoticCurve = curve({
  virtualQuoteReserves: new BN("8000000000"),
  quoteMint: exotic,
  initialVirtualQuoteReserves: new BN("8000000000"),
});
out.buyExotic = getBuyTokenAmountFromSolAmount({
  global, feeConfig, mintSupply: supply, bondingCurve: exoticCurve, amount: new BN("1000000"), quoteMint: exotic,
}).toString();

const fee = { basisPoints: 50, maximumFee: new BN("1000000000") };
out.buyWithTransferFee = getBuyTokenAmountFromSolAmount({
  global, feeConfig, mintSupply: supply, bondingCurve: solCurve, amount: new BN("1000000000"), quoteMint: zero, quoteTransferFee: fee,
}).toString();
out.buySolWithTransferFee = getBuySolAmountFromTokenAmount({
  global, feeConfig, mintSupply: supply, bondingCurve: solCurve, amount: new BN("1000000000000"), quoteMint: zero, quoteTransferFee: fee,
}).toString();
out.sellWithTransferFee = getSellSolAmountFromTokenAmount({
  global, feeConfig, mintSupply: supply, bondingCurve: solCurve, amount: new BN("1000000000000"), quoteTransferFee: fee,
}).toString();

out.marketCap = bondingCurveMarketCap({
  mintSupply: supply,
  virtualQuoteReserves: solCurve.virtualQuoteReserves,
  virtualTokenReserves: solCurve.virtualTokenReserves,
}).toString();

out.feesSol = computeFeesBps({
  global, feeConfig, mintSupply: supply,
  virtualQuoteReserves: solCurve.virtualQuoteReserves,
  virtualTokenReserves: solCurve.virtualTokenReserves,
  quoteMint: zero,
  creatorFeeBps: new BN(0),
});
out.feesSol = { protocol: out.feesSol.protocolFeeBps.toString(), creator: out.feesSol.creatorFeeBps.toString() };
out.feesUsdc = computeFeesBps({
  global, feeConfig, mintSupply: supply,
  virtualQuoteReserves: usdcCurve.virtualQuoteReserves,
  virtualTokenReserves: usdcCurve.virtualTokenReserves,
  quoteMint: usdc,
  creatorFeeBps: new BN(250),
});
out.feesUsdc = { protocol: out.feesUsdc.protocolFeeBps.toString(), creator: out.feesUsdc.creatorFeeBps.toString() };

const fresh = newBondingCurve(global, zero);
out.freshSolVirtual = fresh.virtualQuoteReserves.toString();
out.freshSolQuote = fresh.quoteMint.toBase58();
const freshUsdc = newBondingCurve(global, usdc);
out.freshUsdcVirtual = freshUsdc.virtualQuoteReserves.toString();
const control = { mints: [{ mint: quoteControlMint, initialVirtualQuoteReserves: new BN("123456789") }] };
out.freshControl = initialVirtualQuoteReservesFor(global, quoteControlMint, control).toString();
let unsupported = "";
try {
  initialVirtualQuoteReservesFor(global, exotic, null);
} catch (e) {
  unsupported = e.name;
}
out.unsupported = unsupported;
let nativeErr = "";
try {
  initialVirtualQuoteReservesFor(global, native2022, null);
} catch (e) {
  nativeErr = e.name;
}
out.native2022 = nativeErr;

const v3 = curve({
  virtualTokenReserves: new BN(5000),
  virtualQuoteReserves: new BN(2000),
  realTokenReserves: new BN(1000),
  realQuoteReserves: new BN(500),
  tokenTotalSupply: new BN(100000),
});
const v3Global = { ...global, poolMigrationFee: new BN(50), initialRealTokenReserves: new BN(8000) };
out.v3Within = getBuyV3QuoteAmountFromTokenAmount({
  global: v3Global, feeConfig, mintSupply: supply, bondingCurve: v3, amount: new BN(400), curveBaseTokenBalance: new BN(5000),
}).toString();
out.v3Cross = getBuyV3QuoteAmountFromTokenAmount({
  global: v3Global, feeConfig, mintSupply: supply, bondingCurve: v3, amount: new BN(1500), curveBaseTokenBalance: new BN(5000),
}).toString();
out.v3Exact = getBuyV3TokenAmountFromQuoteAmount({
  global: v3Global, feeConfig, mintSupply: supply, bondingCurve: v3, amount: new BN(3000), curveBaseTokenBalance: new BN(5000),
}).toString();
out.v3ExactCross = getBuyV3TokenAmountFromQuoteAmount({
  global: v3Global, feeConfig, mintSupply: supply, bondingCurve: v3, amount: new BN(out.v3Cross), curveBaseTokenBalance: new BN(5000),
}).toString();
const mayhem = curve({ ...v3, isMayhemMode: true });
out.v3Mayhem = getBuyV3TokenAmountFromQuoteAmount({
  global: v3Global, feeConfig, mintSupply: supply, bondingCurve: mayhem, amount: new BN("100000000"), curveBaseTokenBalance: new BN(5000),
}).toString();

const rootMint = new PublicKey("11111111111111111111111111111112");
const migrated = curve({
  virtualTokenReserves: new BN(0),
  virtualQuoteReserves: new BN(0),
  realTokenReserves: new BN(0),
  realQuoteReserves: new BN(0),
  complete: true,
  depth: 0,
  quoteMint: zero,
  creator: zero,
  isMayhemMode: false,
});
const pumpQuote = {
  mint: rootMint,
  bondingCurve: migrated,
  pool: { baseReserves: new BN("1000000000"), quoteReserves: new BN("2000000000"), virtualQuoteReserves: new BN(0) },
  quoteControl: null,
  chain: [],
};
const seeded = pumpQuoteReserves(global, pumpQuote, null);
out.pumpVirtual = seeded.virtualQuoteReserves.toString();
out.pumpDepth = seeded.depth;
out.maxDepth = MAX_CURVE_DEPTH;

const deep = curve({ depth: 5, quoteMint: zero, isMayhemMode: false, creator: zero });
try {
  pumpQuoteReserves(global, { mint: rootMint, bondingCurve: deep, quoteControl: null }, null);
  out.deep = "ok";
} catch (e) {
  out.deep = e.name + ":" + e.depth + ":" + e.maxCurveDepth;
}

const trading = curve({
  virtualTokenReserves: new BN("1073000000000000"),
  virtualQuoteReserves: new BN("30000000000"),
  realTokenReserves: new BN("793100000000000"),
  realQuoteReserves: new BN(0),
  complete: false,
  depth: 0,
  quoteMint: zero,
  creator: zero,
});
const tradingSeed = pumpQuoteReserves(global, {
  mint: rootMint,
  bondingCurve: trading,
  quoteControl: null,
}, null);
out.tradingVirtual = tradingSeed.virtualQuoteReserves.toString();
out.tradingDepth = tradingSeed.depth;

const childMint = quoteControlMint;
const child = curve({
  virtualTokenReserves: new BN("500000000000"),
  virtualQuoteReserves: new BN("250000000000"),
  realTokenReserves: new BN("400000000000"),
  realQuoteReserves: new BN("1000000000"),
  tokenTotalSupply: new BN("1000000000000000"),
  complete: false,
  depth: 1,
  quoteMint: rootMint,
  creator: zero,
});
const nested = pumpQuoteReserves(global, {
  mint: childMint,
  bondingCurve: child,
  quoteControl: null,
  chain: [{ mint: rootMint, bondingCurve: trading }],
}, null);
out.nestedVirtual = nested.virtualQuoteReserves.toString();
out.nestedDepth = nested.depth;

const tiny = curve({ virtualQuoteReserves: new BN(10), virtualTokenReserves: new BN(10), realTokenReserves: new BN(5), creator });
out.tinySell = getSellSolAmountFromTokenAmount({
  global, feeConfig, mintSupply: supply, bondingCurve: tiny, amount: new BN(1),
}).toString();

const noCreator = curve({ creator: PublicKey.default });
out.buyNoCreator = getBuyTokenAmountFromSolAmount({
  global, feeConfig, mintSupply: supply, bondingCurve: noCreator, amount: new BN("1000000000"), quoteMint: zero,
}).toString();
out.buyNewCurve = getBuyTokenAmountFromSolAmount({
  global, feeConfig, mintSupply: null, bondingCurve: null, amount: new BN("1000000000"), quoteMint: zero,
}).toString();
out.buyNewCreatorFee = getBuyTokenAmountFromSolAmount({
  global, feeConfig, mintSupply: null, bondingCurve: null, amount: new BN("1000000000"), quoteMint: zero, creatorFeeBps: new BN(250),
}).toString();
out.buyNoFeeConfig = getBuyTokenAmountFromSolAmount({
  global, feeConfig: null, mintSupply: supply, bondingCurve: solCurve, amount: new BN("1000000000"), quoteMint: zero,
}).toString();

const mayhemCurve = curve({ isMayhemMode: true });
out.buyMayhemLow = getBuyTokenAmountFromSolAmount({
  global, feeConfig, mintSupply: new BN("100000000000000"), bondingCurve: mayhemCurve, amount: new BN("1000000000"), quoteMint: zero,
}).toString();
out.buySolMayhemLow = getBuySolAmountFromTokenAmount({
  global, feeConfig, mintSupply: new BN("100000000000000"), bondingCurve: mayhemCurve, amount: new BN("1000000000000"), quoteMint: zero,
}).toString();

const zeroExotic = {
  ...feeConfig,
  exoticFlatFees: { lpFeeBps: new BN(0), protocolFeeBps: new BN(0), creatorFeeBps: new BN(0) },
};
out.buyExoticFallback = getBuyTokenAmountFromSolAmount({
  global, feeConfig: zeroExotic, mintSupply: supply, bondingCurve: exoticCurve, amount: new BN("1000000"), quoteMint: exotic,
}).toString();
const noStable = { ...feeConfig, stableFeeTiers: [] };
out.buyUsdcNoStable = getBuyTokenAmountFromSolAmount({
  global, feeConfig: noStable, mintSupply: supply, bondingCurve: usdcCurve, amount: new BN("1000000"), quoteMint: usdc,
}).toString();

const xfer = { basisPoints: 50, maximumFee: new BN("1000000000") };
out.v3WithinFee = getBuyV3QuoteAmountFromTokenAmount({
  global: v3Global, feeConfig, mintSupply: supply, bondingCurve: v3, amount: new BN(400), curveBaseTokenBalance: new BN(5000), quoteTransferFee: xfer,
}).toString();
out.v3CrossFee = getBuyV3QuoteAmountFromTokenAmount({
  global: v3Global, feeConfig, mintSupply: supply, bondingCurve: v3, amount: new BN(1500), curveBaseTokenBalance: new BN(5000), quoteTransferFee: xfer,
}).toString();
out.v3ExactFee = getBuyV3TokenAmountFromQuoteAmount({
  global: v3Global, feeConfig, mintSupply: supply, bondingCurve: v3, amount: new BN(3000), curveBaseTokenBalance: new BN(5000), quoteTransferFee: xfer,
}).toString();

out.wsolSeed = initialVirtualQuoteReservesFor(global, wsol).toString();
out.omittedExotic = initialVirtualQuoteReservesFor(global, exotic).toString();

const smallGlobal = {
  ...global,
  initialVirtualTokenReserves: new BN(10000),
  initialRealTokenReserves: new BN(6000),
  initialVirtualSolReserves: new BN(1000),
  tokenTotalSupply: new BN(10000),
  poolMigrationFee: new BN(10),
  maxCurveDepth: 3,
};
const completing = {
  virtualTokenReserves: new BN(10000),
  virtualQuoteReserves: new BN(1000),
  realTokenReserves: new BN(1000),
  realQuoteReserves: new BN(100),
  tokenTotalSupply: new BN(10000),
  complete: false,
  creator: zero,
  isMayhemMode: false,
  quoteMint: zero,
  creatorFeeBps: new BN(0),
  depth: 0,
};
out.completing = pumpQuoteReserves(smallGlobal, { mint: rootMint, bondingCurve: completing, quoteControl: null }, null).virtualQuoteReserves.toString();

const waiting = curve({ complete: true, virtualTokenReserves: new BN(5), virtualQuoteReserves: new BN(5), realTokenReserves: new BN(1), realQuoteReserves: new BN(1) });
try {
  pumpQuoteReserves(global, { mint: rootMint, bondingCurve: waiting, quoteControl: null }, null);
  out.waiting = "ok";
} catch (e) {
  out.waiting = e.name;
}

out.quoteControlPda = sdk.QUOTE_CONTROL_PDA.toBase58();
out.holderRewards = sdk.holderRewardsPda(exotic).toBase58();
console.log(JSON.stringify(out, null, 2));
