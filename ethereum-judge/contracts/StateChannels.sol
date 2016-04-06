contract StateChannels {    
	mapping (address => uint) balances;

	function StateChannels() {
		balances[tx.origin] = 10000;
	}

	function sendCoin(address receiver, uint amount) returns(bool sufficient) {
		if (balances[msg.sender] < amount) return false;
		balances[msg.sender] -= amount;
		balances[receiver] += amount;
		return true;
	}

    function getBalance(address addr) returns(uint) {
        return balances[addr];
    }
    
    function getChannelState(bytes32 channelId) returns(bytes) {
        return channels[channelId].state;
    }
    
    mapping (bytes32 => Channel) public channels;
    
    struct Channel {
        bytes32 channelId;
        bytes32 pubkey0;
        bytes32 pubkey1;
        bytes state;
        bytes32 fingerprint;
        bytes signature0;
        bytes signature1;
        uint8 phase;
    }
    
    event Error(string message);
    event LogString(string label, string message);
    event LogBytes(string label, bytes32 message);
    event LogBytes32(string label, bytes32 message);
    
    function addChannel(
        bytes32 channelId,
        bytes32 pubkey0,
        bytes32 pubkey1,
        bytes state,
        bytes32 fingerprint,
        bytes signature0,
        bytes signature1
    ) {
        if (channels[channelId].channelId == channelId) {
            Error("channel with that channelId already exists");
            return;
        }
        
        if (fingerprint != sha3(
            channelId,
            pubkey0,
            pubkey1,
            state
        )) {
            Error("fingerprint does not match");
            return;
        }
        
        Channel memory ch = Channel(
            channelId,
            pubkey0,
            pubkey1,
            state,
            fingerprint,
            signature0,
            signature1,
            0
        );
        
        channels[channelId] = ch;
    }
}
