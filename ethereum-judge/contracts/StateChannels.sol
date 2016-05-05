import "ECVerify.sol";

contract StateChannels is ECVerify {    
    uint8 constant PHASE_OPEN = 0;
    uint8 constant PHASE_CHALLENGE = 1;
    uint8 constant PHASE_CLOSED = 2;
    
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
    }
    
    function getChannel(bytes32 channelId) returns(
        address addr0,
        address addr1,
        uint8 phase,
        uint challengePeriod,
        uint closingBlock,
        bytes state,
        uint sequenceNumber
    ) {
        addr0 = channels[channelId].addr0;
        addr1 = channels[channelId].addr1;
        phase = channels[channelId].phase;
        challengePeriod = channels[channelId].challengePeriod;
        closingBlock = channels[channelId].closingBlock;
        state = channels[channelId].state;
        sequenceNumber = channels[channelId].sequenceNumber;
    }

    event Error(string message);
    event LogString(string label, string message);
    event LogBytes(string label, bytes message);
    event LogBytes32(string label, bytes32 message);
    event LogNum256(uint256 num);
    
    function newChannel(
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
            'newChannel',
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
        
        Channel memory channel = Channel(
            channelId,
            addr0,
            addr1,
            PHASE_OPEN,
            challengePeriod,
            0,
            state,
            0
        );
        
        channels[channelId] = channel;
    }
    
    function updateState(
        bytes32 channelId,
        uint256 sequenceNumber,
        bytes state,
        bytes signature0,
        bytes signature1
    ) {
        tryClose(channelId);
        
        if (channels[channelId].phase == PHASE_CLOSED) {
            Error("channel closed");
            return;
        }
        
        bytes32 fingerprint = sha3(
            'updateState',
            channelId,
            sequenceNumber,
            state
        );
        
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
    
    function startChallengePeriod(
        bytes32 channelId,
        bytes signature,
        uint8 participant
    ) {
        if (channels[channelId].phase != PHASE_OPEN) {
            Error("channel not open");
            return;
        }
        
        bytes32 fingerprint = sha3(
            'startChallengePeriod',
            channelId
        );
        
        if (participant == 0) {
            if (!ecverify(fingerprint, signature, channels[channelId].addr0)) {
                Error("signature invalid");
                return;
            }
        } else if (participant == 1) {
            if (!ecverify(fingerprint, signature, channels[channelId].addr1)) {
                Error("signature invalid");
                return;
            }
        } else {
            Error("participant invalid");
            return;
        }
        
        channels[channelId].closingBlock = block.number + channels[channelId].challengePeriod;
        channels[channelId].phase = PHASE_CHALLENGE;
    }
    
    function tryClose(
        bytes32 channelId
    ) {
        if (
            channels[channelId].phase == PHASE_CHALLENGE &&
            block.number > channels[channelId].closingBlock
        ) {
            channels[channelId].phase = PHASE_CLOSED;
        }
    }
}
