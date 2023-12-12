# stakewise-reference

A reference setup script for NodeSet nodes operating for StakeWise vaults.

## Dependencies

You must have docker and docker compose already installed in your target environment. We recommend a fresh and updated Debian installation. For Debian, you can install these dependencies with the following commands:

`sudo apt-get update`

`sudo apt-get install -y docker docker-compose`

This script also assumes you have a systemd-based environment to support automatically restarting the container services on boot and a graceful shutdown process.

## Usage

### First Setup

First, clone this repository into the target location with `git clone git@github.com:nodeset-org/stakewise-reference.git`. Each vault should have its own isolated environment

To set up the environment, simply run the install script as root:
`sudo bash install-node.sh`

Remember to forward your ports so you can find peers! Nimbus uses `9000` and Geth uses `30303` (both TCP & UDP). Note: for simplicity, the included EC configurations are open to all external http requests. You should NOT forward the port for the HTTP endpoints without first limiting the configuration further, otherwise your node may be vulnerable to a DDOS attack.

Once you run the script, logs will be shown. You may exit this view safely with `ctrl+c` and everything will continue running. To see the logs again, use `docker compose logs -f`.

### Reset

If something goes wrong, you can use the `-r` flag to reset the configuration completely before initializing as usual, deleting all the chain data and client caches. On Holesky, resyncing is quick, but _DO NOT DO THIS ON MAINNET_ if you have any active validators!

### Graceful Shutdown

To bring down the node (e.g. for maintenance), use the `-s` or `--shutdown` flag. 

Example: `sh init-node.sh --shutdown VAULT`

You can safely restart everything with the same command, `sh init-node.sh VAULT`.

### Migration

Instead of setting up a new configuration from scratch, this script can import an existing setup using the `-m` or `--mnemonic` option to provide an existing mnemonic.

For example:
`sudo sh init-node.sh -m "correct horse battery staple..." holesky`

### Custom Commands

If you want to run a command on any specific container, you can do so like this:
`docker compose run CONTAINERNAME COMMAND`
E.g., perform a trusted node sync in Nimbus: `docker compose run nimbus trustedNodeSync -d=/home/user/data --network=$NETWORK --trusted-node-url=https://checkpoint-sync.holesky.ethpandaops.io --backfill=false`

## Environment Details

At a high level, the `init-node.sh` script wraps a docker compose setup. It sets up the environment, then executes `docker compose up -d`. This includes everything you need to run a node for a NodeSet StakeWise v3 vault:

[v3-operator](https://github.com/stakewise/v3-operator): the StakeWise software used to generate keystores and validator deposit data, create/manage the node wallet, and register new validators

[Nimbus](https://nimbus.guide/): a lightweight Ethereum validator client.

[Geth](https://geth.ethereum.org/docs): an Ethereum execution client.

[ethdo](https://github.com/wealdtech/ethdo): an eth2 utility developed by Jim McDonald. Used to generate backup exit messages to send to NodeSet.

## Future Improvements

This project needs expansion! Here are some ways you can help out:

- Adding more EL and CL client compatibility
- Adding the ability to reference an external EL client instead of using a local one
- Testing and reporting bugs

**NodeSet will reward your thoughtful contributions!**





## TEMP

`sudo bash install-node.sh` -- installs the node, can use `-r` to remove the previous configuration, use `-d` to pass your own data directory
`nodeset remove` -- removes the installation
`nodeset start` -- starts the node
`nodeset logs` -- self-explanatory
`nodeset shutdown` -- shuts down all the containers