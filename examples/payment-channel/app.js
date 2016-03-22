const request = require('request')
const minimist = require('minimist')

let commands = {
  // View all my accounts
  viewAccounts() {
    get('http://localhost:4545/view_accounts', function (err, res, accounts) {
      if (err) { console.log(err) }
      console.log(accounts)
    })
  },
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



  // Make a new channel
  newChannel({
    channelId,
    accountPubkey,
    counterpartyPubkey,
    myBalance,
    counterpartyBalance
  }) {
    post('http://localhost:4545/new_channel', {
      channelId,
      accountPubkey,
      counterpartyPubkey,
      state: {
        [accountPubkey]: myBalance,
        [counterpartyPubkey]: counterpartyBalance
      },
      holdPeriod: 1000 * 60 * 60 * 24 * 14 // 2 weeks
    }, function (err) {
      if (err) { console.log(err) }
    })
  },



  // Sign the opening tx and send it to the judge to open the channel
  acceptChannel({ channelId }) {
    post('http://localhost:4545/accept_channel', { channelId }, function (err) {
      if (err) { console.log(err) }
    })
  },



  // Sign a closing tx and send it to the judge to cancel the channel
  cancelChannel({ channelId }) {
    post('http://localhost:4545/cancel_channel', { channelId }, function (err) {
      if (err) { console.log(err) }
    })
  },



  // Make a payment
  pay ({ channelId, amount }) {
    get('http://localhost:4545/get_channel/' + channelId, function (err, res, channel) {
      if (err) { console.log(err) }

      // Get state of last full update tx, or use opening tx
      let state = channel.last_full_update_tx ?
        channel.lastFullUpdateTx.state :
        channel.openingTx.state

      // Add to their balance and subtract from mine
      state[channel.account.pubkey] -= amount
      state[channel.counterparty.pubkey] += amount

      // Don't go below zero
      if (state[channel.account.pubkey] < 0) {
        return console.log('Not enough in account')
      }

      post('http://localhost:4545/new_update_tx', {
        channelId,
        state
      }, function (err, res, body) {
        if (err) { console.log(err) }
      })
    })
  },



  // Accept any valid payments and show the channel's balances
  checkBalance ({ channelId }) {
    get('http://localhost:4545/get_channel/' + channelId, function (err, res, channel) {
      if (err) {
        return console.log(err)
      }

      let myPubkey = channel.account.pubkey
      let theirPubkey = channel.counterparty.pubkey

      // Get state of last full update tx, or use opening tx
      let state = channel.lastFullUpdateTx ?
        channel.lastFullUpdateTx.state :
        channel.openingTx.state

      // If the counterparty has sent a payment
      if (channel.theirProposedUpdateTx) {
        let newState = channel.theirProposedUpdateTx.state

        // Get difference in balances
        let diff = {
          [myPubkey]: state[myPubkey] - newState[myPubkey],
          [theirPubkey]: state[theirPubkey] - newState[theirPubkey]
        }

        if (newState[myPubkey] < 0 || newState[theirPubkey] < 0) {
          console.log('Payment rejected: Balances cannot go below zero.')
          post('localhost:4545/reject_update_tx', {
            channelId
          })

        } else if (diff[myPubkey] !== -diff[channel.counterparty.pubkey]) {
          console.log('Payment rejected: Balances must be raised and lowered by the same amount.')
          post('localhost:4545/reject_update_tx', {
            channelId
          })

        } else if (diff[myPubkey] < 0) {
          console.log('Payment rejected: My balance must be higher than it was before.')
          post('localhost:4545/reject_update_tx', {
            channelId
          })

        } else {
          console.log('New payment of ' + diff[myPubkey] + ' received')
          post('localhost:4545/accept_update_tx', {
            channelId
          })

          state = newState
        }
      }
      console.log({
        myBalance: state[myPubkey],
        counterpartyBalance: state[theirPubkey]
      })
    })
  },



  // Send the last full update tx to the judge to close the channel
  closeChannel({ channelId }) {
    post('http://localhost:4545/close_channel', { channelId }, function (err) {
      if (err) { console.log(err) }
    })
  },



  // This needs to be called at least once per hold period to prevent cheating
  checkChannel({ channelId }) {
    post('http://localhost:4545/check_channel', { channelId }, function (err) {
      if (err) { console.log(err) }
    })
  },
}

function get (url, callback) {
  request.get({
    url,
    json: true,
  }, callback)
}

function post (url, body, callback) {
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
