# Universal State Channels

Universal State Channels (or USC) is a platform for building state channels using centralized or decentralized judges.

*What is a state channel?* A state channel is a way for two or more parties to maintain a state (any sequence of bytes) between them, being able to keep it updated without having to trust each other.

*What are state channels used for?* The most well-known use of state channels is in payment channels, a technique which allows two parties to pay each other by adjusting the balance of an escrow account or a special blockchain transaction. Payment channels can also be linked together, allowing payments to be securely transmitted across multiple untrusted parties. Participants in the channel exchanged signed state update transactions, and can close the channel at any time with confidence that the last valid update transaction will be honored. Because only the last update transaction is sent to the judge, they can be used to create very scalable systems. [Here's](http://www.jeffcoleman.ca/state-channels/) an easy explanation of state channels.

Payment channels, and any other state chanel, can be implemented easily on top of USC with a minimum of effort. [Here's](https://github.com/jtremback/universal-state-channels/blob/master/examples/payment-channel/app.js) a 200 line Javascript example app implementing a simple payment channel.

USC Peer handles communication between the participants in the channel, and cryptographically checks all transactions for validity. The developer writing the channel's business logic interacts with a friendly HTTP API over localhost. No specialized knowledge of cryptography is required.

USC Judge is run by a trusted third party and checks the validity of a state channel's final update transaction. It also has an easy API, which the developer can use to write the business logic which reacts to the state contained in the final update transaction. For example, in a payment channel, the USC Judge could be run by a bank and permanently transfer money when the channel closed.

USC will also have blockchain adapters, which run alongside the USC Peer. Blockchain adapters relay transactions to a judge contract on a blockchain. This allows the trusted third party to be replaced by trust in a programmable blockchain, such as Ethereum, Tendermint, or Hyperledger.