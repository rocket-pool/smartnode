const fs = require('fs');
const Web3 = require('web3');
const config = require('./config');

// Contracts
const RocketAdmin = JSON.parse(fs.readFileSync(config.rocketPoolPath + 'build/contracts/RocketAdmin.json'));

// Set node trusted status
async function setNodeTrusted() {
    try {

        // Initialise web3
        const web3 = new Web3(config.providerUrl);

        // Parse arguments
        let args = process.argv.slice(2);
        if (args.length != 2) throw new Error('Usage: node set-node-trusted.js nodeAccountAddress trusted');
        if (!web3.utils.isAddress(args[0])) throw new Error('Incorrect node account address');
        if (!args[1].match(/^(true|false)$/)) throw new Error('Incorrect trusted value');
        let nodeAccountAddress = args[0];
        let trusted = args[1];
        if (trusted == 'true') trusted = true;
        else if (trusted == 'false') trusted = false;

        // Get network ID & accounts
        let networkId = await web3.eth.net.getId();
        let accounts = await web3.eth.getAccounts();

        // Initialise contracts
        let rocketAdmin = new web3.eth.Contract(RocketAdmin.abi, RocketAdmin.networks[networkId].address);

        // Set node trusted status
        await rocketAdmin.methods.setNodeTrusted(nodeAccountAddress, trusted).send({
            from: accounts[0],
            gas: 8000000,
        });

        // Log
        console.log('Node account ' + nodeAccountAddress + ' trusted status successfully set to ' + trusted + '.');

    }
    catch (e) {
        console.log(e.message);
    }
}
setNodeTrusted();
