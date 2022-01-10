package main

import ("testing")

func TestshowBlocks(t *testing.T) {
	sampleBlock := []string{"0xd260eb", "0xd260ee","0xd260ef","0xd260f1"}
	var emptyRespose JPCResponse
	for _, v := range sampleBlock {
		BlockStore[v] = emptyRespose
	}

	sb, err := showBlocks()
	_ = FailOnError(err, "showBlocks")

	for _, v := range *sb {
		if v != sampleBlock[0] && v != sampleBlock[1] && v != sampleBlock[2] && v != sampleBlock[3] && v != sampleBlock[4]{
			t.Error("Block data did not match ones shown")
		}
	}

}