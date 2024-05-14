## x/vesting

### Overview

This page presents comprehensive details regarding the Vesting Module offered by Evmos. The functionality of this module empowers users to transform their accounts into clawback vesting accounts, allowing the reception of tokens from a designated funder. Additionally, it provides the funder with the capability to stipulate specific vesting and locking schedules, and to clawback unvested tokens.
This module is a copy of the original [x/vesting](https://github.com/evmos/evmos/tree/main/x/vesting) module and has been adapted to not rely on Evmos's custom [EthAccount](https://github.com/evmos/evmos/blob/6fda63866715b248bbd03c13461163a58de48c09/proto/ethermint/types/v1/account.proto#L14).
This allows the module to be compatible with other non-EVM Cosmos SDK chains.

Vesting accounts, defined in the `x/vesting` module, are a new type implementing the Cosmos SDK [`VestingAccount` interface](https://docs.cosmos.network/main/modules/auth/vesting#vesting-account-types).
This new type allows users to convert their account into a clawback vesting account.
This account type can receive tokens from a particular address (_funder_),
which are then released according to the defined lockup and vesting schedules.
The funder of the account can claw back unvested tokens at any time.
The clawback can also be enforced through a governance proposal for any account that opts in to this when converting to a vesting account.

For a complete description of the vesting functionalities,
please visit the corresponding page in the [docs](https://docs.evmos.org/protocol/modules/vesting#state-transitions).

## Implementing the Module

The shown example demonstrates how to include the [evmos/vesting](https://github.com/evmos/vesting) module in the Stride chain.
Because each chain may have specific characteristics,
the following steps might not be universally applicable,
and other modifications may be necessary before the module can be used correctly.

### `app.go` Wiring

- Register the [governance clawback](https://github.com/Vvaradinov/stride/blob/ec1df0f0cfc377eda48c84263b3925ca91b4731f/app/app.go#L181) proposal handler in the `govProposalHandlers`.
- Include the vesting module’s `[AppModuleBasic].(https://github.com/Vvaradinov/stride/blob/ec1df0f0cfc377eda48c84263b3925ca91b4731f/app/app.go#L227)` into the chain’s `ModuleBasics`.
- Include the `VestingKeeper` in the [`StrideApp`].(https://github.com/Vvaradinov/stride/blob/ec1df0f0cfc377eda48c84263b3925ca91b4731f/app/app.go#L282) struct.
- Include the vesting module’s `StoreKey` into the the `KVStoreKey` map.
- Initialize the `[VestingKeeper].(https://github.com/Vvaradinov/stride/blob/ec1df0f0cfc377eda48c84263b3925ca91b4731f/app/app.go#L618)`.
- Add the new `AppModule`.

```go
evmosvesting.NewAppModule(app.VestingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
```

- Include the vesting module in the `BeginBlocker` and `EndBlocker` stack.

### Enable `from` flag for the CLI

- Enable the `from` [flag in the CLI](https://github.com/Stride-Labs/stride/blob/405fb9c961e537619092dc51cc70107aedf03ba4/cmd/strided/root.go#L268) to be able to specify the sender key.

### `AnteHandler` for Locked and Unvested tokens

- The Vesting `AnteHandler` is included in [the repo here](https://github.com/evmos/vesting/blob/main/ante/vesting.go). In order to use it, Stride should add the `StakingKeeper` definition to the `HandlerOptions` struct:

```go
// HandlerOptions are the options required for constructing a default SDK AnteHandler.
type HandlerOptions struct {
    AccountKeeper          AccountKeeper
    BankKeeper             types.BankKeeper
    ExtensionOptionChecker ExtensionOptionChecker
    FeegrantKeeper         FeegrantKeeper
    SignModeHandler        authsigning.SignModeHandler
    SigGasConsumer         func(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error
    TxFeeChecker           TxFeeChecker
}
```

- The `AnteHandler` requires a binary codec to be included the `ante_handlers.go` `HandlerOptions` struct:

```go
// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
    ante.HandlerOptions
    Cdc            codec.BinaryCodec // Add this here
    IBCKeeper      *ibckeeper.Keeper
    ConsumerKeeper ccvconsumerkeeper.Keeper
}
```

- Lastly, all expected keepers interfaces should have the required functions, like the `BankKeeper` needs the `GetBalance` function.

## Licensing

### SPDX Identifier

The following header including a license identifier in SPDX short form has been added to all ENCL-1.0 files:

```go
// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
```

Exempted files contain the following SPDX ID:

```go
// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:LGPL-3.0-only
```

### License FAQ

Find below an overview of the Permissions and Limitations of the Evmos Non-Commercial License 1.0.
For more information, check out the full ENCL-1.0 FAQ [here](/LICENSE_FAQ.md).

| Permissions                                                                                                                                                                  | Prohibited                                                                 |
| ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| - Private Use, including distribution and modification<br />- Commercial use on designated blockchains<br />- Commercial use with Evmos permit (to be separately negotiated) | - Commercial use, other than on designated blockchains, without Evmos permit |
