"use strict";

var _get = function get(_x, _x2, _x3) { var _again = true; _function: while (_again) { var object = _x, property = _x2, receiver = _x3; _again = false; if (object === null) object = Function.prototype; var desc = Object.getOwnPropertyDescriptor(object, property); if (desc === undefined) { var parent = Object.getPrototypeOf(object); if (parent === null) { return undefined; } else { _x = parent; _x2 = property; _x3 = receiver; _again = true; desc = parent = undefined; continue _function; } } else if ("value" in desc) { return desc.value; } else { var getter = desc.get; if (getter === undefined) { return undefined; } return getter.call(receiver); } } };

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var factory = function factory(Pudding) {
  // Inherit from Pudding. The dependency on Babel sucks, but it's
  // the easiest way to extend a Babel-based class. Note that the
  // resulting .js file does not have a dependency on Babel.

  var MetaCoin = (function (_Pudding) {
    _inherits(MetaCoin, _Pudding);

    function MetaCoin() {
      _classCallCheck(this, MetaCoin);

      _get(Object.getPrototypeOf(MetaCoin.prototype), "constructor", this).apply(this, arguments);
    }

    return MetaCoin;
  })(Pudding);

  ;

  // Set up specific data for this class.
  MetaCoin.abi = [{ "constant": true, "inputs": [{ "name": "", "type": "bytes32" }], "name": "channels", "outputs": [{ "name": "channelId", "type": "bytes32" }, { "name": "state", "type": "bytes" }], "type": "function" }, { "constant": false, "inputs": [{ "name": "receiver", "type": "address" }, { "name": "amount", "type": "uint256" }], "name": "sendCoin", "outputs": [{ "name": "sufficient", "type": "bool" }], "type": "function" }, { "constant": false, "inputs": [{ "name": "channelId", "type": "bytes32" }, { "name": "state", "type": "bytes" }], "name": "addChannel", "outputs": [], "type": "function" }, { "constant": false, "inputs": [{ "name": "channelId", "type": "bytes32" }], "name": "getChannelState", "outputs": [{ "name": "", "type": "bytes" }], "type": "function" }, { "constant": false, "inputs": [{ "name": "addr", "type": "address" }], "name": "getBalance", "outputs": [{ "name": "", "type": "uint256" }], "type": "function" }, { "inputs": [], "type": "constructor" }, { "anonymous": false, "inputs": [{ "indexed": false, "name": "", "type": "bytes" }], "name": "Log", "type": "event" }];
  MetaCoin.binary = "6060604052600160a060020a033216600090815260208190526040902061271090556104768061002f6000396000f3606060405260e060020a60003504637a7ebd7b811461004757806390b98a1114610065578063bf477ebc14610094578063d18da8b11461021b578063f8b2cb4f14610296575b005b6102bb60043560016020819052600091825260409091208054910182565b61034a60043560243533600160a060020a0316600090815260208190526040812054829010156103ca576103fc565b60408051602060248035600481810135601f810185900485028601850190965285855261004595813595919460449492939092019181908401838280828437509496505050505050506040805160208181018084526000808452845180840186528181529091528351828152855181840152855193947fafabcf2dd47e06a477a89e49c03f8ebe8e0a7e94f775b25bbb24227c9d0110b294879492938493928401928681019291829185918391869190600490601f850104600302600f01f150905090810190601f16801561017d5780820380516001836020036101000a031916815260200191505b509250505060405180910390a15060408051808201825283815260208181018481526000868152600180845294812084518155915180518387018054818552938690209697889795969195600291861615610100026000190190951604601f90810182900485019492939091019083901061044257805160ff19168380011785555b506104399291505b808211156104725760008155600101610207565b61035c60043560408051602081810183526000808352848152600180835290849020810180548551600293821615610100026000190190911692909204601f8101849004840283018401909552848252929390929183018282801561042d5780601f106104025761010080835404028352916020019161042d565b61034a600435600160a060020a0381166000908152602081905260409020545b919050565b6040805183815260208101828152835460026001821615610100026000190190911604928201839052909160608301908490801561033a5780601f1061030f5761010080835404028352916020019161033a565b820191906000526020600020905b81548152906001019060200180831161031d57829003601f168201915b5050935050505060405180910390f35b60408051918252519081900360200190f35b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600302600f01f150905090810190601f1680156103bc5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b5033600160a060020a039081166000908152602081905260408082208054859003905591841681522080548201905560015b92915050565b820191906000526020600020905b81548152906001019060200180831161041057829003601f168201915b505050505090506102b6565b50505050505050565b828001600101855582156101ff579182015b828111156101ff578251826000505591602001919060010190610454565b509056";

  if ("0x9f3c3c30cf987abb308827dc7668c2fafc1cdfb8" != "") {
    MetaCoin.address = "0x9f3c3c30cf987abb308827dc7668c2fafc1cdfb8";

    // Backward compatibility; Deprecated.
    MetaCoin.deployed_address = "0x9f3c3c30cf987abb308827dc7668c2fafc1cdfb8";
  }

  MetaCoin.generated_with = "1.0.3";
  MetaCoin.contract_name = "MetaCoin";

  return MetaCoin;
};

// Nicety for Node.
factory.load = factory;

if (typeof module != "undefined") {
  module.exports = factory;
} else {
  // There will only be one version of Pudding in the browser,
  // and we can use that.
  window.MetaCoin = factory;
}