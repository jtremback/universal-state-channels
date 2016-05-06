/*global StateChannels web3 Uint8Array*/

const request = require('request')
import leftPad from 'left-pad'
import Pudding from 'ether-pudding'
import JSONStorage from 'node-localstorage'

const ECVerify = Pudding.whisk(abi, binary, { gasLimit: 3141592 })
const ec = ECVerify.at('address')

const StateChannels = Pudding.whisk(abi, binary, { gasLimit: 3141592 })
const channels = StateChannels.at('address')
const storage = new JSONStorage('./storage');

let commands = {
  // Propose a new channel and send to counterparty
  async proposeChannel({
    myAddress: address0,
    counterpartyAddress: address1,
    counterpartyUrl,
    channelId,
    state,
    challengePeriod
  }) {
    const fingerprint = solSha3(
      'newChannel',
      channelId,
      address0,
      address1,
      state,
      challengePeriod
    )

    storage.setItem('channels' + channelId, {
      channelId,
      address0,
      address1,
      state,
      challengePeriod,
      theirProposedUpdates: [],
      myProposedUpdates: [],
      acceptedUpdates: []
    })
    
    const signature0 = await web3.promise.eth.sign(address0, fingerprint)

    await post(counterpartyUrl + '/add_proposed_channel', {
      channelId,
      address0,
      address1,
      state,
      challengePeriod,
      signature0
    })
  },



  // Called by the counterparty over the http api, gets added to the
  // proposed channel box
  async addProposedChannel(proposal) {
    const fingerprint = solSha3(
      'newChannel',
      proposal.channelId,
      proposal.address0,
      proposal.address1,
      proposal.state,
      proposal.challengePeriod
    )

    const valid = await ec.ecverify.call(fingerprint, proposal.signature0, proposal.address0)

    if (!valid) {
      throw new Error('signature0 invalid')
    }
    
    let proposedChannels = storage.getItem('proposedChannels')
    proposedChannels.push(proposal)
    storage.setItem('proposedChannels', proposedChannels)
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
    channelId,
    sequenceNumber,
    state
  }) {
    let channel = storage.getItem('channels' + channelId)

    const fingerprint = solSha3(
      'updateState',
      channelId,
      sequenceNumber,
      state
    )
    
    const signature = await web3.promise.eth.sign(
      channel['address' + channel.me],
      fingerprint
    )

    const update = {
      channelId,
      sequenceNumber,
      state,
      ['signature' + channel.me]: signature
    }

    channel.myProposedUpdates.push(update)
    storage.setItem('channels' + channelId, channel)
    
    await post(channel.counterpartyUrl + '/add_proposed_update', update)
  },
  
  

  // Called by the counterparty over the http api, gets verified and
  // added to the proposed update box
  async addProposedUpdate(update) {
    const channel = storage.getItem('channels' + update.channelId)
    
    verifyUpdate(channel, update)
    
    channel.theirProposedUpdates.push(update)
    storage.setItem('channels' + update.channelId, channel)
  },

  

  // Sign the update and send it back to the counterparty
  async acceptUpdate(update) {
    const channel = storage.getItem('channels' + update.channelId)
    
    const fingerprint = verifyUpdate(channel, update)

    const signature = await web3.promise.eth.sign(
      channel['address' + channel.me],
      fingerprint
    )

    update['signature' + channel.me] = signature
    
    channel.acceptedUpdates.push(update)
    storage.setItem('channels' + update.channelId, channel)
    
    await post(channel.counterpartyUrl + '/add_accepted_update', update)
  },



  // Called by the counterparty over the http api, gets verified and
  // added to the accepted update box
  async addAcceptedUpdate(update) {
    const channel = storage.getItem('channels' + update.channelId)
    
    verifyUpdate(channel, update)
    
    channel.acceptedUpdates.push(update)
    storage.setItem('channels' + update.channelId, channel)
  },



  // Post an update to the blockchain
  async postUpdate(update) {
    await channels.updateState(
      update.channelId,
      update.sequenceNumber,
      update.state,
      update.signature0,
      update.signature1
    )
  },

  // Start the challenge period, putting channel closing into motion
  async startChallengePeriod(
    channelId
  ) {
    const channel = storage.getItem('channels' + channelId)
    const fingerprint = solSha3(
      'startChallengePeriod',
      channelId
    )
    
    const signature = await web3.promise.eth.sign(
      channel['address' + channel.me],
      fingerprint
    )
    
    await channels.startChallengePeriod(
      channelId,
      signature
    )
  }
}

async function verifyUpdate(channel, proposal) {
  const fingerprint = solSha3(
    'updateState',
    proposal.channelId,
    proposal.sequenceNumber,
    proposal.state
  )

  const theirAddress = channel['address' + swap[channel.me]]
  const theirSignature = proposal['signature' + swap[channel.me]]

  const valid = await ec.ecverify.call(
    fingerprint,
    theirSignature,
    theirAddress
  )

  if (!valid) {
    throw new Error('their signature invalid')
  }
  
  return fingerprint
}

const swap = [1, 0]

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