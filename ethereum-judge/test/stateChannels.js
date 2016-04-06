/*global StateChannels*/
import secp256k1 from 'secp256k1'
import crypto from 'crypto'
import sha3 from 'js-sha3'
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