# Contract Options

Description of optional arguments with which the token contract will be launched.

## Table of Contents
- [Contract Options](#-contract-options)
	- [Table of Contents](#-table-of-contents)
	- [List of Options](#-list-of-options)
	- [Links](#-links)

## List of Options

List of prohibited functions.
```go
	&ContractOptions{
		DisabledFunctions: []string{"TxTestFunction"},
	}
```

Disallow swaps.
```go
	&ContractOptions{
		DisableSwaps: true,
	}
```

Disallow multiswaps.
```go
	&ContractOptions{
		DisableMultiSwaps: true,
	}
```

Time in seconds for the nonce window. If trying to execute a transaction in a batch that is older than the maximum nonce (at the moment) by more than NonceTTL, it will not be executed and will result in an error. It is advisable to set the parameter to 50.
If NonceTTL is set to 0, the check is done "in the old way" when adding a premadge.

```go
	&ContractOptions{
		NonceTTL: 50,
	}
```

## Links

* No
