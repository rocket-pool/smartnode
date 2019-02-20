const fs = require('fs');
const Web3 = require('web3');
const config = require('./config');

// Contracts
const RocketGroupAPI = JSON.parse(fs.readFileSync(config.rocketPoolPath + 'build/contracts/RocketGroupAPI.json'));
const RocketGroupContract = JSON.parse(fs.readFileSync(config.rocketPoolPath + 'build/contracts/RocketGroupContract.json'));
const RocketGroupSettings = JSON.parse(fs.readFileSync(config.rocketPoolPath + 'build/contracts/RocketGroupSettings.json'));

// Create group & accessor
async function createGroupAccessor() {
    try {

        // Initialise web3
        const web3 = new Web3(config.providerUrl);

        // Parse arguments
        let args = process.argv.slice(2);
        if (args.length != 1) throw new Error('Usage: node create-group-accessor.js groupName');
        let groupName = args[0];

        // Get network ID & accounts
        let networkId = await web3.eth.net.getId();
        let accounts = await web3.eth.getAccounts();

        // Initialise contracts
        let rocketGroupAPI = new web3.eth.Contract(RocketGroupAPI.abi, RocketGroupAPI.networks[networkId].address);
        let rocketGroupSettings = new web3.eth.Contract(RocketGroupSettings.abi, RocketGroupSettings.networks[networkId].address);

        // Get new group fee
        let newGroupFee = await rocketGroupSettings.methods.getNewFee().call();

        // Create group & get address
        let groupAddResult = await rocketGroupAPI.methods.add(groupName, web3.utils.toWei('0', 'ether')).send({
            from: accounts[0],
            gas: 8000000,
            value: newGroupFee,
        });
        let groupContractAddress = groupAddResult.events.GroupAdd.returnValues.ID;

        // Initialise group contract
        let groupContract = new web3.eth.Contract(RocketGroupContract.abi, groupContractAddress);

        // Create accessor & get address
        let accessorCreateResult = await rocketGroupAPI.methods.createDefaultAccessor(groupContractAddress).send({
            from: accounts[0],
            gas: 8000000,
        });
        let accessorContractAddress = accessorCreateResult.events.GroupCreateDefaultAccessor.returnValues.accessorAddress;

        // Add accessor to group
        await groupContract.methods.addDepositor(accessorContractAddress).send({
            from: accounts[0],
            gas: 8000000,
        });
        await groupContract.methods.addWithdrawer(accessorContractAddress).send({
            from: accounts[0],
            gas: 8000000,
        });

        // Log
        console.log('Successfully created group at ' + groupContractAddress + ' and accessor at ' + accessorContractAddress + '.');

    }
    catch (e) {
        console.log(e.message);
    }
}
createGroupAccessor();
