//
// The new assembly support in Solidity makes writing helpers easy.
// Many have complained how complex it is to use `ecrecover`, especially in conjunction
// with the `eth_sign` RPC call. Here is a helper, which makes that a matter of a single call.
//
// Sample input parameters:
// (with v=0)
// "0x47173285a8d7341e5e972fc677286384f802f8ef42a5ec5f03bbfa254cb01fad",
// "0xaca7da997ad177f040240cdccf6905b71ab16b74434388c3a72f34fd25d6439346b2bac274ff29b48b3ea6e2d04c1336eaceafda3c53ab483fc3ff12fac3ebf200",
// "0x0e5cb767cce09a7f3ca594df118aa519be5e2b5a"
//
// (with v=1)
// "0x47173285a8d7341e5e972fc677286384f802f8ef42a5ec5f03bbfa254cb01fad",
// "0xdebaaa0cddb321b2dcaaf846d39605de7b97e77ba6106587855b9106cb10421561a22d94fa8b8a687ff9c911c844d1c016d1a685a9166858f9c7c1bc85128aca01",
// "0x8743523d96a1b2cbe0c6909653a56da18ed484af"
//
// (The hash is a hash of "hello world".)
//
// Written by Alex Beregszaszi (@axic), use it under the terms of the MIT license.
//

contract ECVerify {
    event LogNum(uint8 num);
    event LogNum256(uint256 num);
    event LogBool(bool b);
    function ecrecovery(bytes32 hash, bytes sig) returns (address) {
        bytes32 r;
        bytes32 s;
        uint8 v;
        
        // FIXME: Should this throw, or return 0?
        if (sig.length != 65) {
            return 0;
        }

        // The signature format is a compact form of:
        //   {bytes32 r}{bytes32 s}{uint8 v}
        // Compact means, uint8 is not padded to 32 bytes.
        assembly {
            r := mload(add(sig, 32))
            s := mload(add(sig, 64))
            v := mload(add(sig, 65))
        }
        
        // old geth sends a `v` value of [0,1], while the new, in line with the YP sends [27,28]
        if (v < 27)
          v += 27;

        return ecrecover(hash, v, r, s);
    }
    
    function ecverify(bytes32 hash, bytes sig, address signer) returns (bool b) {
        b = ecrecovery(hash, sig) == signer;
        LogBool(b);
        return b;
    }
}