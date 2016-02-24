var request = require('request');

// OpeningTx and UpdateTx state:
//
// {
//   amount1: 100,
//   amount2: 100
// }

function get (url, callback) {
  request.get({
    url,
    json: true,
  }, callback)
}

function post (url, body, callback) {
  request.post({
    url,
    body
    json: true,
  }, callback)
}

function newChannel (channelId, myAccountPubkey, theirAccountPubkey, myAmount, theirAmount) {
  request.post({
    url: 'localhost:4545/new_channel',
    json: true,
    body: {
      channelId,
      myAccountPubkey,
      theirAccountPubkey,
      state: {
        [myAccountPubkey]: myAmount,
        [theirAccountPubkey]: theirAmount
      }
      holdPeriod: 1000 * 60 * 60 * 24 * 14 // 2 weeks
    }
  }, function (err) {
    if (err) { console.log(err) }
  })
}

function pay (channelId, amount) {
  get('localhost:4545/get_channel/' + channelId, function (err, res, channel) {
    if (err) { console.log(err) }

    state = channel.lastFullUpdateTx ?
      channel.lastFullUpdateTx.state :
      channel.OpeningTx.state

    state[channel.account.pubkey] -= amount
    state[channel.counterparty.pubkey] += amount

    if (state[myAccountPubkey] < 0) {
      return console.log("Not enough in account")
    }

    request.post({
      url: 'localhost:4545/new_update_tx',
      json: true,
      body: {
        channelId,
        state
      }
    }, function (err, res, body) {
      if (err) { console.log(err) }
    })
  })
}

function checkAmounts (channelId) {
  get('localhost:4545/get_channel/' + channelId, function (err, res, channel) {
    if (err) {
      return console.log(err)
    }

    state = channel.lastFullUpdateTx ?
      channel.lastFullUpdateTx.state :
      channel.OpeningTx.state

    if (channel.TheirProposedUpdateTx) {
      if () {
        post('localhost:4545/confirm_update_tx', {
          channelId
        })
      }
    }

    post('localhost:4545/new_update_tx', {
      channelId,
      state
    }, function (err, res, body) {
      if (err) { console.log(err) }
    })
  })
}
