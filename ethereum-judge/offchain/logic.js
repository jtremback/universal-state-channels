/*global Uint8Array*/

import request from 'request'
import leftPad from 'left-pad'

export class Logic {
  constructor({
    storage,
    ec,
    channels,
    web3
  }) {
    this.storage = storage
    this.ec = ec
    this.channels = channels
    this.web3 = web3
  }
  
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

    this.storage.setItem('channels' + channelId, {
      channelId,
      address0,
      address1,
      state,
      challengePeriod,
      theirProposedUpdates: [],
      myProposedUpdates: [],
      acceptedUpdates: []
    })
    
    const signature0 = await this.web3.promise.eth.sign(address0, fingerprint)

    await post(counterpartyUrl + '/add_proposed_channel', {
      channelId,
      address0,
      address1,
      state,
      challengePeriod,
      signature0
    })
  }



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

    const valid = await this.ec.ecverify.call(fingerprint, proposal.signature0, proposal.address0)

    if (!valid) {
      throw new Error('signature0 invalid')
    }
    
    let proposedChannels = this.storage.getItem('proposedChannels')
    proposedChannels.push(proposal)
    this.storage.setItem('proposedChannels', proposedChannels)
  }



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

    const valid = await this.ec.ecverify.call(fingerprint, signature0, address0)

    if (!valid) {
      throw new Error('signature0 invalid')
    }

    const signature1 = await this.web3.promise.eth.sign(address1, fingerprint)

    await this.channels.newChannel(
      channelId,
      address0,
      address1,
      state,
      challengePeriod,
      signature0,
      signature1
    )
  }



  // Propose an update to a channel, sign, and send to counterparty
  async proposeUpdate({
    channelId,
    sequenceNumber,
    state
  }) {
    let channel = this.storage.getItem('channels' + channelId)

    const fingerprint = solSha3(
      'updateState',
      channelId,
      sequenceNumber,
      state
    )
    
    const signature = await this.web3.promise.eth.sign(
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
    this.storage.setItem('channels' + channelId, channel)
    
    await post(channel.counterpartyUrl + '/add_proposed_update', update)
  }
  
  

  // Called by the counterparty over the http api, gets verified and
  // added to the proposed update box
  async addProposedUpdate(update) {
    const channel = this.storage.getItem('channels' + update.channelId)
    
    verifyUpdate(channel, update)
    
    channel.theirProposedUpdates.push(update)
    this.storage.setItem('channels' + update.channelId, channel)
  }

  

  // Sign the update and send it back to the counterparty
  async acceptUpdate(update) {
    const channel = this.storage.getItem('channels' + update.channelId)
    
    const fingerprint = verifyUpdate(channel, update)

    const signature = await this.web3.promise.eth.sign(
      channel['address' + channel.me],
      fingerprint
    )

    update['signature' + channel.me] = signature
    
    channel.acceptedUpdates.push(update)
    this.storage.setItem('channels' + update.channelId, channel)
    
    await post(channel.counterpartyUrl + '/add_accepted_update', update)
  }



  // Called by the counterparty over the http api, gets verified and
  // added to the accepted update box
  async addAcceptedUpdate(update) {
    const channel = this.storage.getItem('channels' + update.channelId)
    
    verifyUpdate(channel, update)
    
    channel.acceptedUpdates.push(update)
    this.storage.setItem('channels' + update.channelId, channel)
  }



  // Post an update to the blockchain
  async postUpdate(update) {
    await this.channels.updateState(
      update.channelId,
      update.sequenceNumber,
      update.state,
      update.signature0,
      update.signature1
    )
  }

  // Start the challenge period, putting channel closing into motion
  async startChallengePeriod(
    channelId
  ) {
    const channel = this.storage.getItem('channels' + channelId)
    const fingerprint = solSha3(
      'startChallengePeriod',
      channelId
    )
    
    const signature = await this.web3.promise.eth.sign(
      channel['address' + channel.me],
      fingerprint
    )
    
    await this.channels.startChallengePeriod(
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

  const valid = await this.ec.ecverify.call(
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
        return this.web3.toHex(arg).slice(2)
      }
    }

    if (typeof arg === 'number') {
      return leftPad((arg).toString(16), 64, 0)
    }
  })

  args = args.join('')

  return '0x' + this.web3.sha3(args, { encoding: 'hex' })
}