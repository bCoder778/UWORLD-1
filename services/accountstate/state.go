package accountstate

import (
	"errors"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/database/statedb"
	"github.com/uworldao/UWORLD/param"
	"sync"
	"time"
)

const accountSate = "account_state"

type AccountState struct {
	stateDb         IAccountStorage
	accountMutex    sync.RWMutex
	contractMutex   sync.RWMutex
	confirmedHeight uint64
}

func NewAccountState(dataDir string) (*AccountState, error) {
	storage := statedb.NewStateStorage(dataDir + "/" + accountSate)
	err := storage.Open()
	if err != nil {
		return nil, err
	}
	return &AccountState{
		stateDb: storage,
	}, nil
}

// Initialize account balance root hash
func (as *AccountState) InitTrie(stateRoot hasharry.Hash) error {
	return as.stateDb.InitTrie(stateRoot)
}

// Get account status, if the account status needs to be updated
// according to the effective block height, it will be updated,
// but not stored.
func (as *AccountState) GetAccountState(stateKey hasharry.Address) types.IAccount {
	as.accountMutex.RLock()
	account := as.stateDb.GetAccountState(stateKey)
	as.accountMutex.RUnlock()

	if account.IsNeedUpdate() {
		account = as.updateAccountLocked(stateKey)
	}
	return account
}

func (as *AccountState) getAccountState(stateKey hasharry.Address) types.IAccount {
	account := as.stateDb.GetAccountState(stateKey)

	if account.IsNeedUpdate() {
		account = as.updateAccountLocked(stateKey)
	}
	return account
}

func (as *AccountState) GetAccountNonce(stateKey hasharry.Address) (uint64, error) {
	as.accountMutex.RLock()
	defer as.accountMutex.RUnlock()

	return as.stateDb.GetAccountNonce(stateKey), nil
}

func (as *AccountState) setAccountState(account types.IAccount) {
	as.stateDb.SetAccountState(account)
}

// Update sender account status based on transaction information
func (as *AccountState) UpdateContractFrom(tx types.ITransaction, blockHeight uint64) error {
	if tx.IsCoinBase() {
		return nil
	}

	as.accountMutex.Lock()
	defer as.accountMutex.Unlock()

	fromAccount := as.stateDb.GetAccountState(tx.From())
	err := fromAccount.Update(as.confirmedHeight)
	if err != nil {
		return err
	}

	err = fromAccount.ContractChangeFrom(tx, blockHeight)
	if err != nil {
		return err
	}

	as.setAccountState(fromAccount)
	return nil
}

// Update sender account status based on transaction information
func (as *AccountState) UpdateTransferFrom(tx types.ITransaction, blockHeight uint64) error {
	if tx.IsCoinBase() {
		return nil
	}

	as.accountMutex.Lock()
	defer as.accountMutex.Unlock()

	fromAccount := as.stateDb.GetAccountState(tx.From())
	err := fromAccount.Update(as.confirmedHeight)
	if err != nil {
		return err
	}

	err = fromAccount.TransferChangeFrom(tx, blockHeight)
	if err != nil {
		return err
	}

	as.setAccountState(fromAccount)
	return nil
}

func (as *AccountState) UpdateTransferV2From(tx types.ITransaction, blockHeight uint64) error {
	if tx.IsCoinBase() {
		return nil
	}

	as.accountMutex.Lock()
	defer as.accountMutex.Unlock()

	fromAccount := as.stateDb.GetAccountState(tx.From())
	err := fromAccount.Update(as.confirmedHeight)
	if err != nil {
		return err
	}

	err = fromAccount.TransferV2ChangeFrom(tx, blockHeight)
	if err != nil {
		return err
	}

	as.setAccountState(fromAccount)
	return nil
}

// Update the receiver's account status based on transaction information
func (as *AccountState) UpdateTransferV2To(tx types.ITransaction, blockHeight uint64) error {
	as.accountMutex.Lock()
	defer as.accountMutex.Unlock()

	var toAccount types.IAccount

	receivers := tx.GetTxBody().ToAddress().ReceiverList()
	for _, re := range receivers {
		toAccount = as.stateDb.GetAccountState(re.Address)
		err := toAccount.Update(as.confirmedHeight)
		if err != nil {
			return err
		}
		err = toAccount.TransferV2ChangeTo(re, tx.GetTxBody().GetContract(), blockHeight)
		if err != nil {
			return err
		}
		as.setAccountState(toAccount)
	}
	return nil
}

