package test

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lvbin2012/NeuralChain/common"
	"github.com/lvbin2012/NeuralChain/core/types"
	"github.com/lvbin2012/NeuralChain/crypto"

	"github.com/lvbin2012/NeuralChain/neutclient"
)

/* These tests are done on a chain with already setup account/ contracts.
To run these test, please deploy your own account/ contract and extract privatekey inorder to get the expected result
*/

// TestInteractToEnterpriseSmartContractWithValidProviderSignatureFromAccountWithoutGas
// Will attempt to reproduce logic of provider paying gas fee.
// It should be send from address without any native token
// The balance of provider should be check prior and after the transaction is mined to
// assure the correctness of the program.
func TestInteractToEnterpriseSmartContractWithValidProviderSignatureFromAccountWithoutGas(t *testing.T) {
	var (
		senderAddr, _ = common.NeutAddressStringToAddressCheck(senderWithoutGasAddrStr)
		contractAddr  = prepareNewContract(true)
	)

	spk, err := crypto.HexToECDSA(senderWithoutGasPK)
	assert.NoError(t, err)

	ppk, err := crypto.HexToECDSA(providerPK)
	assert.NoError(t, err)
	signer := types.BaseSigner{}
	neutClient, err := neutclient.Dial(neutRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := neutClient.PendingNonceAt(context.Background(), senderAddr)
	assert.NoError(t, err)
	gasPrice, err := neutClient.SuggestGasPrice(context.Background())
	assert.NoError(t, err)

	// data to interact with a function of this contract
	dataBytes := []byte("0x3fb5c1cb0000000000000000000000000000000000000000000000000000000000000002")
	transaction := types.NewTransaction(nonce, *contractAddr, big.NewInt(0), testGasLimit, gasPrice, dataBytes)
	transaction, err = types.SignTx(transaction, signer, spk)
	assert.NoError(t, err)
	transaction, err = types.ProviderSignTx(transaction, signer, ppk)
	assert.NoError(t, err)
	require.NoError(t, neutClient.SendTransaction(context.Background(), transaction))
	providerAddr, _ := common.NeutAddressStringToAddressCheck(providerAddrStr)
	assertTransactionSuccess(t, neutClient, transaction.Hash(), false, providerAddr)
}

// Interact with a payable function and sending some native token along with transaction
// Please make sure the sender does not have any funds
// expected to get revert as sender's balance is not enough for transaction amount
func TestInteractWithAmountToEnterpriseSmartContractWithValidProviderSignatureFromAccountWithoutGas(t *testing.T) {
	senderAddr, _ := common.NeutAddressStringToAddressCheck(senderWithoutGasAddrStr)

	contractAddr := prepareNewContract(false)
	assert.NotNil(t, contractAddr)

	spk, err := crypto.HexToECDSA(senderWithoutGasPK)
	assert.NoError(t, err)

	ppk, err := crypto.HexToECDSA(providerPK)
	assert.NoError(t, err)
	signer := types.BaseSigner{}
	neutClient, err := neutclient.Dial(neutRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := neutClient.PendingNonceAt(context.Background(), senderAddr)
	assert.NoError(t, err)
	gasPrice, err := neutClient.SuggestGasPrice(context.Background())
	assert.NoError(t, err)

	// data to interact with a function of this contract
	dataBytes := []byte("0x3fb5c1cb0000000000000000000000000000000000000000000000000000000000000002")
	transaction := types.NewTransaction(nonce, *contractAddr, big.NewInt(1000000), testGasLimit, gasPrice, dataBytes)
	transaction, err = types.SignTx(transaction, signer, spk)
	assert.NoError(t, err)
	transaction, err = types.ProviderSignTx(transaction, signer, ppk)
	assert.NoError(t, err)

	require.Error(t, neutClient.SendTransaction(context.Background(), transaction))
}

// Interact with a payable function and sending some native token along with transaction
// Please make sure sender has enough balance to cover transaction amount
// expected to get passed as sender's balance is enough for transaction amount
func TestInteractWithAmountToEnterpriseSmartContractWithValidProviderSignatureFromAccountWithEnoughBalance(t *testing.T) {
	senderAddr, _ := common.NeutAddressStringToAddressCheck(senderAddrStr)
	contractAddr := prepareNewContract(true)
	assert.NotNil(t, contractAddr)

	spk, err := crypto.HexToECDSA(senderPK)
	assert.NoError(t, err)

	ppk, err := crypto.HexToECDSA(providerPK)
	assert.NoError(t, err)
	signer := types.BaseSigner{}
	neutClient, err := neutclient.Dial(neutRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := neutClient.PendingNonceAt(context.Background(), senderAddr)
	assert.NoError(t, err)
	gasPrice, err := neutClient.SuggestGasPrice(context.Background())
	assert.NoError(t, err)

	// data to interact with a function of this contract
	dataBytes := []byte("0x3fb5c1cb0000000000000000000000000000000000000000000000000000000000000002")
	transaction := types.NewTransaction(nonce, *contractAddr, big.NewInt(1000000), testGasLimit, gasPrice, dataBytes)
	transaction, err = types.SignTx(transaction, signer, spk)
	assert.NoError(t, err)
	transaction, err = types.ProviderSignTx(transaction, signer, ppk)
	assert.NoError(t, err)

	require.NoError(t, neutClient.SendTransaction(context.Background(), transaction))
	providerAddr, _ := common.NeutAddressStringToAddressCheck(providerAddrStr)
	assertTransactionSuccess(t, neutClient, transaction.Hash(), false, providerAddr)
}

// Interact with enterprise contract where provider has zero gas
// Please make sure sender has balance and provider has zero balance
// Expected to get failure as provider's balance is not enough for transaction fee
// Please check error message
func TestInteractEnterpriseSmartContractWithValidProviderSignatureWithoutGas(t *testing.T) {
	senderAddr, _ := common.NeutAddressStringToAddressCheck(senderAddrStr)
	contractAddr, _ := common.NeutAddressStringToAddressCheck(contractProviderWithoutGas)

	spk, err := crypto.HexToECDSA(senderPK)
	assert.NoError(t, err)

	ppk, err := crypto.HexToECDSA(providerWithoutGasPK)
	assert.NoError(t, err)
	signer := types.BaseSigner{}
	neutClient, err := neutclient.Dial(neutRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := neutClient.PendingNonceAt(context.Background(), senderAddr)
	assert.NoError(t, err)
	gasPrice, err := neutClient.SuggestGasPrice(context.Background())
	assert.NoError(t, err)

	// data to interact with a function of this contract
	dataBytes := []byte("0x3fb5c1cb0000000000000000000000000000000000000000000000000000000000000002")
	transaction := types.NewTransaction(nonce, contractAddr, big.NewInt(0), testGasLimit, gasPrice, dataBytes)
	transaction, err = types.SignTx(transaction, signer, spk)
	assert.NoError(t, err)
	transaction, err = types.ProviderSignTx(transaction, signer, ppk)
	assert.NoError(t, err)

	require.Error(t, neutClient.SendTransaction(context.Background(), transaction))
}
