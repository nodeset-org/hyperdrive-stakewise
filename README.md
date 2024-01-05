# {::} Hyperdrive - StakeWise

A bash-based utility for operating NodeSet nodes for StakeWise vaults.

## Dependencies

Each node corresponds to a single vault and should have its own isolated environment, which must have [docker and docker compose](https://docs.docker.com/engine/install/) already installed. We recommend a fresh and updated Debian installation.

This project also assumes you have a systemd-based environment to support automatically restarting the container services on boot and a graceful shutdown process. If you use another init system, you must provide your own automation using `nodeset start` for and `nodeset shutdown` for a graceful boot and shutdown.

## Usage

### First Setup

First, clone this repository with `git clone https://github.com/nodeset-org/hyperdrive-stakewise.git`, then `cd hyperdrive-stakewise` to enter the new application directory.

To set up the environment, simply run the install script as root:
`sudo bash install-node.sh`

To see full documentation for the installation script, use the "-h" or "--help" option: `sudo bash install-node.sh --help`

Remember to forward ports on your router so you can find peers! CL clients use `9000` and EL clients use `30303` (both TCP & UDP) for peering by default. Note that the node is configured to accept all REST API requests by default, so you should NOT forward the port used for this (`5052` by default), as this poses a security risk.

After installation, the node will start syncing immediately. You will not be able to use any `nodeset` commands until after you first reload your environment (it's easiest to log out and log back in again).

### Wait for Sync

You can use `nodeset logs` to check the output logs for the node. Although testnets may be faster, it can take a long time to sync a mainnet node (12-48 hours depending on peering).

### Updating Your Node

Any time the node is started (either on OS boot or via `nodeset start`), it will automatically pull any updates for the stakewise operator binary and EL & CL clients. However, upgrading your OS and this utility must be done manually. 

To update this utility, simply delete the hyperdrive-stakewise application directory, then clone the repository again into the same location. As long as you clone it into the same location as your old application directory, everything should continue to work normally.

DO NOT delete your installation directory (default is `~/.node-data`) or you will have to reinstall and resync your node!

### Maintenance

To bring down the node (e.g. for maintenance), use `nodeset shutdown`. You can restart the node using `nodeset start` (or simply reboot the machine).

## Environment Details

At a high level, the `install-node.sh` script creates a docker compose setup in your specified data directory. It sets up the configuration files, then executes `docker compose up -d`. This includes everything you need to run a node for a NodeSet StakeWise v3 vault:

[v3-operator](https://github.com/stakewise/v3-operator): the StakeWise software used to generate keystores and validator deposit data, create/manage the node wallet, and register new validators

[ethdo](https://github.com/wealdtech/ethdo): an eth2 utility used to generate signed exit messages to send to NodeSet.

Execution layer clients: Nethermind and Geth are currently supported (more coming soon)

Consensus layer clients: Nimbus is currently supported (more coming soon)

### Custom Commands (Advanced)

If you want to run a command on any specific container, you must first source the appropriate configuration data:

E.g. `source /home/myuser/.node-data/nodeset.env`

Then, you can use docker compose to send the command:

E.g. `docker compose -f "/home/myuser/compose.yaml" run nimbus trustedNodeSync -d=/home/user/data --network=$NETWORK --trusted-node-url=https://checkpoint-sync.holesky.ethpandaops.io --backfill=false`

Keep in mind that any commands run this way will be executed inside the container, so any paths should be relative to the mounted volumes specified in the compose files located in your data directory (not to your wrapping environment).

## Future Improvements

This project needs expansion! Here are some ways you can help out:

- Adding more EL and CL client compatibility
- Adding more first-class support for referencing external EL/CL clients
- Testing and reporting bugs

**NodeSet will reward thoughtful contributions!**
