const fs = require('fs');
const Web3 = require('web3');
const config = require('./config');

// Contracts
const DummyRocketPoolToken = JSON.parse(fs.readFileSync(config.rocketPoolPath + 'build/contracts/DummyRocketPoolToken.json'));

// Mint RPL to an address
async function mintRpl() {
    try {

        // Initialise web3
        const web3 = new Web3(config.providerUrl);

        // Parse arguments
        let args = process.argv.slice(2);
        if (args.length != 2) throw new Error('Usage: node mint-rpl.js toAddress rplAmount');
        if (!web3.utils.isAddress(args[0])) throw new Error('Incorrect to address');
        if (isNaN(parseFloat(args[1]))) throw new Error('Incorrect RPL amount');
        let toAddress = args[0];
        let rplAmount = parseFloat(args[1]);

        // Get network ID & accounts
        let networkId = await web3.eth.net.getId();
        let accounts = await web3.eth.getAccounts();

        // Initialise contracts
        let rocketPoolToken = new web3.eth.Contract(DummyRocketPoolToken.abi, DummyRocketPoolToken.networks[networkId].address);

        // Mint RPL
        await rocketPoolToken.methods.mint(toAddress, web3.utils.toWei('' + rplAmount, 'ether')).send({
            from: accounts[0],
            gas: 8000000,
        });

        // Log
        console.log(rplAmount + ' RPL successfully minted to ' + toAddress + '.');

    }
    catch (e) {
        console.log(e.message);
    }
}
mintRpl();
