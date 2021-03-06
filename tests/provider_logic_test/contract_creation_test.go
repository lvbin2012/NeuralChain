package test

import (
	"context"
	"errors"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lvbin2012/NeuralChain/common"
	"github.com/lvbin2012/NeuralChain/common/hexutil"
	"github.com/lvbin2012/NeuralChain/core/types"
	"github.com/lvbin2012/NeuralChain/crypto"
	"github.com/lvbin2012/NeuralChain/neutclient"
)

// TODO: create a global nonce provider so tests not need to wait others complete
/* These tests are done on a chain with already setup account/ contracts.
To run these test, please deploy your own account/ contract and extract privatekey inorder to get the expected result
Adjust these params to match deployment on local machine:
*/

/*
	Test Send ETH to a normal address
		- No provider signature is required
*/

func TestMain(m *testing.M) {
	neutClient, err := neutclient.Dial(neutRPCEndpoint)
	if err != nil {
		panic(err)
	}
	for {
		block, err := neutClient.BlockByNumber(context.Background(), nil)
		if err != nil {
			panic(err)
		}
		if block.Number().Cmp(big.NewInt(2)) > 0 {
			break
		}
	}
	code := m.Run()
	os.Exit(code)
}

func TestCreateContractWithProviderAddress(t *testing.T) {
	spk, err := crypto.HexToECDSA(senderPK)
	assert.NoError(t, err)
	sender, _ := common.NeutAddressStringToAddressCheck(senderAddrStr)
	providerAddr, _ := common.NeutAddressStringToAddressCheck(providerAddrStr)
	payLoadBytes, err := hexutil.Decode(payload)
	assert.NoError(t, err)
	var option types.CreateAccountOption
	option.ProviderAddress = &providerAddr

	neutClient, err := neutclient.Dial(neutRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := neutClient.NonceAt(context.Background(), sender, nil)
	assert.NoError(t, err)
	tx := types.NewContractCreation(nonce, big.NewInt(0), testGasLimit, big.NewInt(testGasPrice), payLoadBytes, option)
	tx, err = types.SignTx(tx, types.BaseSigner{}, spk)
	assert.NoError(t, err)
	err = errors.New("owner is required")
	assert.Error(t, err, neutClient.SendTransaction(context.Background(), tx))
}

func TestCreateContractWithProviderAndOwner(t *testing.T) {
	spk, err := crypto.HexToECDSA(senderPK)
	assert.NoError(t, err)
	sender, _ := common.NeutAddressStringToAddressCheck(senderAddrStr)
	providerAddr, _ := common.NeutAddressStringToAddressCheck(providerAddrStr)
	payLoadBytes, err := hexutil.Decode(payload)
	assert.NoError(t, err)
	var option types.CreateAccountOption
	option.OwnerAddress = &sender
	option.ProviderAddress = &providerAddr

	neutClient, err := neutclient.Dial(neutRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := neutClient.NonceAt(context.Background(), sender, nil)
	assert.NoError(t, err)
	tx := types.NewContractCreation(nonce, big.NewInt(0), testGasLimit, big.NewInt(testGasPrice), payLoadBytes, option)
	tx, err = types.SignTx(tx, types.BaseSigner{}, spk)
	assert.NoError(t, err)
	require.NoError(t, neutClient.SendTransaction(context.Background(), tx))
	assertTransactionSuccess(t, neutClient, tx.Hash(), true, sender)
}

func TestCreateContractWithoutProviderAddress(t *testing.T) {
	spk, err := crypto.HexToECDSA(senderPK)
	assert.NoError(t, err)
	sender, _ := common.NeutAddressStringToAddressCheck(senderAddrStr)
	payLoadBytes, err := hexutil.Decode(payload)
	assert.NoError(t, err)

	neutClient, err := neutclient.Dial(neutRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := neutClient.PendingNonceAt(context.Background(), sender)
	assert.NoError(t, err)
	tx := types.NewContractCreation(nonce, big.NewInt(0), testGasLimit, big.NewInt(testGasPrice), payLoadBytes)
	tx, err = types.SignTx(tx, types.BaseSigner{}, spk)
	assert.NoError(t, err)

	require.NoError(t, neutClient.SendTransaction(context.Background(), tx))
	assertTransactionSuccess(t, neutClient, tx.Hash(), true, sender)
}

func TestCreateContractWithProviderSignature(t *testing.T) {
	spk, err := crypto.HexToECDSA(senderPK)
	assert.NoError(t, err)
	ppk, err := crypto.HexToECDSA(providerPK)
	assert.NoError(t, err)
	sender, _ := common.NeutAddressStringToAddressCheck(senderAddrStr)
	payLoadBytes, err := hexutil.Decode(payload)
	assert.NoError(t, err)

	neutClient, err := neutclient.Dial(neutRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := neutClient.PendingNonceAt(context.Background(), sender)
	assert.NoError(t, err)
	tx := types.NewContractCreation(nonce, big.NewInt(0), testGasLimit, big.NewInt(testGasPrice), payLoadBytes)
	tx, err = types.SignTx(tx, types.BaseSigner{}, spk)
	assert.NoError(t, err)
	tx, err = types.ProviderSignTx(tx, types.BaseSigner{}, ppk)
	assert.NoError(t, err)
	require.Error(t, neutClient.SendTransaction(context.Background(), tx), "Must return error: redundant provider's signature")
}

func TestCreateContractWithProviderAddressWithoutGas(t *testing.T) {
	spk, err := crypto.HexToECDSA(senderPK)
	assert.NoError(t, err)
	sender, _ := common.NeutAddressStringToAddressCheck(senderAddrStr)
	providerAddr, _ := common.NeutAddressStringToAddressCheck(providerWithoutGasAddr)
	var option types.CreateAccountOption
	option.ProviderAddress = &providerAddr
	option.OwnerAddress = &sender
	payLoadBytes, err := hexutil.Decode(payload)
	assert.NoError(t, err)

	neutClient, err := neutclient.Dial(neutRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := neutClient.PendingNonceAt(context.Background(), sender)
	assert.NoError(t, err)
	tx := types.NewContractCreation(nonce, big.NewInt(0), testGasLimit, big.NewInt(testGasPrice), payLoadBytes, option)
	tx, err = types.SignTx(tx, types.BaseSigner{}, spk)
	assert.NoError(t, err)
	require.NoError(t, neutClient.SendTransaction(context.Background(), tx))
	assertTransactionSuccess(t, neutClient, tx.Hash(), true, sender)
}

func TestCreateContractWithProviderAddressMustHaveOwnerAddress(t *testing.T) {
	spk, err := crypto.HexToECDSA(senderPK)
	assert.NoError(t, err)
	sender, _ := common.NeutAddressStringToAddressCheck(senderAddrStr)
	providerAddr, _ := common.NeutAddressStringToAddressCheck(providerAddrStr)
	payLoadBytes, err := hexutil.Decode(payload)
	assert.NoError(t, err)
	var option types.CreateAccountOption
	option.ProviderAddress = &providerAddr
	option.OwnerAddress = &sender

	neutClient, err := neutclient.Dial(neutRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := neutClient.PendingNonceAt(context.Background(), sender)
	assert.NoError(t, err)
	tx := types.NewContractCreation(nonce, big.NewInt(0), testGasLimit, big.NewInt(testGasPrice), payLoadBytes, option)
	tx, err = types.SignTx(tx, types.BaseSigner{}, spk)
	assert.NoError(t, err)
	assert.Equal(t, senderAddrStr, common.AddressToNeutAddressString(*tx.Owner()))
	assert.Equal(t, providerAddrStr, common.AddressToNeutAddressString(*tx.Provider()))
}

func TestCreateNormalContractMustHaveNoOwnerAndProviderAddress(t *testing.T) {
	spk, err := crypto.HexToECDSA(senderPK)
	assert.NoError(t, err)
	sender, _ := common.NeutAddressStringToAddressCheck(senderAddrStr)
	payLoadBytes, err := hexutil.Decode(payload)
	assert.NoError(t, err)

	neutClient, err := neutclient.Dial(neutRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := neutClient.PendingNonceAt(context.Background(), sender)
	assert.NoError(t, err)
	tx := types.NewContractCreation(nonce, big.NewInt(0), testGasLimit, big.NewInt(testGasPrice), payLoadBytes)
	tx, err = types.SignTx(tx, types.BaseSigner{}, spk)
	assert.NoError(t, err)
	assert.Nil(t, tx.Owner())
	assert.Nil(t, tx.Provider())
}

func assertTransactionSuccess(t *testing.T, client *neutclient.Client, txHash common.Hash, contractCreation bool, gasPayer common.Address) {
	for i := 0; i < getReceiptMaxRetries; i++ {
		var receipt *types.Receipt
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			assert.Equal(t, uint64(1), receipt.Status)
			if contractCreation {
				assert.NotEqual(t, receipt.ContractAddress, common.Address{}, "not contract creation")
			}
			assert.Equal(t, gasPayer, receipt.GasPayer, "unexpected gas payer")
			return
		}
		time.Sleep(1 * time.Second)
	}
	t.Errorf("transaction %s not found", txHash.Hex())
}
