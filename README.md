# stakewise-reference

A reference setup script for NodeSet nodes operating for StakeWise vaults.

## Dependencies

You must have docker and docker compose already installed in your target environment. We recommend a fresh and updated Debian installation. For Debian, you can install these dependencies with the following commands:

`sudo apt-get update`

`sudo apt-get install -y docker docker-compose`

This script also assumes you have a systemd-based environment to support automatically restarting the container services on boot and a graceful shutdown process.

## Usage

First, clone this repository into the target location with `git clone git@github.com:nodeset-org/stakewise-reference.git`. Each vault should have its own isolated environment

Next, check the settings in the vault env file (`holesky.env` or `gravita.env`) to ensure they match your needs.

To set up the environment, simply run `init-node.sh`.

Usage: `init-node.sh VAULT [--reset|-r]`

Supported vaults: `holesky`, `gravita`

Example: `sh init-node.sh holesky` will initialize a node for [NodeSet's test vault on Holesky](https://app.stakewise.io/vault/0x01b353abc66a65c4c0ac9c2ecf82e693ce0303bc).

If something goes wrong, you can use the `-r` flag to reset the configuration completely before initializing as usual, deleting all the chain data and client caches. On Holesky, resyncing is quick, but _DO NOT DO THIS ON MAINNET_ if you have any active validators!

Remember to forward your ports so you can find peers! Nimbus uses `9000` and Geth uses `30303` (both TCP & UDP).

Once you run the script, logs will be shown. You may exit this view safely with `ctrl+c` and everything will continue running. To see the logs again, use `docker compose logs -f`. To bring down the containers (e.g. for maintenance), use `docker compose down`. You can safely restart everything with the same command (`sh init-node.sh VAULT`) or simply use `docker compose up -d`.

## Migration

This script will only set up a node environment for you. It will not help you migrate your setup to another machine. You must do this manually using the your mnemonic (and ideally your keystore backups). If you need help to migrate your setup to another machine, please visit the [#tech-chat channel of our Discord](https://discord.gg/fDK3TzctPD).

## Environment Details

At a high level, the `init-node.sh` script wraps a docker compose setup. It sets up the environment, then executes `docker compose up -d`. This includes everything you need to run a node for a NodeSet StakeWise v3 vault:

[v3-operator](https://github.com/stakewise/v3-operator): the StakeWise software used to generate keystores and validator deposit data

[Nimbus](https://nimbus.guide/): a lightweight Ethereum validator client

[Geth](https://geth.ethereum.org/docs): an Ethereum execution client

[ethdo](https://github.com/wealdtech/ethdo): an eth2 utility developed by Jim McDonald used to generate exit messages

## Future Improvements

This project needs expansion! Here are some ways you can help out:

- Adding more EL and CL client compatibility
- Adding the ability to reference an external EL client instead of using a local one
- Testing and reporting bugs

**NodeSet will reward your thoughtful contributions!**
