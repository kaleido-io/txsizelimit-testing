# txSizeLimitTesting
Testing the network effects of the transaction size limit flag we've submitted to Quorum: https://github.com/jpmorganchase/quorum/pull/575

These tests are crucial for our understanding before further PR to Quorum and/or geth main.

## Setup binaries
- Clone our Quorum with the PR: https://github.com/kaleido-io/quorum/tree/photic-tx-size-limit
- `make all` in the Quorum dir
- Add bootnode and geth binaries to your path:
  - Ex: ln -s `<your_path_to>quorum/build/bin/{geth|bootnode}` to `/usr/bin/{geth|bootnode}`

## Running the test environment
- Instantiate bootnode: `bootnode -nodekey boot.key -verbosity -9 -addr :30310`
- Instantiate Node 1: 
  - ` /Users/Mac/Documents/quorum/build/bin/geth  --datadir node1/ --syncmode 'full' --port 30311 --rpc --rpcaddr 'localhost' --rpcport 8501 --rpcapi 'personal,db,eth,net,web3,txpool,miner' --bootnodes 'enode://291eee7eac2ea6305dd132c38d873ff950b26b46342818189b3d19934808c22a5da6f46451b87d9b45231356e5c982cdca95bb9f7dc975cfe39992ae69a34675@127.0.0.1:30310' --networkid 1515 --gasprice '1' --txsizelimit 40 -unlock '0x7a1d3415081837b3664cb909afdcb2a4a64912b2' --password node1/password.txt --mine`
- Instantiate Node 2: 
  - `/Users/Mac/Documents/quorum/build/bin/geth --datadir node2/ --syncmode 'full' --port 30312 --rpc --rpcaddr 'localhost' --rpcport 8502 --rpcapi 'personal,db,eth,net,web3,txpool,miner' --bootnodes 'enode://291eee7eac2ea6305dd132c38d873ff950b26b46342818189b3d19934808c22a5da6f46451b87d9b45231356e5c982cdca95bb9f7dc975cfe39992ae69a34675@127.0.0.1:30310' --networkid 1515 --gasprice '1' --txsizelimit 32 -unlock '0x86b7f283ad91e9a7a58d7087846b3622d510e4f7' --password node2/password.txt --mine`
- Instantiate Node 3: 
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
