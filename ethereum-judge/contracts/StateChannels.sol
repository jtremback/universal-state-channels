import "ECVerify.sol";

contract StateChannels is ECVerify {
    // Not sure if we need this stuff
    // uint constant UPDATE_LIMIT = 100;
    // uint constant EVIDENCE_LIMIT = 100;
    
    uint8 constant PHASE_OPEN = 0;
    uint8 constant PHASE_CHALLENGE = 1;
    uint8 constant PHASE_CLOSED = 2;
     
    function getChannelState(bytes32 channelId) returns(bytes) {
        return channels[channelId].state;
    }
    
    mapping (bytes32 => Channel) channels;
    
    struct Channel {
        bytes32 channelId;
        address addr0;
        address addr1;
        uint8 phase;
        uint challengePeriod;
        uint closingBlock;
        bytes state;
        uint sequenceNumber;
        bytes[] evidence;
    }

    event Error(string message);
    event LogString(string label, string message);
    event LogBytes(string label, bytes32 message);
    event LogBytes32(string label, bytes32 message);
    event LogNum256(uint256 num);
    
    function addChannel(
        bytes32 channelId,
        address addr0,
        address addr1,
        bytes state,
        uint256 challengePeriod,
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
        
        bytes[] memory evidence;
        
        Channel memory channel = Channel(
            channelId,
            addr0,
            addr1,
            PHASE_OPEN,
            challengePeriod,
            0,
            state,
            0,
            evidence
        );
        
        channels[channelId] = channel;
    }
    
    function addUpdateTx(
        bytes32 channelId,
        uint256 sequenceNumber,
        bytes state,
        bytes signature0,
        bytes signature1
    ) { 
        bytes32 fingerprint = sha3(
            channelId,
            sequenceNumber,
            state
        );
        
        LogNum256(sequenceNumber);
        LogBytes32('fingerprint', fingerprint);
        
        if (!ecverify(fingerprint, signature0, channels[channelId].addr0)) {
            Error("signature0 invalid");
            return;
        }
        
        if (!ecverify(fingerprint, signature1, channels[channelId].addr1)) {
            Error("signature1 invalid");
            return;
        }

        if (sequenceNumber <= channels[channelId].sequenceNumber) {
            Error("sequence number too low");
            return;
        }
        
        channels[channelId].state = state;
        channels[channelId].sequenceNumber = sequenceNumber;
    }
    
    // function addClosingTx(
    //     bytes32 channelId,
    //     bytes signature
    // ) {
    //     if (channels[channelId].phase != PHASE_OPEN) {
    //         Error("channel not open");
    //         return;
    //     }
        
    //     bytes32 fingerprint = sha3(
    //         channelId
    //     );
        
    //     if (
    //         !(ecverify(fingerprint, signature, channels[channelId].addr0)) &&
    //         !(ecverify(fingerprint, signature, channels[channelId].addr1))
    //     ) {
    //         Error("signature invalid");
    //         return;
    //     }

    //     channels[channelId].closingBlock = block.timestamp;
    //     channels[channelId].phase = PHASE_CHALLENGE;
    // }
    
    // function addEvidenceTx(
    //     bytes32 channelId,
    //     bytes state,
    //     bytes signature
    // ) {
        
    // }
    
    // function requestSettlement(
    //     bytes32 channelId
    // ) {
    //     if () {
            
    //     }
    // }
}
