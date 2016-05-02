import "ECVerify.sol";

contract StateChannels is ECVerify {    
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
        address addr0;
        address addr1;
        bytes state;
        // bytes32 fingerprint;
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
        address addr0,
        address addr1,
        bytes state,
        // bytes32 fingerprint,
        bytes signature0,
        bytes signature1
    ) { 
        if (channels[channelId].channelId == channelId) {
            Error("channel with that channelId already exists");
            return;
        }
        
        bytes32 fingerprint = sha3(
            channelId,
            addr0,
            addr1,
            state
        );
        
        if (!ecverify(fingerprint, signature0, addr0)) {
            Error("signature0 invalid");
            return;
        }
        
        if (!ecverify(fingerprint, signature1, addr1)) {
            Error("signature1 invalid");
            return;
        }
        
        Channel memory ch = Channel(
            channelId,
            addr0,
            addr1,
            state,
            // fingerprint,
            signature0,
            signature1,
            0
        );
        
        channels[channelId] = ch;
    }
}
