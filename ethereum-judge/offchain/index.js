require('babel-register')
require('babel-polyfill')

const Web3 = require('web3')
const web3 = new Web3()
const Pudding = require('ether-pudding')
Pudding.setWeb3(web3)

// console.log(Web3)
web3.setProvider(new Web3.providers.HttpProvider('http://localhost:8545'));
const ECVerify = require('../environments/test/contracts/ECVerify.sol.js')
ECVerify.load(Pudding)
const ec = ECVerify.at('address')

const StateChannels = require('../environments/test/contracts/StateChannels.sol.js')
StateChannels.load(Pudding)
const channels = StateChannels.at('address')

const {JSONStorage} = require('node-localstorage')
const storage = new JSONStorage('../data/storage')

const globals = {
  storage,
  ec,
  channels,
  web3
}

const caller = require('./servers/caller.js').default(globals)
const peer = require('./servers/peer.js').default(globals)

caller.listen(3020, () => console.log('caller api listening on 3020'))
peer.listen(4020, () => console.log('peer api listening on 4020'))