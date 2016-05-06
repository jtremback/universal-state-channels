/*global StateChannels web3 Uint8Array*/

const request = require('request')
const minimist = require('minimist')
import leftPad from 'left-pad'
import BigNumber from 'bignumber.js'
import Pudding from 'ether-pudding'

const ECVerify = Pudding.whisk(abi, binary, { gasLimit: 3141592 })
const ec = ECVerify.at('address')

const StateChannels = Pudding.whisk(abi, binary, { gasLimit: 3141592 })
const channels = StateChannels.at('address')

function padNumber(num) {
  return leftPad((num).toString(16), 64, 0)
}

let commands = {
  // View all channels
  viewChannels({ accountPubkey }) {
    get('http://localhost:4545/channels/' + accountPubkey, function (err, res, channels) {
      if (err) { console.log(err) }
      console.log(channels)
    })
  },
  // View all counterparties
  viewCounterparties() {
    get('http://localhost:4545/view_counterparties', function (err, res, counterparties) {
      if (err) { console.log(err) }
      console.log(counterparties)
    })
  },



  // Propose a new channel and send to counterparty
  async proposeChannel({
    myAddress,
    counterpartyAddress,
    counterpartyUrl,
    channelId,
    state,
    challengePeriod
  }) {
    const fingerprint = solSha3(
      'newChannel',
      channelId,
      myAddress,
      counterpartyAddress,
      state,
      challengePeriod
    )

    const signature0 = await web3.promise.eth.sign(myAddress, fingerprint)

    await post(counterpartyUrl + '/propose_channel', {
      channelId,
      address0: myAddress,
      address1: counterpartyAddress,
      state,
      challengePeriod,
      signature0
    })
  },



  // Sign the opening tx and post it to the blockchain to open the channel
  async acceptChannel({
    channelId,
    address0,
    address1,
    state,
    challengePeriod,
    signature0
  }) {
    const fingerprint = solSha3(
      'newChannel',
      channelId,
      address0,
      address1,
      state,
      challengePeriod
    )

    const valid = await ec.ecverify.call(fingerprint, signature0, address0)

    if (!valid) {
      throw new Error('signature0 invalid')
    }

    const signature1 = await web3.promise.eth.sign(address1, fingerprint)

    await channels.newChannel(
      channelId,
      address0,
      address1,
      state,
      challengePeriod,
      signature0,
      signature1
    )
  },



  // Propose an update to a channel, sign, and send to counterparty
  async proposeUpdate({
    myAddress,
    counterpartyAddress,
    counterpartyUrl,
    channelId,
    state,
    sequenceNumber
  }) {
    const fingerprint = solSha3(
      'updateState',
      channelId,
      sequenceNumber,
      state
    )

    const signature0 = await web3.promise.eth.sign(myAddress, fingerprint)

    await post(counterpartyUrl + '/propose_update', {
      address0: myAddress,
      address1: counterpartyAddress,
      channelId,
      sequenceNumber,
      state,
      signature0
    })
  },



  // Sign the opening tx and send it back to the counterparty
  async acceptUpdate({
    myAddress,
    counterpartyAddress,
    counterpartyUrl,
    channelId,
    sequenceNumber,
    state,
    signature0
  }) {
    const fingerprint = solSha3(
      'updateState',
      channelId,
      sequenceNumber,
      state
    )

    const valid = await ec.ecverify.call(fingerprint, signature0, counterpartyAddress)

    if (!valid) {
      throw new Error('signature0 invalid')
    }

    const signature1 = await web3.promise.eth.sign(myAddress, fingerprint)

    await post(counterpartyUrl + '/accept_update', {
      channelId,
      sequenceNumber,
      state,
      signature0,
      signature1
    })
  },



  // Post an update to the blockchain
  async postUpdate({
    address0,
    address1,
    channelId,
    sequenceNumber,
    state,
    signature0
  }) {
    const fingerprint = solSha3(
      'updateState',
      channelId,
      sequenceNumber,
      state
    )

    const valid = await ec.ecverify.call(fingerprint, signature0, address0)

    if (!valid) {
      throw new Error('signature0 invalid')
    }

    const signature1 = await web3.promise.eth.sign(address1, fingerprint)

    await channels.updateState(
      channelId,
      web3.toHex(sequenceNumber),
      state,
      signature0,
      signature1
    )
  },

  // Start the challenge period, putting channel closing into motion
  async startChallengePeriod({
    channelId,
    myAddress
  }) {
    const fingerprint = solSha3(
      'startChallengePeriod',
      channelId
    )
    
    const signature = await web3.promise.eth.sign(myAddress, fingerprint)
    
    await channels.startChallengePeriod(
      '0x' + channelId,
      myAddress,
      signature
    )
  }
}

function get(url, callback) {
  request.get({
    url,
    json: true,
  }, callback)
}

function post(url, body, callback) {
  request.post({
    url,
    body,
    json: true,
  }, callback)
}

let argv = require('minimist')(process.argv.slice(2))
if (commands[argv._[0]]) {
  commands[argv._[0]](argv)
} else if (argv._[0]) {
  console.log(argv._[0] + ' is not a command')
} else {
  console.log('please enter a command')
}


function solSha3(...args) {
  args = args.map(arg => {
    if (typeof arg === 'string') {
      if (arg.substring(0, 2) === '0x') {
        return arg.slice(2)
      } else {
        return web3.toHex(arg).slice(2)
      }
    }

    if (typeof arg === 'number') {
      return leftPad((arg).toString(16), 64, 0)
    }
  })

  args = args.join('')

  return '0x' + web3.sha3(args, { encoding: 'hex' })
}