/*global StateChannels*/
import secp256k1 from 'secp256k1'
import crypto from 'crypto'
import sha3 from 'js-sha3'
import leftPad from 'left-pad'
import BigNumber from 'bignumber.js' 

const keccak = sha3.keccak_256

contract('StateChannels', function (accounts) {
    it('adds channel and checks state', mochaAsync(async () => {
        const meta = StateChannels.deployed();
        const channelId = '1000000000000000000000000000000000000000000000000000000000000000'
        const state = '1111'
        const fingerprint = keccak(hexStringToByte(
            web3.toHex('newChannel').slice(2) +
            channelId +
            web3.eth.accounts[0].slice(2) +
            web3.eth.accounts[1].slice(2) +
            state
        ))

        const sig0 = web3.eth.sign(web3.eth.accounts[0], '0x' + fingerprint)
        const sig1 = web3.eth.sign(web3.eth.accounts[1], '0x' + fingerprint)

        await meta.newChannel(
            '0x' + channelId,
            web3.eth.accounts[0],
            web3.eth.accounts[1],
            '0x' + state,
            1,
            sig0,
            sig1
        )

        const savedChannel = await meta.getChannel.call(
            '0x' + channelId
        )
        
        assert.equal(savedChannel[0], web3.eth.accounts[0], 'addr0')
        assert.equal(savedChannel[1], web3.eth.accounts[1], 'addr1')
        assert.equal(savedChannel[2].toString(10), '0', 'phase')
        assert.equal(savedChannel[3].toString(10), '1', 'challengePeriod')
        assert.equal(savedChannel[4].toString(10), '0', 'closingBlock')
        assert.equal(savedChannel[5], '0x' + state, 'state')
        assert.equal(savedChannel[6].toString(10), '0', 'sequenceNumber')
    }));

    it('rejects channel with existant channelId', mochaAsync(async () => {
        const meta = StateChannels.deployed();
        const errLog = meta.Error([{ code: 1 }]);
        const channelId = '1000000000000000000000000000000000000000000000000000000000000000'
        const state = '1111'
        const fingerprint = keccak(hexStringToByte(
            web3.toHex('newChannel').slice(2) +
            channelId +
            web3.eth.accounts[0].slice(2) +
            web3.eth.accounts[1].slice(2) +
            state
        ))

        const sig0 = web3.eth.sign(web3.eth.accounts[0], '0x' + fingerprint)
        const sig1 = web3.eth.sign(web3.eth.accounts[1], '0x' + fingerprint)

        await meta.newChannel(
            '0x' + channelId,
            web3.eth.accounts[0],
            web3.eth.accounts[1],
            '0x' + state,
            1,
            sig0,
            sig1
        )

        const logs = await errLog.get()

        assert.equal('channel with that channelId already exists', logs[0].args.message, 'did not return error');
    }));

    it('rejects channel with non-valid signature0', mochaAsync(async () => {
        const meta = StateChannels.deployed();
        const errLog = meta.Error();
        const channelId = '2000000000000000000000000000000000000000000000000000000000000000'
        const state = '1111'
        const fingerprint = keccak(hexStringToByte(
            web3.toHex('newChannel').slice(2) +
            channelId +
            web3.eth.accounts[0].slice(2) +
            web3.eth.accounts[1].slice(2) +
            state
        ))

        const sig0 = web3.eth.sign(web3.eth.accounts[2], '0x' + fingerprint) // Wrong account
        const sig1 = web3.eth.sign(web3.eth.accounts[1], '0x' + fingerprint)

        await meta.newChannel(
            '0x' + channelId,
            web3.eth.accounts[0],
            web3.eth.accounts[1],
            '0x' + state,
            1,
            sig0,
            sig1
        )
        const logs = await errLog.get()

        assert.equal(logs[0].args.message, 'signature0 invalid', 'did not return error');
    }));

    it('rejects channel with non-valid signature1', mochaAsync(async () => {
        const meta = StateChannels.deployed();
        const errLog = meta.Error();
        const channelId = '3000000000000000000000000000000000000000000000000000000000000000'
        const state = '1111'
        const fingerprint = keccak(hexStringToByte(
            web3.toHex('newChannel').slice(2) +
            channelId +
            web3.eth.accounts[0].slice(2) +
            web3.eth.accounts[1].slice(2) +
            state
        ))

        const sig0 = web3.eth.sign(web3.eth.accounts[0], '0x' + fingerprint)
        const sig1 = web3.eth.sign(web3.eth.accounts[2], '0x' + fingerprint) // Wrong account

        await meta.newChannel(
            '0x' + channelId,
            web3.eth.accounts[0],
            web3.eth.accounts[1],
            '0x' + state,
            1,
            sig0,
            sig1
        )
        const logs = await errLog.get()

        assert.equal(logs[0].args.message, 'signature1 invalid', 'did not return error');
    }));

    it('update state', mochaAsync(async () => {
        const meta = StateChannels.deployed()
        const channelId = '1000000000000000000000000000000000000000000000000000000000000000'
        const state = '2222'
        const sequenceNumber = 1
        const fingerprint = keccak(hexStringToByte(
            web3.toHex('updateState').slice(2) +
            channelId +
            leftPad((sequenceNumber).toString(16), 64, 0) +  
            state
        ))
        
        const sig0 = web3.eth.sign(web3.eth.accounts[0], '0x' + fingerprint)
        const sig1 = web3.eth.sign(web3.eth.accounts[1], '0x' + fingerprint)
        
        await meta.updateState(
            '0x' + channelId,
            web3.toHex(sequenceNumber),
            '0x' + state,
            sig0,
            sig1
        )
        
        const savedChannel = await meta.getChannel.call(
            '0x' + channelId
        )

        assert.equal(savedChannel[5], '0x' + state, 'state')
        assert.equal(savedChannel[6].toString(10), '1', 'sequenceNumber')
    }));
    
    it('start challenge period', mochaAsync(async () => {
        const meta = StateChannels.deployed()
        const channelId = '1000000000000000000000000000000000000000000000000000000000000000'
        const fingerprint = keccak(hexStringToByte(
            web3.toHex('startChallengePeriod').slice(2) +
            channelId
        ))
        
        const sig = web3.eth.sign(web3.eth.accounts[0], '0x' + fingerprint)
        
        await meta.startChallengePeriod(
            '0x' + channelId,
            sig,
            0
        )
        
        const savedChannel = await meta.getChannel.call(
            '0x' + channelId
        )
        
        assert.equal(savedChannel[0], web3.eth.accounts[0], 'addr0')
        assert.equal(savedChannel[1], web3.eth.accounts[1], 'addr1')
        assert.equal(savedChannel[2].toString(10), '1', 'phase')
        assert.equal(savedChannel[3].toString(10), '1', 'challengePeriod')
        assert.isAbove(savedChannel[4].toString(10), '1', 'closingBlock')
        // assert.equal(savedChannel[5], '0x' + state, 'state')
        assert.equal(savedChannel[6].toString(10), '1', 'sequenceNumber')
    }));
});

