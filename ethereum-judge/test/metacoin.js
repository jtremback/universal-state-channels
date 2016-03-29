contract('MetaCoin', function(accounts) {
  it("add channel", function(done) {
    var meta = MetaCoin.deployed();
    
    meta.addChannel(
        0x0000000000000000000000000000000000000000000000000000000000000000,
        
        // 0x47995556cf3633cd22e4ea51dfaf52b49a9a1d2eb52ddf8fcd309f4bed33c800,
        
        // 0x47995556cf3633cd22e4ea51dfaf52b49a9a1d2eb52ddf8fcd309f4bed33c800,
        
        // 172800000,
        
        // 0xa6b3556fd0b6eb4c042d9dd1626ac9f53b19ff63421987140556524861d4b184,
        // 0x1c5461f65bb15b4570b9ac1f3f974a1af705487f45246651252152fc439a45ef4dea602b2c29c98bd80626e2cefda1d8cfd8454919e5d9649697f8c310b7f502,
        // 0xf92052e4f008ac07c555c360e4bf885b21568fe81892a4abb3cc1f74a56c181802ca0d20fab2513c4773d89d73dbd842221c1956895209fcae09db6fa741ec07,
        
        0x1111
    ).then(function() {
        return meta.getChannelState.call(0x0000000000000000000000000000000000000000000000000000000000000000)
    }).then(function(state) {
        assert.equal(state, 0x1111, "state was not equal");
    }).then(done).catch(done);
  });

//   it("add channel", function(done) {
//     var meta = MetaCoin.deployed();

//     var allEvents = meta.allEvents();

//     allEvents.watch(function(error, result){
//         console.log(error, result)
//     });

//     meta.addChannel(0x1111).then(function() {
//         return meta.getChannelState.call()
//     }).then(function(state) {
//         assert.equal(state, 0xffff, "state was not equal");
//     }).then(done).catch(done);
//   });

});

