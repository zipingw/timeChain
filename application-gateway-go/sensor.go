package main

import (
	"bytes"
	//"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	//"errors"
	"context"
	"errors"
	"fmt"
	"log"
	"io/ioutil"
	"math/big"
	mrand "math/rand"
	"os"
	"path"
	"sync"
	"time"

	"timeChain/bptreeS"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	//"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	//"google.golang.org/grpc/status"
)

const (
	mspID        = "Org1MSP"
	cryptoPath   = "/home/zpwang/go/src/github.com/zipingw/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com"
	certPath     = cryptoPath + "/users/User1@org1.example.com/msp/signcerts/cert.pem"
	keyPath      = cryptoPath + "/users/User1@org1.example.com/msp/keystore/"
	tlsCertPath  = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint = "localhost:7051"
	gatewayPeer  = "peer0.org1.example.com"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

//var mutex sync.Mutex

func main() {
	var ccn = "vTree"
	// The gRPC client connection should be shared by all Gateway connections to this endpoint
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gw.Close()

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := ccn
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	channelName := "mychannel"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	// Context used for event listening
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//startChaincodeEventListening(ctx, network)
	//hashRoot, hashLeaf, storeNodes := randString(58), randString(1), randString(1)
	//txID, txTime := Update(contract, hashRoot, hashLeaf, storeNodes)
	//fmt.Printf("tx on chain : txID: %v , timecost: %v\n", txID, txTime)

	//test_trans_onchain_T_find_store_node(contract, 4)
	//test_T_query_onchain_prefix(ctx, contract, network, 10000)
	test_T_query_onchain(ctx, contract, network, 10000)
}

func replayChaincodeEvents(ctx context.Context, network *client.Network, chaincodeName string, checkpoint client.Checkpoint) {
	fmt.Println("\n*** Start chaincode event replay")

	//events, err := network.ChaincodeEvents(ctx, chaincodeName, client.WithStartBlock(startBlock))
	events, err := network.ChaincodeEvents(ctx, chaincodeName, client.WithCheckpoint(checkpoint))
	if err != nil {
		panic(fmt.Errorf("failed to start chaincode event listening: %w", err))
	}

	for {
		select {
		case <-time.After(10 * time.Second):
			panic(errors.New("timeout waiting for event replay"))

		case event := <-events:
			asset := formatJSON(event.Payload)
			fmt.Printf("\n<-- Chaincode event replayed: %s - %s\n", event.EventName, asset)
			return
		}
	}
}
func test_T_query_onchain_prefix(ctx context.Context, contract *client.Contract, network *client.Network, numTx int){
	// 将 numTx 笔交易上链，记录下event
	chaincodeName := "vTree"
	events, err := network.ChaincodeEvents(ctx, chaincodeName)
	if err != nil {
		panic(fmt.Errorf("failed to start chaincode event listening: %w", err))
	}
	for i := 0; i < numTx; i++ {
		hashRoot, hashLeaf, storeNodes := randString(58), randString(1), randString(1)
		txID, txTime := Update(contract, hashRoot, hashLeaf, storeNodes)
		fmt.Printf("tx on chain : txID: %v , timecost: %v\n", txID, txTime)
	}
	/*
		event := &ChaincodeEvent{
			BlockNumber: 1,
			ChaincodeName: "vTree",
			EventName:
			Payload:
			TransactionID:
		}*/
	var eventSet []*client.ChaincodeEvent
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		cnt := 0
		for event := range events {
			fmt.Println("before append")
			eventSet = append(eventSet, event)
			fmt.Print(eventSet)
			fmt.Print(len(eventSet))
			fmt.Println("after append")
			args := formatJSON(event.Payload)
			fmt.Printf("\n<-- Chaincode event received: %s - %s\n", event.EventName, args)
			cnt += 1
			if cnt == numTx {
				return
			}
		}
	}()
	wg.Wait()
	fmt.Println("=============>  all events printed!")

	fmt.Println(len(eventSet))
	jsonEventSet, err := json.Marshal(eventSet)
	if err != nil {
		fmt.Println("序列化失败：", err)
		return
	}
	err = ioutil.WriteFile("./eventSet.json", jsonEventSet, 0644)
	if err != nil {
		fmt.Println("写入文件失败：", err)
		return
	}
	fmt.Println("切片对象成功序列化到文件 eventSet.json 中.")
}
func test_T_query_onchain(ctx context.Context, contract *client.Contract, network *client.Network, testNum int) {
	var eventSet []*client.ChaincodeEvent
	chaincodeName := "vTree"
	fileData, err := ioutil.ReadFile("./eventSet.json")
	if err != nil {
		fmt.Println("读取文件失败：", err)
		return
	}
	err = json.Unmarshal(fileData, &eventSet)
	if err != nil {
		fmt.Println("解码JSON失败：", err)
		return
	}
	fmt.Printf("加载序列化文件eventSet成功,该切片长度为%d", len(eventSet))
	// 随机抽取 100个 测试时间
	var queryTime []time.Duration
        var totalTime time.Duration
	mrand.Seed(time.Now().UnixNano())
        min := 0
        max := len(eventSet) - 1
        for i := 0; i < testNum; i++ {
                randomNumber := mrand.Intn(max-min) + min
 	
		checkpointer := new(client.InMemoryCheckpointer)
		checkpointer.CheckpointChaincodeEvent(eventSet[randomNumber])
                start := time.Now()
		replayChaincodeEvents(ctx, network, chaincodeName, checkpointer)
		elapsed := time.Since(start)
                
		totalTime += elapsed
                queryTime = append(queryTime, elapsed)
                fmt.Printf("Query %d took %v\n", randomNumber, elapsed)
        }
	fmt.Printf("Total time for query:", totalTime)
        fmt.Println()
        fmt.Printf("Average time for query:", totalTime/time.Duration(int64(testNum)))

}
func startChaincodeEventListening(ctx context.Context, network *client.Network) {
	fmt.Println("\n*** Start chaincode event listening")

	events, err := network.ChaincodeEvents(ctx, "vTree")
	if err != nil {
		panic(fmt.Errorf("failed to start chaincode event listening: %w", err))
	}

	go func() {
		for event := range events {
			//fmt.Print(event)
			//fmt.Println("=========")
			asset := formatJSON(event.Payload)
			fmt.Printf("\n<-- Chaincode event received: %s - %s\n", event.EventName, asset)
		}
	}()
}
func test_trans_onchain_T_find_store_node(contract *client.Contract, numTx int) {
	var submitTime []time.Duration
	var readTime []time.Duration
	var submitTotalTime time.Duration
	var readTotalTime time.Duration
	var cnt = numTx
	hashRoot, hashLeaf, storeNodes := randString(58), randString(1), randString(1)
	for i := 0; i < cnt; i++ {
		_, txTime := Update(contract, hashRoot, hashLeaf, storeNodes)
		rTime := Read(contract)
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

func Update(contract *client.Contract, hashRoot string, hashLeaf string, storeNodes string) (string, time.Duration) {
	fmt.Println("\n--> Submit Transaction: Update arguments \n")

	start := time.Now()
	result, err := contract.SubmitTransaction("Update", hashRoot, hashLeaf, storeNodes)
	updateTime := time.Since(start)

	if err != nil {
		panic(fmt.Errorf("Failed to submit transaction: %w", err))
	}
	fmt.Printf("*** Transaction committed successfully\n")
	if len(result) == 0 {
		log.Println("Result is empty")
	} else {
		log.Println("TxId:", string(result))
	}
	txID := string(result)
	return txID, updateTime
}

// Format JSON data
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}

func Read(contract *client.Contract) time.Duration {
	fmt.Printf("\n--> Evaluate Transaction: Read, function returns validation information.\n")

	start := time.Now()
	evaluateResult, err := contract.EvaluateTransaction("Read", "fakeKey")
	readTime := time.Since(start)
	log.Printf("Read World State Time :%v \n", readTime)

	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("*** Result:%s\n", result)
	return readTime
}

func randString(length int) string {
	bytes := make([]byte, length)
	for i := range bytes {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		bytes[i] = letterBytes[n.Int64()]
	}
	return string(bytes)
}

// width是B+树的宽度（20为合适值）,numNode为测试时插入的节点总数
// 测试逻辑为先插入numNode个节点， 再反复进行随机的插入，每插入一个就删除一个
// testNum为反复测试的次数
func test_T_update_tree(width int, numNode int, testNum int) {
	bpt := bptreeS.NewBPTree(width)
	var value, id string
	var updateBptTime []time.Duration
	var totalTime time.Duration
	var all_nodes []string
	for i := 0; i < numNode; i++ {
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
		// 插入这个地方会报错
		bpt.Remove(id)
	}
	fmt.Printf("Total time for insertions:", totalTime)
	fmt.Println()
	fmt.Printf("Average time for insertions:", totalTime/time.Duration(int64(testNum)))
}

// 测试逻辑为先插入numNode个数，并记录下节点的id
// 随机选择树中的节点id，进行testNum次查询所用时间
func test_T_search_tree_Q(width int, numNode int, testNum int) {
	bpt := bptreeS.NewBPTree(width)
	var value, id string
	var updateBptTime []time.Duration
	var totalTime time.Duration
	var all_nodes []string
	for i := 0; i < numNode; i++ {
		value = randString(4)
		id = randString(4)
		bpt.Set(id, value)
		all_nodes = append(all_nodes, id)
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

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() *grpc.ClientConn {
	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
	certificate, err := loadCertificate(certPath)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() identity.Sign {
	files, err := os.ReadDir(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key directory: %w", err))
	}
	privateKeyPEM, err := os.ReadFile(path.Join(keyPath, files[0].Name()))

	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}