func (as *AccountState) UpdateTransferTo(tx types.ITransaction, blockHeight uint64) error {
	as.accountMutex.Lock()
	defer as.accountMutex.Unlock()

	var toAccount types.IAccount

	receivers := tx.GetTxBody().ToAddress().ReceiverList()
	re := receivers[0]
	toAccount = as.stateDb.GetAccountState(re.Address)
	err := toAccount.Update(as.confirmedHeight)
	if err != nil {
		return err
	}

	err = toAccount.TransferChangeTo(re, tx.GetFees(), tx.GetTxBody().GetContract(), blockHeight)
	if err != nil {
		return err
	}

	as.setAccountState(toAccount)
	return nil
}

func (as *AccountState) UpdateContractTo(tx types.ITransaction, blockHeight uint64) error {
	as.accountMutex.Lock()
	defer as.accountMutex.Unlock()

	var toAccount types.IAccount

	receivers := tx.GetTxBody().ToAddress().ReceiverList()
	re := receivers[0]
	toAccount = as.stateDb.GetAccountState(re.Address)
	err := toAccount.Update(as.confirmedHeight)
	if err != nil {
		return err
	}

	err = toAccount.ContractChangeTo(re, tx.GetTxBody().GetContract(), blockHeight)
	if err != nil {
		return err
	}

	as.setAccountState(toAccount)
	return nil
}

func (as *AccountState) UpdateFees(fees, blockHeight uint64) error {
	as.accountMutex.Lock()
	defer as.accountMutex.Unlock()

	var account types.IAccount

	account = as.stateDb.GetAccountState(param.FeeAddress)
	err := account.Update(as.confirmedHeight)
	if err != nil {
		return err
	}
	account.FeesChange(fees, blockHeight)
	as.setAccountState(account)
	return nil
}

func (as *AccountState) UpdateConsumption(fees, blockHeight uint64) error {
	if fees == 0 {
		return nil
	}

	as.accountMutex.Lock()
	defer as.accountMutex.Unlock()

	var account types.IAccount

	account = as.stateDb.GetAccountState(param.EaterAddress)
	err := account.Update(as.confirmedHeight)
	if err != nil {
		return err
	}
	account.ConsumptionChange(fees, blockHeight)
	as.setAccountState(account)
	return nil
}

// Update the locked balance of an account
func (as *AccountState) updateAccountLocked(stateKey hasharry.Address) types.IAccount {
	account := as.stateDb.GetAccountState(stateKey)
	account.Update(as.confirmedHeight)
	return account
}

func (as *AccountState) UpdateConfirmedHeight(height uint64) {
	as.confirmedHeight = height
}

// Verify the status of the trading account
func (as *AccountState) VerifyState(tx types.ITransaction) error {
	switch tx.GetTxType() {
	default:
		return as.verifyTxState(tx)
	}
}
func (as *AccountState) Transfer(from, to, token hasharry.Address, amount uint64, height uint64) error {
	as.accountMutex.Lock()
	defer as.accountMutex.Unlock()

	fromAcc := as.getAccountState(from)
	if err := fromAcc.TransferOut(token, amount, height); err != nil {
		return err
	}
	toAcc := as.getAccountState(to)
	if err := toAcc.TransferIn(token, amount, height); err != nil {
		return err
	}
	as.setAccountState(fromAcc)
	as.setAccountState(toAcc)
	return nil
}

func (as *AccountState) PreTransfer(from, to, token hasharry.Address, amount uint64, height uint64) error {
	as.accountMutex.RLock()
	defer as.accountMutex.RUnlock()

	fromAcc := as.getAccountState(from)
	if err := fromAcc.TransferOut(token, amount, height); err != nil {
		return err
	}
	toAcc := as.getAccountState(to)
	if err := toAcc.TransferIn(token, amount, height); err != nil {
		return err
	}
	return nil
}

func (as *AccountState) verifyTxState(tx types.ITransaction) error {
	if tx.GetTime() > uint64(time.Now().Unix()) {
		return errors.New("incorrect transaction time")
	}

	account := as.GetAccountState(tx.From())
	return account.VerifyTxState(tx)
}

func (as *AccountState) StateTrieCommit() (hasharry.Hash, error) {
	return as.stateDb.Commit()
}

func (as *AccountState) RootHash() hasharry.Hash {
	//as.Print()
	return as.stateDb.RootHash()
}

func (as *AccountState) Print() {
	as.stateDb.Print()
}

func (as *AccountState) Close() error {
	return as.stateDb.Close()
}
