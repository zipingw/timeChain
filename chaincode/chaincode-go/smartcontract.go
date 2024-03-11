package main

import (
	"fmt"
	"encoding/json"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
//	"time"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

type ValidationTree struct {
	HashRoot   string `json:"HashRoot"`
	HashLeaf   string `json:"HashLeaf"`
	StoreNodes string `json:"StoreNodes"`
}

// Create record
func (s *SmartContract) Create(ctx contractapi.TransactionContextInterface, hashRoot string, hashLeaf string, storeNodes string) string {
	vTree := ValidationTree{
		HashRoot:   hashRoot,
		HashLeaf:   hashLeaf,
		StoreNodes: storeNodes,
	}
	vTreeJSON, _ := json.Marshal(vTree)
	txId := ctx.GetStub().GetTxID()
	log.Println("!ok")
	log.Printf("Get TxId : %v\n", txId)
	ctx.GetStub().PutState("fakeKey", vTreeJSON)

	ctx.GetStub().SetEvent("Update", []byte(vTreeJSON))
	//多次PutState可以绑定相关关键信息
	return txId
}

func (s *SmartContract) Update(ctx contractapi.TransactionContextInterface, hashRoot string, hashLeaf string, storeNodes string) string {
	exists, _ := s.Exists(ctx, "fakeKey")
	if !exists {
		return s.Create(ctx, hashRoot, hashLeaf, storeNodes)
	}
	vTree := ValidationTree{
		HashRoot:   hashRoot,
		HashLeaf:   hashLeaf,
		StoreNodes: storeNodes,
	}
	vTreeJSON, _ := json.Marshal(vTree)
	log.Println("!ok")
	txId := ctx.GetStub().GetTxID()
	log.Printf("Get TxId : %v\n", txId)
	//多次PutState可以绑定相关关键信息
	ctx.GetStub().PutState("fakeKey", vTreeJSON)
	ctx.GetStub().SetEvent("Update", []byte(vTreeJSON))
	return txId
}
func (s *SmartContract) Exists(ctx contractapi.TransactionContextInterface, key string) (bool, error) {
        vTreeJSON, err := ctx.GetStub().GetState(key)
        if err != nil {
                return false, fmt.Errorf("failed to read from world state: %v", err)
        }
	return vTreeJSON != nil, nil
}
func (s *SmartContract) Read(ctx contractapi.TransactionContextInterface, key string) (*ValidationTree, error) {
	vTreeJSON, err := ctx.GetStub().GetState(key)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if vTreeJSON == nil {
		return nil, fmt.Errorf("the vtree %s does not exist", key)
	}
	var vTree ValidationTree
	err = json.Unmarshal(vTreeJSON, &vTree)
	if err != nil {
		return nil, err
	}
        return &vTree, nil
}
//func (s *SmartContract) GetTxId(ctx contractapi.TransactionContextInterface) (string)
func main() {
	recordChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating ipfscc-basic chaincode: %v", err)
	}
	if err := recordChaincode.Start(); err != nil {
		log.Panicf("Error starting ipfscc-basic chaincode: %v", err)
	}
}
