package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"time"
)

func main() {
	txsizelimit := flag.Int("txsizelimit", 32, "transaction size limit")
	flag.Parse()

	ipcPath := "/Users/Mac/Documents/kaleidotest/quorumtest/node1/geth.ipc"
	client, err := ethclient.Dial(ipcPath)
	if err != nil {
		log.Fatalf("failed to connect to eth client: %v", err)
	}

	rawTransaction(client, txsizelimit)
}

const key = `{"address":"7a1d3415081837b3664cb909afdcb2a4a64912b2","crypto":{"cipher":"aes-128-ctr","ciphertext":"ea14ec4797a0a0bcd823a710e798309b717198029356a47c1e0e4c43cbd8e593","cipherparams":{"iv":"0e0c2e3e7cf9ae04857c907e6f4c362b"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"fca87c435129b58d51e668f49712e93251e973dfb2c3428e2d80a9e016eb6c62"},"mac":"5ea24dc3713aa20110dded400b93d2fd940c6cca806d492f14648cebdbd215b3"},"id":"367b544d-bf71-425c-bb73-77437b1f0bbb","version":3}`

func rawTransaction(client *ethclient.Client, txsizelimit *int) {
	d := time.Now().Add(5 * time.Minute)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	unlockedKey, err := keystore.DecryptKey([]byte(key), "pass1")
	nonce, _ := client.NonceAt(ctx, unlockedKey.Address, nil)

	if err != nil {
		fmt.Println("Wrong password")
	} else {
		data := make([]byte, (*txsizelimit*1024)+1)
		tx := types.NewTransaction(nonce,
			common.Address{},
			big.NewInt(100),
			uint64(1000000),
			big.NewInt(0),
			data,
		)
		signTx, _ := types.SignTx(tx, types.HomesteadSigner{}, unlockedKey.PrivateKey)
		err = client.SendTransaction(ctx, signTx)

		if err != nil {
			fmt.Println(err, nonce)
		} else {
			fmt.Println(tx.Hash().String())
		}
	}
}
