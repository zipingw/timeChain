package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	bptreeS "timeChain/bptreeS"
	//"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	gateway "timeChain/gateway"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	//"github.com/hyperledger/fabric-protos-go/common"
	//pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	//更新索引树时间
	"crypto/rand"
	"math/big"

	mrand "math/rand"
	//"github.com/hyperledger/fabric/protoutil"
	//"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"
)

var FormatISOTime = "2006-01-02T15:04:05"
var TimeInterval = 10

func main() {
	/*
		// 连接sensor
		sensor := connectToSensor()
		// 构建空B+树
		bpt := bptreeS.NewBPTree(5)
		// 主动接收sensor发送过来的数据

		ticker := time.NewTicker(TimeInterval * time.Second)
		defer ticker.Stop()
		for {
			select {
				case <- ticker.C:{
					data := sensor.get()
					// 此批数据Data[]构建ValidationTree,返回三个信息
					vTree := new ValidationTree()
					hashRoot, hashLeaf, storeNodes := vTree.Digest()
					// 将这三个信息通过区块链上链,返回TxID和BlockID
					txID, blockID := SubmitTxGetAddr(hashRoot, hashLeaf, storeNodes)
					// 将该批数据与其交易地址作为B+树的键值对存入B+树中
					addNodeToBptree(bpt, data, txID, blockID)
					// 发送必要信息给Data Storer
					sendMsgToStorer(data, hashRoot, hashLeaf, storeNodes)
				}
			}
		}
	*/
	//test_T_search_tree_Q(50, 1000, 1000)
	//test_T_update_tree(5, 1000, 1000)
	test_trans_onchain_T_find_store_node(8)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func test_trans_onchain_T_find_store_node(numTx int) {
	var submitTime []time.Duration
	var readTime []time.Duration
	var submitTotalTime time.Duration
	var readTotalTime time.Duration
	var cnt = numTx
	hashRoot, hashLeaf, storeNodes := randString(58), randString(1), randString(1)
	for i := 0; i < cnt; i++ {
		_, _, txTime, rTime := SubmitTxGetTxAddr(hashRoot, hashLeaf, storeNodes)
		submitTime = append(submitTime, txTime)
		readTime = append(readTime, rTime)
		submitTotalTime += txTime
		readTotalTime += rTime
		fmt.Printf("submit time: %v \n", txTime)
		fmt.Printf("read time: %v \n", rTime)
	}
	fmt.Printf("Total time for submit: ", submitTotalTime)
	fmt.Printf("Average time for submit", submitTotalTime/time.Duration(int64(cnt)))
	fmt.Printf("Total time for read: ", readTotalTime)
	fmt.Printf("Average time for read", readTotalTime/time.Duration(int64(cnt)))
}
func test_T_update_tree(width int, numNode int, testNum int) {
	bpt := bptreeS.NewBPTree(width)
	var value, id string
	var updateBptTime []time.Duration
	var totalTime time.Duration
	var all_nodes []string
	var cnt = numNode
	for i := 0; i < cnt; i++ {
		value = randString(4)
		id = randString(4)
		bpt.Set(id, value)
		all_nodes = append(all_nodes, value)
	}
	for i := 0; i < testNum; i++ {
		value = randString(4)
		id = randString(4)
		start := time.Now()
		bpt.Set(id, value)
		elapsed := time.Since(start)
		totalTime += elapsed
		updateBptTime = append(updateBptTime, elapsed)
		fmt.Printf("Insertion %d took %v\n", i+1, elapsed)
		bpt.Remove(id)
	}
	fmt.Printf("Total time for insertions:", totalTime)
	fmt.Println()
	fmt.Printf("Average time for insertions:", totalTime/time.Duration(int64(testNum)))
}
func test_T_search_tree_Q(width int, numNode int, testNum int) {
	bpt := bptreeS.NewBPTree(width)
	var value, id string
	var updateBptTime []time.Duration
	var totalTime time.Duration
	var all_nodes []string
	var cnt = numNode
	for i := 0; i < cnt; i++ {
		value = randString(4)
		id = randString(4)
		//start := time.Now()
		bpt.Set(id, value)
		all_nodes = append(all_nodes, id)
		//elapsed := time.Since(start)
		//totalTime += elapsed
		//updateBptTime = append(updateBptTime, elapsed)
		//fmt.Printf("Insertion %d took %v\n", i+1, elapsed)
	}
	mrand.Seed(time.Now().UnixNano())
	min := 0
	max := len(all_nodes)
	for i := 0; i < testNum; i++ {
		randomNumber := mrand.Intn(max-min) + min
		start := time.Now()
		bpt.Get(all_nodes[randomNumber])
		elapsed := time.Since(start)
		totalTime += elapsed
		updateBptTime = append(updateBptTime, elapsed)
		fmt.Printf("Query %d took %v\n", randomNumber, elapsed)
	}
	fmt.Printf("Total time for query:", totalTime)
	fmt.Println()
	fmt.Printf("Average time for query:", totalTime/time.Duration(int64(testNum)))

}
func randString(length int) string {
	bytes := make([]byte, length)
	for i := range bytes {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		bytes[i] = letterBytes[n.Int64()]
	}
	return string(bytes)
}
func addNodeToBptree(bpt *bptreeS.BPTree, data []string, txID string, blockID string) {
	for _, value := range data {
		bpt.Set(value, txID+blockID)
	}
}
func SubmitTxGetTxAddr(hashRoot string, hashLeaf string, storeNodes string) (string, string, time.Duration, time.Duration) {
	gw, contract := connectToCC("vTree")
	defer disconnectToCC(gw)
	err, txID, txTime, readTime := vTreeCC(contract, hashRoot, hashLeaf, storeNodes)
	if err != nil {
		log.Fatalf("Failed : %v", err)
	}
	// 根据result中的信息获取 TxID 和 BlockID
	// txID, blockID := result.txID, result.blockID
	blockID := "blk"
	getBlockNumberByTxId(txID, gw)

	return txID, blockID, txTime, readTime
}

func getBlockNumberByTxId(txId string, gw *gateway.Gateway) {
	// 根据txId构建出 fab.Transacation
	//sdk, err := fabsdk.New(config.FromFile("/home/zpwang/go/src/github.com/zipingw/fabric-samples/test-network/networkConfig.yaml"))
	//sdk, err := fabsdk.New(config.FromFile("/home/zpwang/go/src/github.com/zipingw/fabric-samples/test-network/configtx/configtx.yaml"))
	//sdk, err := fabsdk.New(config.FromFile("/home/zpwang/go/src/github.com/zipingw/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/connection-org1.yaml"))
	sdk, err := fabsdk.New(config.FromFile("/home/zpwang/go/src/github.com/zipingw/fabric-samples/test-network/compose/compose-test-net.yaml"))
	if err != nil {
		log.Panicf("failed to create fabric sdk: %s", err)
	}
	var channelProvider context.ChannelProvider
	channelProvider = sdk.ChannelContext("mychannel", fabsdk.WithIdentity(gw.options.Identity), fabsdk.WithOrg(gw.getOrg()))
	fmt.Println("-----------------------")
	c, err := ledger.New(channelProvider)
	
	if err != nil {
		fmt.Println("failed to create client")
		log.Println("Error:", err)
	}
	// block类型为 *common.Block

	//Header               *BlockHeader   `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	//Data                 *BlockData     `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	//Metadata             *BlockMetadata `protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`

	// 得到block，可以从中解析出Tx信息，但是无法得到BlockNumber
	            c.QueryBlockByTxID(fab.TransactionID(txId))
	// processedTransaction, _  c.QueryTransaction(fab.TransactionID(txId))
	// 这里payload 类型 common.Payload
	// payload, _ := protoutil.UnmarshalPayload(processedTransaction.TransactionEnvelope.Payload)
	// 获得block number
}

/*
	func GetTransactionInfoFromTxData(data []type, needArgs bool)(*TransactionDetail, error){
		// 获取common.Envelope
		env, err := protoutil.GetEnvelopeFromBlock(data)
		if err != nil {
			return nil, errors.Wrap(err, "error extracting Envelope from block")
		}
		if env == nil {
			return nil, errors.New("nil envelope")
		}
		// 获取peer.ChaincodeActionPayload, peer.ChaincodeAction
		payloadbytes := env.GetPayload()
		payload := &common.Payload{}
		proto.Unmarshal(payloadbytes, payload)
		payload.GetData()
		// 获取pb.ChaincodeProposalPayload
		propPayload := &pb.ChaincodeProposalPayload{}
		if err := proto.Unmarshal(ccPayload.ChaincodeProposalPayload, propPayload); err != nil {
			return nil, errors.Wrap(err, "error extracting ChannelHeader from payload")
		}
		invokeSpec := &pb.ChaincodeInvocationSpec{}
		err = proto.Unmarshal(propPayload.Input, invokeSpec)
		if err != nil {
			return nil, errors.Wrap(err, "error sxtracting ChannelHander from payload")
		}
	}
*/
func vTreeCC(contract *gateway.Contract, hRoot string, hLeaf string, sNodes string) (error, string, time.Duration, time.Duration) {
	log.Println("==> Submit Transaction: Update (hashRoot, hashLeaf, storeNodes)")

	var txTime time.Duration
	start := time.Now()
	result, err := contract.SubmitTransaction("Update", hRoot, hLeaf, sNodes)
	txTime = time.Since(start)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}
	if len(result) == 0 {
		log.Println("Result is empty")
	} else {
		log.Println("TxId:", string(result))
	}
	txID := string(result)

	key := "fakeKey"
	start = time.Now()
	contract.EvaluateTransaction("ReadTime", key)
	readTime := time.Since(start)
	fmt.Printf("Read World State Time :%v \n", readTime)

	return err, txID, txTime, readTime
}
func connectToCC(ccn string) (*gateway.Gateway, *gateway.Contract) {
	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environment variable: %v", err)
	}

	walletPath := "wallet"
	// remove any existing wallet from prior runs
	os.RemoveAll(walletPath)
	wallet, err := gateway.NewFileSystemWallet(walletPath)
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}
	ccpPath := filepath.Join(
		"/home",
		"zpwang",
		"go",
		"src",
		"github.com",
		"zipingw",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
		//"fabric-ca-client-config.yaml",
	)
	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	// connect to channel
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	log.Println("--> Connecting to channel:", channelName)
	network, err := gw.GetNetwork(channelName)
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}
	// chaincodeName := "basic"
	chaincodeName := ccn

	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}
	contract := network.GetContract(chaincodeName)
	return gw, contract
}

func disconnectToCC(gw *gateway.Gateway) {
	gw.Close()
}
func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"/home",
		"zpwang",
		"go",
		"src",
		"github.com",
		"zipingw",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	//certPath := filepath.Join(credPath, "signcerts", "User1@org1.example.com-cert.pem")
	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := os.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := os.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := os.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}
