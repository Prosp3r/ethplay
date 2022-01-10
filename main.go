package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gorilla/mux"
	"github.com/jeffprestes/ethclient/common/types"
)

var (
	InfuraLink      = "https://mainnet.infura.io/v3/4b0fe94094e047ffa6292fc8065e42b8"
	BlockStore      = make(map[string]JPCResponse)
	ReceiptStore    = make(map[string]*types.Receipt)
	mutx            = &sync.Mutex{}
	apiport         = ":8080"
	maxStoredBlocks int
)

type JPCResponse struct {
	Jsonrpc      string      `json:"jsonrpc"`
	ID           int         `json:"id"`
	Transactions Transaction `json:"result"`
}

type Transaction struct {
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	From             string `json:"from"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	Nonce            string `json:"nonce"`
	R                string `json:"r"`
	S                string `json:"s"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
	Type             string `json:"type"`
	V                string `json:"v"`
	Value            string `json:"value"`
}

type Block struct {
	Number     string
	Hash       common.Hash
	ParentHash common.Hash
}

var blocks []Block

type JPCReceipts struct {
	Jsonrpc string  `json:"jsonrpc"`
	ID      int     `json:"id"`
	Result  TResult `json:"result"`
}

type TResult struct {
	BlockHash        common.Hash    `json:"blockhash,omitempty"`
	BlockNumber      *uint64        `json:"blocknumber,omitempty"`
	From             common.Address `json:"from,omitempty"`
	Gas              *uint64        `json:"gas,omitempty"`
	GasPrice         *uint64        `json:"gasprice,omitempty"`
	Hash             common.Hash    `json:"hash,omitempty"`
	Input            string         `json:"input,omitempty"`
	Nonce            *uint64        `json:"nonce,omitempty"`
	R                *uint64        `json:"r,omitempty"`
	S                *uint64        `json:"s,omitempty"`
	To               common.Address `json:"to,omitempty"`
	TransactionIndex *uint64        `json:"transactionIndex,omitempty"`
	Type             string         `json:"type,omitempty"`
	V                *uint64        `json:"v,omitempty"`
	Value            *uint64        `json:"value,omitempty"`
}

//FailOnError - A uniform yet detailed error Utility. Useful for easy error logging
func FailOnError(err error, note string) bool {
	if err != nil {
		fmt.Printf("Failed trying to %v with ERROR : %v \n", note, err)
		return true
	}
	return false
}

func main() {
	//Scan for user inputs
	client, err := rpc.Dial(InfuraLink)
	_ = FailOnError(err, "Dialing ETHClient")
	defer client.Close()

	go getLastMainnetBlocks(client, 11)
	startApiServer()
}

func getLastMainnetBlocks(c *rpc.Client, maxblocks int) *Block {
	var lb Block

	for n := 0; n < maxblocks; {

		err := c.Call(&lb, "eth_getBlockByNumber", "latest", true)
		_ = FailOnError(err, "eth_getBlockByNumber")

		blocks = append(blocks, lb)
		mutx.Lock()
		if val, ok := BlockStore[lb.Number]; ok {
			//Index exists
			fmt.Sprintf("Block :%v alread exists \n\n", val.Transactions.BlockNumber)
			mutx.Unlock()
		} else {
			var response JPCResponse
			BlockStore[lb.Number] = response
			mutx.Unlock()
			//showBlocks()
			go func() {
				trx, err := getBlockTransactionCurl(lb.Number)
				if err != nil {
					fmt.Sprintf("Failed to getBlockTransactionCurl : %v\n", err)
				}
				// _ = FailOnError(err, "getBlockTransactionCurl")

				trxHash := common.HexToHash(trx.Transactions.Hash)
				_, err = getMainnetTransactionReceipt(context.Background(), c, trxHash)
				_ = FailOnError(err, "getBlockTransactionCurl")
			}()
			n++
		}
		time.Sleep(time.Second * 5)
	}
	return &lb
}

func showBlocks() (*[]string, error) {
	var result []string
	mutx.Lock()
	for i := range BlockStore {
		result = append(result, i)
	}
	mutx.Unlock()
	return &result, nil
}

func getAllTransactions() (*map[string]JPCResponse, error) {
	var result map[string]JPCResponse
	mutx.Lock()
	result = BlockStore
	mutx.Unlock()
	return &result, nil
}

func getAllReceipts() (*map[string]*types.Receipt, error) {
	var result map[string]*types.Receipt
	mutx.Lock()
	defer mutx.Unlock()
	result = ReceiptStore
	return &result, nil
}

func getMainnetTransactionReceipt(ctxt context.Context, c *rpc.Client, transactionHash common.Hash) (*types.Receipt, error) {
	var receipt *types.Receipt
	err := c.CallContext(ctxt, &receipt, "eth_getTransactionReceipt", transactionHash)
	if err == nil {
		if receipt == nil {
			return nil, errors.New("Missing Receipt")
		}
	}
	ReceiptStore[receipt.TxHash.String()] = receipt
	return receipt, err
}

func getBlockTransactionCurl(blocknumber string) (*JPCResponse, error) {
	payload := strings.NewReader("{\"jsonrpc\":\"2.0\",\"method\":\"eth_getTransactionByBlockNumberAndIndex\",\"params\": [\"" + blocknumber + "\",\"0x0\"],\"id\":1}")

	client := &http.Client{}
	req, err := http.NewRequest("POST", InfuraLink, payload)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	e := FailOnError(err, "Client.Do(req)")
	if e == true {
		res.Body.Close()
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	_ = FailOnError(err, "ReadingBody")
	var response JPCResponse

	err = json.Unmarshal(body, &response)

	mutx.Lock()
	BlockStore[blocknumber] = response
	mutx.Unlock()

	return &response, nil
}

//router & end-points
func startApiServer() {
	router := mux.NewRouter()

	router.HandleFunc("/", homepage).Methods("GET")
	router.HandleFunc("/blocks", getBlocks).Methods("GET")
	router.HandleFunc("/transactions", getTransactions).Methods("GET")
	router.HandleFunc("/receipts", getReceipts).Methods("GET")

	fmt.Printf("Starting server at port %s \n", apiport)
	log.Fatal(http.ListenAndServe(apiport, router))
}

//handlers
func homepage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "You've reached my Ethereum Load test /homepage \n\n Try the following end points \n /blocks to view a list of blocks fetched from the mainnet. Updates ever five seconds\n /transactions to view the JSON formated list of transactions carried out on the fetched block numbers\n /receipts to view a JSON formated list of receipts \n ")
}

func getBlocks(w http.ResponseWriter, r *http.Request) {
	blocks, err := showBlocks()
	e := FailOnError(err, "Failed to get blocks")
	if e == true {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(blocks)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blocks)
}

func getTransactions(w http.ResponseWriter, r *http.Request) {
	transactions, err := getAllTransactions()
	e := FailOnError(err, "Failed to get transactions")
	if e == true {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("Failed to get transactions")
	}

	tranx, err := json.MarshalIndent(transactions, "", " ")
	_ = FailOnError(err, "json.MarshalIndent-transactions")

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(tranx))
}

func getReceipts(w http.ResponseWriter, r *http.Request) {
	receipts, err := getAllReceipts()
	e := FailOnError(err, "Failed to get receipts")
	if e == true {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("Failed to get receipts")
	}

	receiptx, err := json.MarshalIndent(receipts, "", " ")
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(receiptx))
}
