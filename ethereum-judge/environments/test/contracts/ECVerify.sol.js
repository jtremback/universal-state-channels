// Factory "morphs" into a Pudding class.
// The reasoning is that calling load in each context
// is cumbersome.

(function() {

  var contract_data = {
    abi: [{"constant":false,"inputs":[{"name":"hash","type":"bytes32"},{"name":"sig","type":"bytes"},{"name":"signer","type":"address"}],"name":"ecverify","outputs":[{"name":"","type":"bool"}],"type":"function"},{"constant":false,"inputs":[{"name":"hash","type":"bytes32"},{"name":"sig","type":"bytes"}],"name":"ecrecovery","outputs":[{"name":"","type":"address"}],"type":"function"}],
    binary: "60606040526101a0806100126000396000f3606060405260e060020a600035046339cdde32811461002657806377d32e9414610087575b005b60408051602060046024803582810135601f81018590048502860185019096528585526100ee95833595939460449493929092019181908401838280828437509496505093359350505050600081600160a060020a031661018e85856100d1565b60408051602060046024803582810135601f810185900485028601850190965285855261010095833595939460449493929092019181908401838280828437509496505050505050505b60006000600060008451604114151561011d575b50505092915050565b60408051918252519081900360200190f35b60408051600160a060020a03929092168252519081900360200190f35b505050602082015160408301516041840151600116601b81101561013f57601b015b600186828585604051808581526020018460ff1681526020018381526020018281526020019450505050506020604051808303816000866161da5a03f1156100025750506040515193506100e5565b600160a060020a03161494935050505056",
    unlinked_binary: "60606040526101a0806100126000396000f3606060405260e060020a600035046339cdde32811461002657806377d32e9414610087575b005b60408051602060046024803582810135601f81018590048502860185019096528585526100ee95833595939460449493929092019181908401838280828437509496505093359350505050600081600160a060020a031661018e85856100d1565b60408051602060046024803582810135601f810185900485028601850190965285855261010095833595939460449493929092019181908401838280828437509496505050505050505b60006000600060008451604114151561011d575b50505092915050565b60408051918252519081900360200190f35b60408051600160a060020a03929092168252519081900360200190f35b505050602082015160408301516041840151600116601b81101561013f57601b015b600186828585604051808581526020018460ff1681526020018381526020018281526020019450505050506020604051808303816000866161da5a03f1156100025750506040515193506100e5565b600160a060020a03161494935050505056",
    address: "",
    generated_with: "2.0.6",
    contract_name: "ECVerify"
  };

  function Contract() {
    if (Contract.Pudding == null) {
      throw new Error("ECVerify error: Please call load() first before creating new instance of this contract.");
    }

    Contract.Pudding.apply(this, arguments);
  };

  Contract.load = function(Pudding) {
    Contract.Pudding = Pudding;

    Pudding.whisk(contract_data, Contract);

    // Return itself for backwards compatibility.
    return Contract;
  }

  Contract.new = function() {
    if (Contract.Pudding == null) {
      throw new Error("ECVerify error: Please call load() first before calling new().");
    }

    return Contract.Pudding.new.apply(Contract, arguments);
  };

  Contract.at = function() {
    if (Contract.Pudding == null) {
      throw new Error("ECVerify error: lease call load() first before calling at().");
    }

    return Contract.Pudding.at.apply(Contract, arguments);
  };

  Contract.deployed = function() {
    if (Contract.Pudding == null) {
      throw new Error("ECVerify error: Please call load() first before calling deployed().");
    }

    return Contract.Pudding.deployed.apply(Contract, arguments);
  };

  if (typeof module != "undefined" && typeof module.exports != "undefined") {
    module.exports = Contract;
  } else {
    // There will only be one version of Pudding in the browser,
    // and we can use that.
    window.ECVerify = Contract;
  }

})();
