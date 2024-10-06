package ipfs

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	ipfs_rpc "github.com/ipfs/kubo/client/rpc"
)

type IpfsHandler struct {
	ipfsClient *ipfs_rpc.HttpApi
}

func New() *IpfsHandler {
	client, err := ipfs_rpc.NewLocalApi()
	if err != nil {
		log.Fatalf("failed to create IPFS RPC client: %v", err)
	}

	return &IpfsHandler{
		ipfsClient: client,
	}
}

func (client *IpfsHandler) FetchFromIPFPLocal(cid string) ([]byte, error) {
	if client.ipfsClient == nil {
		return nil, fmt.Errorf("IPFS client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	req := client.ipfsClient.Request("cat", cid)

	res, err := req.Send(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to perform 'cat' request: %w", err)
	}
	defer res.Output.Close()

	data, err := io.ReadAll(res.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to read data from 'cat' response: %w", err)
	}

	return data, nil
}
