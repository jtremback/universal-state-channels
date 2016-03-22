'use strict';

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

var request = require('request');
var minimist = require('minimist');

var commands = {
  // View all my accounts

  viewAccounts: function viewAccounts() {
    get('http://localhost:4545/view_accounts', function (err, res, accounts) {
      if (err) {
        console.log(err);
      }
      console.log(accounts);
    });
  },

  // View all channels for a given account
  viewChannels: function viewChannels(_ref) {
    var accountPubkey = _ref.accountPubkey;

    get('http://localhost:4545/view_channels/' + accountPubkey, function (err, res, channels) {
      if (err) {
        console.log(err);
      }
      console.log(channels);
    });
  },

  // View all counterparties
  viewCounterparties: function viewCounterparties() {
    get('http://localhost:4545/view_counterparties', function (err, res, counterparties) {
      if (err) {
        console.log(err);
      }
      console.log(counterparties);
    });
  },


  // Make a new channel
  newChannel: function newChannel(_ref2) {
    var _state;

    var channelId = _ref2.channelId;
    var accountPubkey = _ref2.accountPubkey;
    var counterpartyPubkey = _ref2.counterpartyPubkey;
    var myBalance = _ref2.myBalance;
    var counterpartyBalance = _ref2.counterpartyBalance;

    post('http://localhost:4545/new_channel', {
      channelId: channelId,
      accountPubkey: accountPubkey,
      counterpartyPubkey: counterpartyPubkey,
      state: (_state = {}, _defineProperty(_state, accountPubkey, myBalance), _defineProperty(_state, counterpartyPubkey, counterpartyBalance), _state),
      holdPeriod: 1000 * 60 * 60 * 24 * 14 // 2 weeks
    }, function (err) {
      if (err) {
        console.log(err);
      }
    });
  },


  // Sign the opening tx and send it to the judge to open the channel
  acceptChannel: function acceptChannel(_ref3) {
    var channelId = _ref3.channelId;

    post('http://localhost:4545/accept_channel', { channelId: channelId }, function (err) {
      if (err) {
        console.log(err);
      }
    });
  },


  // Sign a closing tx and send it to the judge to cancel the channel
  cancelChannel: function cancelChannel(_ref4) {
    var channelId = _ref4.channelId;

    post('http://localhost:4545/cancel_channel', { channelId: channelId }, function (err) {
      if (err) {
        console.log(err);
      }
    });
  },


  // Make a payment
  pay: function pay(_ref5) {
    var channelId = _ref5.channelId;
    var amount = _ref5.amount;

    get('http://localhost:4545/get_channel/' + channelId, function (err, res, channel) {
      if (err) {
        console.log(err);
      }

      // Get state of last full update tx, or use opening tx
      var state = channel.last_full_update_tx ? channel.lastFullUpdateTx.state : channel.openingTx.state;

      // Add to their balance and subtract from mine
      state[channel.account.pubkey] -= amount;
      state[channel.counterparty.pubkey] += amount;

      // Don't go below zero
      if (state[channel.account.pubkey] < 0) {
        return console.log('Not enough in account');
      }

      post('http://localhost:4545/new_update_tx', {
        channelId: channelId,
        state: state
      }, function (err, res, body) {
        if (err) {
          console.log(err);
        }
      });
    });
  },


  // Accept any valid payments and show the channel's balances
  checkBalance: function checkBalance(_ref6) {
    var channelId = _ref6.channelId;

    get('http://localhost:4545/get_channel/' + channelId, function (err, res, channel) {
      if (err) {
        return console.log(err);
      }

      var myPubkey = channel.account.pubkey;
      var theirPubkey = channel.counterparty.pubkey;

      // Get state of last full update tx, or use opening tx
      var state = channel.lastFullUpdateTx ? channel.lastFullUpdateTx.state : channel.openingTx.state;

      // If the counterparty has sent a payment
      if (channel.theirProposedUpdateTx) {
        var _diff;

        var newState = channel.theirProposedUpdateTx.state;

        // Get difference in balances
        var diff = (_diff = {}, _defineProperty(_diff, myPubkey, state[myPubkey] - newState[myPubkey]), _defineProperty(_diff, theirPubkey, state[theirPubkey] - newState[theirPubkey]), _diff);

        if (newState[myPubkey] < 0 || newState[theirPubkey] < 0) {
          console.log('Payment rejected: Balances cannot go below zero.');
          post('localhost:4545/reject_update_tx', {
            channelId: channelId
          });
        } else if (diff[myPubkey] !== -diff[channel.counterparty.pubkey]) {
          console.log('Payment rejected: Balances must be raised and lowered by the same amount.');
          post('localhost:4545/reject_update_tx', {
            channelId: channelId
          });
        } else if (diff[myPubkey] < 0) {
          console.log('Payment rejected: My balance must be higher than it was before.');
          post('localhost:4545/reject_update_tx', {
            channelId: channelId
          });
        } else {
          console.log('New payment of ' + diff[myPubkey] + ' received');
          post('localhost:4545/accept_update_tx', {
            channelId: channelId
          });

          state = newState;
        }
      }
      console.log({
        myBalance: state[myPubkey],
        counterpartyBalance: state[theirPubkey]
      });
    });
  },


  // Send the last full update tx to the judge to close the channel
  closeChannel: function closeChannel(_ref7) {
    var channelId = _ref7.channelId;

    post('http://localhost:4545/close_channel', { channelId: channelId }, function (err) {
      if (err) {
        console.log(err);
      }
    });
  },


  // This needs to be called at least once per hold period to prevent cheating
  checkChannel: function checkChannel(_ref8) {
    var channelId = _ref8.channelId;

    post('http://localhost:4545/check_channel', { channelId: channelId }, function (err) {
      if (err) {
        console.log(err);
      }
    });
  }
};

function get(url, callback) {
  request.get({
    url: url,
    json: true
  }, callback);
}

function post(url, body, callback) {
  request.post({
    url: url,
    body: body,
    json: true
  }, callback);
}

var argv = require('minimist')(process.argv.slice(2));
if (commands[argv._[0]]) {
  commands[argv._[0]](argv);
} else if (argv._[0]) {
  console.log(argv._[0] + ' is not a command');
} else {
  console.log('please enter a command');
}
