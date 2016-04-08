/*global StateChannels*/
import secp256k1 from 'secp256k1'
import crypto from 'crypto'
import sha3 from 'js-sha3'
 
// generate message to sign 
var msg = hexStringToByte('0101')
 
// generate privKey 
var privKey
do {
  privKey = crypto.randomBytes(32)
} while (!secp256k1.privateKeyVerify(privKey))
 
// get the public key in a compressed format 
var pubKey = secp256k1.publicKeyCreate(privKey)
 
// sign the message 
var sigObj = secp256k1.sign(msg, privKey)

console.log(web3.eth)

// var sig2Obj = web3.eth.sign(msg, privKey) 

console.log(sigObj)
 
// verify the signature 
console.log(secp256k1.verify(msg, sigObj.signature, pubKey))


const keccak = sha3.keccak_256

const pubkey1 = '47995556cf3633cd22e4ea51dfaf52b49a9a1d2eb52ddf8fcd309f4bed33c800'
const pubkey2 = '47995556cf3633cd22e4ea51dfaf52b49a9a1d2eb52ddf8fcd309f4bed33c800'

contract('StateChannels', function(accounts) {
  it('adds channel and checks state', mochaAsync(async () => {
    const meta = StateChannels.deployed();
    const channelId = '1000000000000000000000000000000000000000000000000000000000000000'
    const state = '1111'
    const fingerprint = keccak(hexStringToByte(
        channelId + pubkey1 + pubkey2 + state
    ))
    
    await meta.addChannel(
        '0x' + channelId,
        '0x' + pubkey1,
        '0x' + pubkey2,
        '0x' + state,
        
        '0x' + fingerprint,
        
        '0x1c5461f65bb15b4570b9ac1f3f974a1af705487f45246651252152fc439a45ef4dea602b2c29c98bd80626e2cefda1d8cfd8454919e5d9649697f8c310b7f502',
        '0xf92052e4f008ac07c555c360e4bf885b21568fe81892a4abb3cc1f74a56c181802ca0d20fab2513c4773d89d73dbd842221c1956895209fcae09db6fa741ec07'
    )
    
    const savedState = await meta.getChannelState.call(
       '0x' + channelId
    )
    
    assert.equal(savedState, '0x' + state, 'state was not equal');
  }));
  
  
  
  it('rejects channel with existant channelId', mochaAsync(async () => {
    const meta = StateChannels.deployed();
    const errLog = meta.Error([{code: 1}]);
    const channelId = '1000000000000000000000000000000000000000000000000000000000000000'
    const state = '1111'
    const fingerprint = keccak(hexStringToByte(
        channelId + pubkey1 + pubkey2 + state
    ))
    
    await meta.addChannel(
        '0x' + channelId,
        '0x' + pubkey1,
        '0x' + pubkey2,
        '0x' + state,
        
        '0x' + fingerprint,
        
        '0x1c5461f65bb15b4570b9ac1f3f974a1af705487f45246651252152fc439a45ef4dea602b2c29c98bd80626e2cefda1d8cfd8454919e5d9649697f8c310b7f502',
        '0xf92052e4f008ac07c555c360e4bf885b21568fe81892a4abb3cc1f74a56c181802ca0d20fab2513c4773d89d73dbd842221c1956895209fcae09db6fa741ec07'
    )
    
    const logs = await errLog.get()
    
    assert.equal('channel with that channelId already exists', logs[0].args.message, 'did not return error');
  }));
  
  
  
  it('rejects channel with non-valid fingerprint', mochaAsync(async () => {
    const meta = StateChannels.deployed();
    const errLog = meta.Error();
    const channelId = '2000000000000000000000000000000000000000000000000000000000000000'
    const state = '1111'
    const fingerprint = keccak(hexStringToByte(
        channelId + pubkey1 + pubkey2 + state
    ))
    
    await meta.addChannel(
        '0x' + channelId,
        '0x' + pubkey1,
        '0x' + pubkey2,
        '0x' + state,
        
        '0x' + fingerprint,
        
        '0x1c5461f65bb15b4570b9ac1f3f974a1af705487f45246651252152fc439a45ef4dea602b2c29c98bd80626e2cefda1d8cfd8454919e5d9649697f8c310b7f502',
        '0xf92052e4f008ac07c555c360e4bf885b21568fe81892a4abb3cc1f74a56c181802ca0d20fab2513c4773d89d73dbd842221c1956895209fcae09db6fa741ec07'
    )
    const logs = await errLog.get()
    
    assert.equal('fingerprint does not match', 'fingerprint does not match', 'did not return error');
  }));
  
  it('ecrecover test', mochaAsync(async () => {
    const meta = StateChannels.deployed();
    
        var sig = secp256k1.sign(msgHash, privateKey)
        var ret = {}
        ret.r = sig.signature.slice(0, 32)
        ret.s = sig.signature.slice(32, 64)
        ret.v = sig.recovery + 27
  })
});

function mochaAsync (fn) {
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
  for (var i = 0, len = str.length; i < len; i+=2) {
    a.push(parseInt(str.substr(i,2),16));
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