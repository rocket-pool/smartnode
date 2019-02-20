const fs = require('fs');
const Web3 = require('web3');

// Rocketpool project path
const rpPath = __dirname + '/../../../../../../rocketpool/rocketpool/';

// Contracts
const RocketGroupAccessorContract = JSON.parse(fs.readFileSync(rpPath + 'build/contracts/RocketGroupAccessorContract.json'));

// Make user deposit
async function userDeposit() {
    try {

        // Initialise web3
        const web3 = new Web3('http://localhost:8545');

        // Parse arguments
        let args = process.argv.slice(2);
        if (args.length != 3) throw new Error('Usage: node user-deposit.js depositorAddress durationID etherAmount');
        if (!web3.utils.isAddress(args[0])) throw new Error('Incorrect depositor address');
        if (isNaN(parseFloat(args[2]))) throw new Error('Incorrect ether amount');
        let depositorAddress = args[0];
        let durationID = args[1];
        let etherAmount = parseFloat(args[2]);

        // Get accounts
        let accounts = await web3.eth.getAccounts();

        // Initialise contracts
        let depositorContract = new web3.eth.Contract(RocketGroupAccessorContract.abi, depositorAddress);

        // Deposit
        await depositorContract.methods.deposit(durationID).send({
            from: accounts[0],
            value: web3.utils.toWei('' + etherAmount, 'ether'),
            gas: 8000000,
        });

        // Log
        console.log(etherAmount + ' ether successfully deposited to ' + depositorAddress + ', staking for ' + durationID + '.');

    }
    catch (e) {
        console.log(e.message);
    }
}
userDeposit();
