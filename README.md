# stakewise-testing

A script which sets up a node for StakeWise vaults.

## Dependencies

You must have docker and docker compose already installed in your target environment. We recommend a fresh and updated Debian installation.

## Usage

First, clone this repository into the target location with `git clone git@github.com:nodeset-org/stakewise-reference.git`.

To set up the environment, simply run `init-node.sh`.
Usage: `init-node.sh VAULT [--reset|-r]`
Supported vaults: `holesky`, `gravita`
Example: `sh init-node.sh holesky` will initialize a node for [NodeSet's test vault on Holesky](https://app.stakewise.io/vault/0x01b353abc66a65c4c0ac9c2ecf82e693ce0303bc).

If something goes wrong, you can use the -r flag, which deletes all the chain data and client caches. On Holesky, resyncing is quick, but don't do this on mainnet if you have any active validators!

Remember to forward your ports so you can find peers! Nimbus uses 9000 and Geth uses 30303.

Once you run the script, logs will be shown. You may exit this view safely with `ctrl+c` and everything will continue running. To see the logs again, use `docker compose logs -f`. To bring down the containers (e.g. for maintenance), use `docker compose down`. You can safely restart everything with the same command

## Environment

At a high level, the `init-node.sh` script wraps a docker compose file, `compose.yaml`. It sets up the environment, then executes `docker compose up -d`. To reset

## Internal Software

[`operator`](https://github.com/stakewise/v3-operator): the StakeWise v3-operator software used to generate keystores and validator deposit data

[Nimbus](https://nimbus.guide/): a lightweight Ethereum Validator Client

[`ethdo`](https://github.com/wealdtech/ethdo): an eth2 utility developed by Jim McDonald used to generate
