package main 

import(
	"testing"
	"github.com/ethereum/go-ethereum/rpc"
)

func Test_getLastMainnetBlocks(t *testing.T) {
	sampleHex := "0xd260eb"
	client, err := rpc.Dial(InfuraLink)
	_ = FailOnError(err, "rpc.Dial Failed to connect to Ethereum Server")
	getblock := getLastMainnetBlocks(client, 1)
	if len(getblock.Number) != len(sampleHex) {
		t.Error("Received Block number does not match sample")
	}
}