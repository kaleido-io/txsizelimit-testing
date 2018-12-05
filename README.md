# txsizelimit-testing
Here we test the network effects of the transaction size limit flag submitted to Quorum: https://github.com/jpmorganchase/quorum/pull/575

## Motivation for PR

Our PR inserts a transaction size limit flag (`--txsizelimit`) as part of the geth cli. This allows the user to set the maximum size in KB that a node can process, validate, mine, and broadcast. We believe this is necessary for many enterprise blockchain clients in various scenarios, such as multi-party signatures. These tests allow us to see how nodes of different `txsizelimit`s interact before making the change live.

## Test Details

Here we present a test scenario where we have 3 nodes, with `txsizelimit`=[40,32,32]. A transaction of size 39KB is sent to node 1 (limit=40), and with the additional logging statements we observe how each node processes the transaction.

## Primary findings
In our test scenario, we uncover that mixed transaction size limits across the blockchain network does not result in non-determinism and corrupt the blockchain.  

When a 39KB transaction is broadcast to our test network, 2 important things happen:

1. Nodes with the lower limit (32kb) *do reject* the 39kb transaction within their individual transaction pools, throwing `oversized data`. The node with higher limit (40kb) accepts it as expected.
2. However when it comes to importing the new chain segment, the nodes with the lower limit import the new chain segment containing the 39KB transaction *without a problem*. 

So while the mixed transction causes incongruities during pool-level validation, the block-level validation is not halted. 

Rather, the transaction gets mined when node1 mines the next block and the transaction is eventually accepted into the network (regardless of the pool-level rejection by node2 and node3). 

By placing logging statements in `blockchain.go`'s `insertChain()` logic, we can verify this to be true. 

After the 39kb transaction is sent to node 1:

Node 1:
![node 1](https://i.ibb.co/5TxpByW/node-1.png)

Node 2:

![node_2](https://i.ibb.co/bQNSRP9/node-2.png)

Node 3:
![node_3](https://i.ibb.co/HgmKVhk/node-3.png)

Comparing the timestamps between screenshots, we see that at:

- `16:16:26` - node 1 tx pool accepts, node 2&3 tx pool rejects
- `16:16:30` - new chain segment containing the transaction recognized by all nodes

As a result of the test, we believe that adding a transaction size limit does not cause network affects beyond pool-level validation; it does not halt the block mining process or cause permanent forks, and is a safe addition to Quorum's cli.

# Running the tests:

## Setup binaries
- `cd quorum/ && make all`. This quorum directory contains the txsizelimit cli flag in the PR, in addition to logging statements to help describe the logic when nodes of different `txsizelimit` interact with each other. The logging statements are primarily in `tx_pool.go` and `blockchain.go`. 


- Add bootnode and geth binaries to your path:
  - Ex: ln -s `<your_path_to>quorum/build/bin/{geth|bootnode}` to `/usr/bin/{geth|bootnode}`

## Running the test environment
- Instantiate bootnode: `bootnode -nodekey boot.key -verbosity -9 -addr :30310`
- Instantiate Node 1 (setting`txsizelimit`=40): 
  - ` /Users/Mac/Documents/quorum/build/bin/geth  --datadir node1/ --syncmode 'full' --port 30311 --rpc --rpcaddr 'localhost' --rpcport 8501 --rpcapi 'personal,db,eth,net,web3,txpool,miner' --bootnodes 'enode://291eee7eac2ea6305dd132c38d873ff950b26b46342818189b3d19934808c22a5da6f46451b87d9b45231356e5c982cdca95bb9f7dc975cfe39992ae69a34675@127.0.0.1:30310' --networkid 1515 --gasprice '1' --txsizelimit 40 -unlock '0x7a1d3415081837b3664cb909afdcb2a4a64912b2' --password node1/password.txt --mine`
- Instantiate Node 2 (setting`txsizelimit`=32): 
  - `/Users/Mac/Documents/quorum/build/bin/geth --datadir node2/ --syncmode 'full' --port 30312 --rpc --rpcaddr 'localhost' --rpcport 8502 --rpcapi 'personal,db,eth,net,web3,txpool,miner' --bootnodes 'enode://291eee7eac2ea6305dd132c38d873ff950b26b46342818189b3d19934808c22a5da6f46451b87d9b45231356e5c982cdca95bb9f7dc975cfe39992ae69a34675@127.0.0.1:30310' --networkid 1515 --gasprice '1' --txsizelimit 32 -unlock '0x86b7f283ad91e9a7a58d7087846b3622d510e4f7' --password node2/password.txt --mine`
- Instantiate Node 3 (setting`txsizelimit`=32): 
  - `/Users/Mac/Documents/quorum/build/bin/geth --datadir node3/ --syncmode 'full' --port 30313 --rpc --rpcaddr 'localhost' --rpcport 8503 --rpcapi 'personal,db,eth,net,web3,txpool,miner' --bootnodes 'enode://291eee7eac2ea6305dd132c38d873ff950b26b46342818189b3d19934808c22a5da6f46451b87d9b45231356e5c982cdca95bb9f7dc975cfe39992ae69a34675@127.0.0.1:30310' --networkid 1515 --gasprice '1' --txsizelimit 32 -unlock '0xc7f97587789713ae70e81a11b67059ff4d1dca4d' --password node3/password.txt --mine`
- Within `sendTx.go` (serves as go client that sends test txns): 
  - edit the `ipcPath` variable to point to your path's `node1/geth.ipc`
  - `go run sendTx.go -txsizelimit=<limit for testing>`

## Findings
- Setting `--txsizelimit` for nodes 1,2,3 = [40,32,32]:
  - **Finding #1**: Sending tx size of 35KB to Node 1 results in undisturbed mining on all 3 nodes
  - **Finding #2**: Sending tx size of >40KB to Node 1 results in Oversided Data Error
- Setting `--txsizelimit` for nodes 1,2,3 = [40,40,32]:
  - **Finding #3**: Node 1 and 2 instantiated first and allowed to mine a few 35KB transactions. Then node 3 joins. When 35KB transaction is sent to node 1, then the result is that node 3 forms a side fork.
