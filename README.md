# stakewise-testing

A script which sets up a node for StakeWise vaults.

## Dependencies

You must have docker and docker compose already installed in your target environment. We recommend a fresh and updated Debian installation.

## Usage

To set up the environment, simply run `initialize.sh`.
Usage: `initialize.sh VAULT [--reset|-r]`
Supported vaults: `holesky`, `gravita`
Example: `sh initialize.sh holesky` will initialize a node for [NodeSet's test vault on Holesky](https://app.stakewise.io/vault/0x01b353abc66a65c4c0ac9c2ecf82e693ce0303bc).

[`operator`](https://github.com/stakewise/v3-operator): the StakeWise v3-operator software used to generate keystores and validator deposit data

[Nimbus](https://nimbus.guide/): a lightweight Ethereum Validator Client

[`ethdo`](https://github.com/wealdtech/ethdo): an eth2 utility developed by Jim McDonald used to generate

NOTE: Remember to forward your ports so you can find peers! Nimbus uses 9000 and Geth uses 30303.