function mochaAsync(fn) {
    return async (done) => {
        try {
            await fn();
            done();
        } catch (err) {
            done(err);
        }
    };
};


function byteToHexString(uint8arr) {
    if (!uint8arr) {
        return '';
    }

    var hexStr = '';
    for (var i = 0; i < uint8arr.length; i++) {
        var hex = (uint8arr[i] & 0xff).toString(16);
        hex = (hex.length === 1) ? '0' + hex : hex;
        hexStr += hex;
    }

    return hexStr.toUpperCase();
}

function hexStringToByte(str) {
    if (!str) {
        return new Uint8Array();
    }

    var a = [];
    for (var i = 0, len = str.length; i < len; i += 2) {
        a.push(parseInt(str.substr(i, 2), 16));
    }

    return new Uint8Array(a);
}

function concatenate(resultConstructor, ...arrays) {
    let totalLength = 0;
    for (let arr of arrays) {
        totalLength += arr.length;
    }
    let result = new resultConstructor(totalLength);
    let offset = 0;
    for (let arr of arrays) {
        result.set(arr, offset);
        offset += arr.length;
    }
    return result;
}

// Extra Addresses

// 0xF18506eD9AdcA974c0e803859994d11fc8753885
// 579ac7f421b256fd8b9bd7a5f384f6499b98c9409f9d431137b9d69db129d65f

// 0xF8D07F73f5336b8b77D52143906a216E454E8f3a
// 14475f8d92fbee4e20ce9cb8fe8b434e57e5d80787e9ddd433ee66f848210ea9

// 0x8fb411A5Bb2F0fa6B247409F05494B56E9Fa730a
// 16add8e48cdfd12a07dd8ec86db7c284a41fbc0d7454272a332520bd2cf64180

// 0xff7FC071Eb3385D1A810bAABD3d870156a965b12
// 1d9cc52f5a6dbabb5dce7bec96fe729ca45a72323379d48b5641db36d5240c5d

// 0xf8c138b08cb32391C7Ab8Edbda61E023943f72d7
// 6712eb15afa15159ca2f8ae405bb6286929e81b1d1865186717500202cfcf9b8

// 0x763e646f269d9c50f24d2c4802859ccd185148497774bff4525426d4eb771d0b23e5157cc8dba35bd6eb075cbe7e3854e2775ad44f8c5ae3d3c7ec7c278947081b 

// 41b1a0649752af1b28b3dc29a1556eee781e4a4c3a1f7f53f90fa834de098c4d