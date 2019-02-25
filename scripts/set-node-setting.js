const fs = require('fs');
const Web3 = require('web3');
const config = require('./config');

// Contracts
const RocketNodeSettings = JSON.parse(fs.readFileSync(config.rocketPoolPath + 'build/contracts/RocketNodeSettings.json'));

// Set node setting
async function setNodeSetting() {
    try {

        // Initialise web3
        const web3 = new Web3(config.providerUrl);

        // Parse arguments
        let args = process.argv.slice(2);
        if (args.length < 2) throw new Error('Usage: node set-node-setting.js setting values...');
        let setting = args[0];
        let values = args.slice(1);
        values.forEach((v, vi) => {
            if (v == 'true') values[vi] = true;
            else if (v == 'false') values[vi] = false;
        });

        // Get network ID & accounts
        let networkId = await web3.eth.net.getId();
        let accounts = await web3.eth.getAccounts();

        // Initialise contracts
        let rocketNodeSettings = new web3.eth.Contract(RocketNodeSettings.abi, RocketNodeSettings.networks[networkId].address);

        // Set node setting
        await rocketNodeSettings.methods[setting](...values).send({
            from: accounts[0],
            gas: 8000000,
        });

        // Log
        console.log('RocketNodeSettings setting ' + setting + ' successfully updated.');

    }
    catch (e) {
        console.log(e.message);
    }
}
setNodeSetting();
