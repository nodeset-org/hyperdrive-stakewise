/*** Dependencies ********************/

const hre = require('hardhat');
const pako = require('pako');
const fs = require('fs');
const Web3 = require('web3');


/*** Utility Methods *****************/


// Compress / decompress ABIs
function compressABI(abi) {
    return Buffer.from(pako.deflate(JSON.stringify(abi))).toString('base64');
}
function decompressABI(abi) {
    return JSON.parse(pako.inflate(Buffer.from(abi, 'base64'), {to: 'string'}));
}

// Load ABI files and parse
function loadABI(abiFilePath) {
    return JSON.parse(fs.readFileSync(abiFilePath));
}


/*** Contracts ***********************/

// Multicall
const multicall = artifacts.require('Multicall2.sol');

// Balance Batcher
const balanceBatcher = artifacts.require('BalanceChecker.sol');


/*** Deployment **********************/

// Deploy contracts needed to test the StakeWise module
export async function deployContracts() {
    // Set our web3 provider
    const network = hre.network;
    let $web3 = new Web3(network.provider);

    // Accounts
    let accounts = await $web3.eth.getAccounts(function(error, result) {
        if(error != null) {
            console.log(error);
            console.log("Error retrieving accounts.'");
        }
        return result;
    });

    console.log(`Using network: ${network.name}`);
    console.log(`Deploying from: ${accounts[1]}`)
    console.log('\n');

    // Deploy the Beacon deposit contract
    const casperDepositABI = loadABI('./contracts/Deposit.abi');
    const casperDeposit = new $web3.eth.Contract(casperDepositABI, null, {
        from: accounts[0],
        gasPrice: '20000000000' // 20 gwei
    });
    const casperDepositContract = await casperDeposit.deploy(
        {
            data: fs.readFileSync('./contracts/Deposit.bin').toString()
        }).send({
        from: accounts[0],
        gas: 8000000,
        gasPrice: '20000000000'
    });

    // Set the Casper deposit address
    let casperDepositAddress = casperDepositContract._address;
    console.log('   Beacon Deposit Address');
    console.log('     ' + casperDepositAddress);

    // Deploy Multicall
    var multicallInstance = await multicall.new({from: accounts[1]});
    multicall.setAsDeployed(multicallInstance);
    const multicallAddress = (await multicall.deployed()).address;
    console.log('   Multicall Address');
    console.log('     ' + multicallAddress);
    
    // Deploy Balance Batcher
    var balanceBatcherInstance = await balanceBatcher.new({from: accounts[1]});
    balanceBatcher.setAsDeployed(balanceBatcherInstance);
    const balanceBatcherAddress = (await balanceBatcher.deployed()).address;
    console.log('   Balance Batcher Address');
    console.log('     ' + balanceBatcherAddress);

    // Log it
    console.log('\n');
    console.log('  Done!');
    console.log('\n');
};

// Run it
deployContracts().then(function() {
    process.exit(0);
});