'use strict';

module.exports = {
	minVersion: "5.1.0t",
	currentVersion: "5.2.0t",
	activeDelegates: 101,
	addressLength: 208,
	blockHeaderLength: 248,
	blockTime: 27000,
	blockReceiptTimeOut: blockTime/1000*12, // 12 blocks
	confirmationLength: 77,
	epochTime: new Date(Date.UTC(2016, 4, 24, 17, 0, 0, 0)),
	fees: {
		send: 10000000,
		vote: 100000000,
		secondsignature: 500000000,
		delegate: 6000000000,
		multisignature: 500000000,
		dapp: 2500000000
	},
	feeStart: 1,
	feeStartVolume: 10000 * 100000000,
	fixedPoint: Math.pow(10, 8),
	maxAddressesLength: 208 * 128,
	maxAmount: 100000000,
	maxConfirmations: 77 * 100,
	maxPayloadLength: 1024 * 1024,
	maxPeers: 100,
	maxRequests: 10000 * 12,
	maxSharedTxs: 100,
	maxSignaturesLength: 196 * 256,
	maxTxsPerBlock: 25,
	maxTxsPerQueue: 5000,
	minBroadhashConsensus: 51,
	nethashes: [
		// Mainnet
		'7337a324ef27e1e234d1e9018cacff7d4f299a09c2df9be460543b8f7ef652f1',
		// Testnet
		'cba57b868c8571599ad594c6607a77cad60cf0372ecde803004d87e679117c12'
	],
	numberLength: 100000000,
	requestLength: 104,
	rewards: {
		milestones: [
            100000000,  // Initial reward
            70000000,  // Milestone 1
            50000000,  // Milestone 2
            30000000,  // Milestone 3
            20000000   // Milestone 4
		],
		offset: 10,   // Start rewards at block (n)
		distance: 3000000, // Distance between each milestone
	},
	signatureLength: 196,
	totalAmount: 1009000000000000,
	unconfirmedTransactionTimeOut: 10800 // 1080 blocks
};
