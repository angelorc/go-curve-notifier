package query

import (
	"context"
	"github.com/cometbft/cometbft/libs/json"
	"google.golang.org/grpc"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type QueryClient struct {
	Authz authz.QueryClient
	Bank  banktypes.QueryClient
	Wasm  wasmtypes.QueryClient
}

func NewQueryClient(conn *grpc.ClientConn) (*QueryClient, error) {
	client := &QueryClient{
		Authz: authz.NewQueryClient(conn),
		Bank:  banktypes.NewQueryClient(conn),
		Wasm:  wasmtypes.NewQueryClient(conn),
	}
	return client, nil
}

type QueryNftContractInfoRequest struct {
	ContractInfo struct{} `json:"contract_info"`
}

type QueryNftContractInfoResponse struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Uri    string `json:"uri"`
}

func NewQueryNftContractInfoRequest() *QueryNftContractInfoRequest {
	return &QueryNftContractInfoRequest{}
}

func (qc *QueryClient) NftContractInfo(contractAddr string) (*QueryNftContractInfoResponse, error) {
	queryData, err := json.Marshal(NewQueryNftContractInfoRequest())
	if err != nil {
		return nil, err
	}

	resp, err := qc.Wasm.SmartContractState(context.Background(), &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: queryData,
	})
	if err != nil {
		return nil, err
	}

	var contractInfo QueryNftContractInfoResponse
	err = json.Unmarshal(resp.Data, &contractInfo)
	if err != nil {
		return nil, err
	}

	return &contractInfo, nil
}
