// Copyright 2015 The NeuralChain Authors
// This file is part of the NeuralChain library .
//
// The NeuralChain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The NeuralChain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the NeuralChain library . If not, see <http://www.gnu.org/licenses/>.

// Package ethapi implements the general NeuralChain API functions.
package neutapi

import (
	"context"
	"math/big"

	"github.com/lvbin2012/NeuralChain/accounts"
	"github.com/lvbin2012/NeuralChain/common"
	"github.com/lvbin2012/NeuralChain/core"
	"github.com/lvbin2012/NeuralChain/core/state"
	"github.com/lvbin2012/NeuralChain/core/types"
	"github.com/lvbin2012/NeuralChain/core/vm"
	"github.com/lvbin2012/NeuralChain/event"
	"github.com/lvbin2012/NeuralChain/neut/downloader"
	"github.com/lvbin2012/NeuralChain/neutdb"
	"github.com/lvbin2012/NeuralChain/params"
	"github.com/lvbin2012/NeuralChain/rpc"
)

// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type Backend interface {
	// General NeuralChain API
	Downloader() *downloader.Downloader
	ProtocolVersion() int
	SuggestPrice(ctx context.Context) (*big.Int, error)
	ChainDb() neutdb.Database
	EventMux() *event.TypeMux
	AccountManager() *accounts.Manager
	ExtRPCEnabled() bool
	RPCGasCap() *big.Int // global gas cap for neut_call over rpc: DoS protection

	// BlockChain API
	SetHead(number uint64)
	HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error)
	BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error)
	StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error)
	GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error)
	GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error)
	GetTd(blockHash common.Hash) *big.Int
	GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header) (*vm.EVM, func() error, error)
	SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription
	SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription
	SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription

	// TxPool API
	SendTx(ctx context.Context, signedTx *types.Transaction) error
	GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error)
	GetPoolTransactions() (types.Transactions, error)
	GetPoolTransaction(txHash common.Hash) *types.Transaction
	GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error)
	Stats() (pending int, queued int)
	TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions)
	SubscribeNewTxsEvent(chan<- core.NewTxsEvent) event.Subscription

	ChainConfig() *params.ChainConfig
	CurrentBlock() *types.Block
}

func GetAPIs(apiBackend Backend) []rpc.API {
	nonceLock := new(AddrLocker)
	return []rpc.API{
		{
			Namespace: "neut",
			Version:   "1.0",
			Service:   NewPublicNeuralChainAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "neut",
			Version:   "1.0",
			Service:   NewPublicBlockChainAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "neut",
			Version:   "1.0",
			Service:   NewPublicTransactionPoolAPI(apiBackend, nonceLock),
			Public:    true,
		}, {
			Namespace: "txpool",
			Version:   "1.0",
			Service:   NewPublicTxPoolAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(apiBackend),
		}, {
			Namespace: "neut",
			Version:   "1.0",
			Service:   NewPublicAccountAPI(apiBackend.AccountManager()),
			Public:    true,
		}, {
			Namespace: "personal",
			Version:   "1.0",
			Service:   NewPrivateAccountAPI(apiBackend, nonceLock),
			Public:    false,
		},
	}
}
