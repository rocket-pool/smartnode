const Web3 = require('web3');

// Send ether to an account
async function sendEther() {
    try {

        // Initialise web3
        const web3 = new Web3('http://localhost:8545');

        // Parse arguments
        let args = process.argv.slice(2);
        if (args.length != 2) throw new Error('Usage: node send-ether.js toAddress etherAmount');
        if (!web3.utils.isAddress(args[0])) throw new Error('Incorrect to address');
        if (isNaN(parseFloat(args[1]))) throw new Error('Incorrect ether amount');
        let toAddress = args[0];
        let etherAmount = parseFloat(args[1]);

        // Get accounts
        let accounts = await web3.eth.getAccounts();

        // Send ether
        await web3.eth.sendTransaction({
            from: accounts[0],
            to: toAddress,
            value: web3.utils.toWei('' + etherAmount, 'ether'),
            gas: 8000000,
        });

        // Log
        console.log(etherAmount + ' ether successfully sent to ' + toAddress + '.');

    }
    catch (e) {
        console.log(e.message);
    }
}
sendEther();
