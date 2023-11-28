# stakewise-testing

A dockerfile for testing the NodeSet StakeWise operator experience

Use `sudo docker build . -t swtest` to build the container and `docker run -it swtest` to enter it.

From inside the container's shell, you have access to:

`operator`: the StakeWise operator software used to generate keystores and validator deposit data

[Nimbus](https://nimbus.guide/): a lightweight Ethereum Validator Client

`deposit`: a custom fork of the [staking-deposit-cli tool](https://github.com/nodeset-org/staking-deposit-cli) which can be used to generate signed voluntary exit messages from keystores or a mnemonic

NOTES:
Store chain data outside of the container and use that to start nimbus/geth instead?
