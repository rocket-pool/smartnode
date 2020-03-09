const fs = require('fs');
const Web3 = require('web3');
const config = require('./config');

// Contracts
const RocketAdmin = JSON.parse(fs.readFileSync(config.rocketPoolPath + 'build/contracts/RocketAdmin.json'));
const RocketNodeAPI = JSON.parse(fs.readFileSync(config.rocketPoolPath + 'build/contracts/RocketNodeAPI.json'));
const RocketNodeWatchtower = JSON.parse(fs.readFileSync(config.rocketPoolPath + 'build/contracts/RocketNodeWatchtower.json'));


// Get the Rocket Pool withdrawal pubkey
function getWithdrawalPubkey() {
    return Buffer.from('0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef', 'hex');
}


// Get the Rocket Pool withdrawal credentials
function getWithdrawalCredentials(web3) {
    return Buffer.concat([
        Buffer.from('00', 'hex'), // BLS_WITHDRAWAL_PREFIX_BYTE
        Buffer.from(web3.utils.sha3(getWithdrawalPubkey()).substr(2), 'hex').slice(1) // Last 31 bytes of withdrawal pubkey hash
    ], 32);
}


// Set withdrawal credentials
async function setWithdrawalCredentials() {
    try {

        // Initialise web3
        const web3 = new Web3(config.providerUrl);

        // Parse arguments
        let args = process.argv.slice(2);
        if (args.length != 0) throw new Error('Usage: node set-withdrawal-credentials.js');

        // Get network ID & accounts
        let networkId = await web3.eth.net.getId();
        let accounts = await web3.eth.getAccounts();
        let owner = accounts[0];
        let nodeOperator = accounts[9];

        // Initialise contracts
        let rocketAdmin = new web3.eth.Contract(RocketAdmin.abi, RocketAdmin.networks[networkId].address);
        let rocketNodeAPI = new web3.eth.Contract(RocketNodeAPI.abi, RocketNodeAPI.networks[networkId].address);
        let rocketNodeWatchtower = new web3.eth.Contract(RocketNodeWatchtower.abi, RocketNodeWatchtower.networks[networkId].address);

        // Register node
        await rocketNodeAPI.methods.add('Australia/Brisbane').send({
            from: nodeOperator,
            gas: 8000000,
        });

        // Set node trusted status
        await rocketAdmin.methods.setNodeTrusted(nodeOperator, true).send({
            from: owner,
            gas: 8000000,
        });

        // Set withdrawal credentials
        await rocketNodeWatchtower.methods.updateWithdrawalKey(getWithdrawalPubkey(), getWithdrawalCredentials(web3)).send({
            from: nodeOperator,
            gas: 8000000,
        });

        // Log
        console.log('Rocket Pool withdrawal credentials set successfully.');

    }
    catch (e) {
        console.log(e.message);
    }
}
setWithdrawalCredentials();
