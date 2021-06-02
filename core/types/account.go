package types

import (
	"errors"
	"fmt"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/param"
)

// Account information, not sending a transaction or sending an
// action will increase the nonce value by 1
type Account struct {
	Address         hasharry.Address
	Nonce           uint64
	Time            uint64
	ConfirmedHeight uint64
	ConfirmedNonce  uint64
	ConfirmedTime   uint64
	Coins           *Coins
	OutJournal      *outJournal
	InJournal       *inJournal
}

func NewAccount(stateKey hasharry.Address) *Account {
	coin := &CoinAccount{
		Contract: param.Token.String(),
		Balance:  0,
		LockOut:  0,
		LockIn:   0,
	}
	return &Account{
		Address:         stateKey,
		Nonce:           0,
		Time:            0,
		ConfirmedHeight: 0,
		ConfirmedNonce:  0,
		ConfirmedTime:   0,
		Coins:           &Coins{coin},
		OutJournal:      newOutJournal(),
		InJournal:       newInJournal(),
	}
}

// Calculate the available balance of the current account based on the current effective block height
func (a *Account) Update(confirmedHeight uint64) error {
	confirmedNonce := a.ConfirmedNonce
	confirmedTime := a.ConfirmedTime
	// Update through the account transfer log information
	for _, out := range a.OutJournal.GetJournalOuts(confirmedHeight) {
		coinAccount, ok := a.Coins.Get(out.Contract)
		if !ok {
			return errors.New("wrong journal")
		}
		if coinAccount.LockOut >= out.Amount {
			coinAccount.LockOut -= out.Amount
			a.Coins.Set(coinAccount)

			tokenAccount, ok := a.Coins.Get(param.Token.String())
			if !ok {
				return errors.New("wrong journal")
			}
			if tokenAccount.LockOut >= out.Fees {
				tokenAccount.LockOut -= out.Fees
				a.Coins.Set(tokenAccount)
			} else {
				return errors.New("locked out amount not enough when update account journal")
			}
			a.OutJournal.Remove(out.Height)

		} else {
			return errors.New("locked out amount not enough when update account journal")
		}

		if out.Nonce > confirmedNonce {
			confirmedNonce = out.Nonce
		}
		if out.Time > confirmedTime {
			confirmedTime = out.Time
		}
	}

	// Update through account transfer log information
	for _, in := range a.InJournal.GetJournalIns(confirmedHeight) {
		coinAccount, ok := a.Coins.Get(in.Contract)
		if !ok {
			coinAccount = &CoinAccount{
				Contract: in.Contract,
				Balance:  0,
				LockOut:  0,
				LockIn:   0,
			}
		}
		if coinAccount.LockIn >= in.Amount {
			coinAccount.Balance += in.Amount
			coinAccount.LockIn -= in.Amount
			a.Coins.Set(coinAccount)
			a.InJournal.Remove(in.Height, in.Contract)
		} else {
			return errors.New("locked in amount not enough when update account Journal")
		}
	}
	a.ConfirmedHeight = confirmedHeight
	a.ConfirmedNonce = confirmedNonce
	a.ConfirmedTime = confirmedTime
	return nil
}

func (a *Account) StateKey() hasharry.Address {
	return a.Address
}

func (a *Account) IsExist() bool {
	return !hasharry.EmptyAddress(a.Address)
}

// Determine whether the account needs to be updated. If both
// the transfer-out and transfer-in are 0, no update is required.
func (a *Account) IsNeedUpdate() bool {
	for _, coinContract := range *a.Coins {
		if coinContract.LockOut != 0 || coinContract.LockIn != 0 {
			return true
		}
	}
	return false
}

// Change the account status of the party that transferred the transaction
func (a *Account) TransferChangeFrom(tx ITransaction, blockHeight uint64) error {
	if a.Nonce+1 != tx.GetNonce() {
		return ErrNonce
	}
	contract := tx.GetTxBody().GetContract()
	if contract == param.Token {
		return a.changeUWDTransferFrom(tx, blockHeight)
	} else {
		return a.changeCoinTransferFrom(tx, blockHeight)
	}
}

func (a *Account) TransferV2ChangeFrom(tx ITransaction, blockHeight uint64) error {
	if a.Nonce+1 != tx.GetNonce() {
		return ErrNonce
	}
	contract := tx.GetTxBody().GetContract()
	if contract == param.Token {
		return a.changeUWDTransferV2From(tx, blockHeight)
	} else {
		return a.changeCoinTransferFrom(tx, blockHeight)
	}
}

func (a *Account) ContractChangeFrom(tx ITransaction, blockHeight uint64) error {
	if a.Nonce+1 != tx.GetNonce() {
		return ErrNonce
	}
	fees := tx.GetFees()
	uwd, ok := a.Coins.Get(param.Token.String())
	if !ok {
		return errors.New("account is not exist")
	}
	if uwd.Balance < fees {
		return ErrNotEnoughFees
	}
	uwd.Balance -= fees
	uwd.LockOut += fees
	a.Coins.Set(uwd)
	a.Nonce = tx.GetNonce()
	a.Time = tx.GetTime()
	a.OutJournal.Add(blockHeight, param.Token, 0, fees, tx.GetNonce(), tx.GetTime())
	return nil
}

// Change the primary account status of one party to the transaction transfer
func (a *Account) changeUWDTransferFrom(tx ITransaction, blockHeight uint64) error {
	amount := tx.GetTxBody().GetAmount()
	fees := tx.GetFees()
	if !a.IsExist() {
		a.Address = tx.From()
	}
	uwd, ok := a.Coins.Get(param.Token.String())
	if !ok {
		return errors.New("account is not exist")
	}
	if uwd.Balance < amount {
		return ErrNotEnoughBalance
	}
	if a.Nonce+1 != tx.GetNonce() {
		return ErrNonce
	}

	uwd.Balance -= amount
	uwd.LockOut += amount
	a.Coins.Set(uwd)
	a.Nonce = tx.GetNonce()
	a.Time = tx.GetTime()
	a.OutJournal.Add(blockHeight, param.Token, amount-fees, fees, tx.GetNonce(), tx.GetTime())
	return nil
}

// Change the primary account status of one party to the transaction transfer
func (a *Account) changeUWDTransferV2From(tx ITransaction, blockHeight uint64) error {
	amount := tx.GetTxBody().GetAmount()
	fees := tx.GetFees()
	if !a.IsExist() {
		a.Address = tx.From()
	}
	uwd, ok := a.Coins.Get(param.Token.String())
	if !ok {
		return errors.New("account is not exist")
	}
	if uwd.Balance < amount+fees {
		return ErrNotEnoughBalance
	}
	if a.Nonce+1 != tx.GetNonce() {
		return ErrNonce
	}

	uwd.Balance -= amount + fees
	uwd.LockOut += amount + fees
	a.Coins.Set(uwd)
	a.Nonce = tx.GetNonce()
	a.Time = tx.GetTime()
	a.OutJournal.Add(blockHeight, param.Token, amount, fees, tx.GetNonce(), tx.GetTime())
	return nil
}

// Change the status of the secondary account of the transaction transfer party.
// The transaction of the secondary account needs to consume the fee of the
// primary account.
func (a *Account) changeCoinTransferFrom(tx ITransaction, blockHeight uint64) error {
	fees := tx.GetFees()
	txBody := tx.GetTxBody()
	amount := txBody.GetAmount()
	contract := txBody.GetContract()
	tokenAccount, ok := a.Coins.Get(param.Token.String())
	if !ok {
		return errors.New("account is not exist")
	}
	if tokenAccount.Balance < fees {
		return ErrNotEnoughFees
	}

	coinAccount, ok := a.Coins.Get(contract.String())
	if !ok {
		return errors.New("account is not exist")
	}
	if coinAccount.Balance < amount {
		return ErrNotEnoughBalance
	}

	tokenAccount.Balance -= fees
	tokenAccount.LockOut += fees
	coinAccount.Balance -= amount
	coinAccount.LockOut += amount
	a.Coins.Set(tokenAccount)
	a.Coins.Set(coinAccount)
	a.Nonce = tx.GetNonce()
	a.Time = tx.GetTime()
	a.OutJournal.Add(blockHeight, contract, amount, fees, tx.GetNonce(), tx.GetTime())
	return nil
}

// Change of contract information
func (a *Account) ContractChangeTo(re *Receiver, contract hasharry.Address, blockHeight uint64) error {
	if !a.IsExist() {
		a.Address = re.Address
	}

	amount := re.Amount
	coinAccount, ok := a.Coins.Get(contract.String())
	if ok {
		coinAccount.LockIn += amount
	} else {
		coinAccount = &CoinAccount{
			Contract: contract.String(),
			Balance:  0,
			LockIn:   amount,
			LockOut:  0,
		}
	}
	a.Coins.Set(coinAccount)
	a.InJournal.Add(contract, re.Amount, blockHeight)
	return nil
}

// Change the status of the recipient of the transaction
func (a *Account) TransferChangeTo(re *Receiver, fees uint64, contract hasharry.Address, blockHeight uint64) error {
	if !a.IsExist() {
		a.Address = re.Address
	}
	/*	if tx.GetTxType() == Contract_ {
		return a.toContractChange(tx, blockHeight)
	}*/

	if contract.IsEqual(param.Token) {
		re.Amount = re.Amount - fees
	}

	coinAccount, ok := a.Coins.Get(contract.String())
	if ok {
		coinAccount.LockIn += re.Amount
	} else {
		coinAccount = &CoinAccount{
			Contract: contract.String(),
			Balance:  0,
			LockOut:  0,
			LockIn:   re.Amount,
		}
	}
	a.Coins.Set(coinAccount)
	a.InJournal.Add(contract, re.Amount, blockHeight)
	return nil
}

// Change the status of the recipient of the transaction
func (a *Account) TransferV2ChangeTo(re *Receiver, contract hasharry.Address, blockHeight uint64) error {
	if !a.IsExist() {
		a.Address = re.Address
	}
	/*	if tx.GetTxType() == Contract_ {
		return a.toContractChange(tx, blockHeight)
	}*/

	/*	if contract.IsEqual(param.Token) {
		re.Amount = re.Amount - fees
	}*/

	coinAccount, ok := a.Coins.Get(contract.String())
	if ok {
		coinAccount.LockIn += re.Amount
	} else {
		coinAccount = &CoinAccount{
			Contract: contract.String(),
			Balance:  0,
			LockOut:  0,
			LockIn:   re.Amount,
		}
	}
	a.Coins.Set(coinAccount)
	a.InJournal.Add(contract, re.Amount, blockHeight)
	return nil
}

func (a *Account) TransferOut(token hasharry.Address, amount, height uint64) error {
	tokenInfo, ok := a.Coins.Get(token.String())
	if !ok {
		return ErrNotEnoughBalance
	}
	if tokenInfo.Balance < amount {
		return ErrNotEnoughBalance
	}
	tokenInfo.Balance -= amount
	tokenInfo.LockOut += amount
	a.Coins.Set(tokenInfo)
	a.OutJournal.Add(height, token, amount, 0, 0, 0)
	return nil
}

func (a *Account) TransferIn(token hasharry.Address, amount, height uint64) error {
	tokenInfo, ok := a.Coins.Get(token.String())
	if ok {
		tokenInfo.LockIn += amount
	} else {
		tokenInfo = &CoinAccount{
			Contract: token.String(),
			Balance:  0,
			LockIn:   amount,
			LockOut:  0,
		}
	}
	a.Coins.Set(tokenInfo)
	a.InJournal.Add(token, amount, height)
	return nil
}

// Change the status of the secondary account of the transaction transfer party.
// The transaction of the secondary account needs to consume the fee of the
// primary account.
func (a *Account) fromCoinChange(tx ITransaction, blockHeight uint64) error {
	fees := tx.GetFees()
	txBody := tx.GetTxBody()
	amount := txBody.GetAmount()
	contract := txBody.GetContract()
	tokenAccount, ok := a.Coins.Get(param.Token.String())
	if !ok {
		return errors.New("account is not exist")
	}
	if tokenAccount.Balance < fees {
		return ErrNotEnoughFees
	}

	coinAccount, ok := a.Coins.Get(contract.String())
	if !ok {
		return errors.New("account is not exist")
	}
	if coinAccount.Balance < amount {
		return ErrNotEnoughBalance
	}

	tokenAccount.Balance -= fees
	tokenAccount.LockOut += fees
	coinAccount.Balance -= amount
	coinAccount.LockOut += amount
	a.Coins.Set(tokenAccount)
	a.Coins.Set(coinAccount)
	a.Nonce = tx.GetNonce()
	a.Time = tx.GetTime()
	a.OutJournal.Add(blockHeight, contract, amount, fees, tx.GetNonce(), tx.GetTime())
	return nil
}

// Change of contract information
func (a *Account) fromContractChange(tx ITransaction, blockHeight uint64) error {
	fees := tx.GetFees()
	tokenAccount, ok := a.Coins.Get(param.Token.String())
	if !ok {
		return errors.New("account is not exist")
	}
	if tokenAccount.Balance < fees {
		return ErrNotEnoughFees
	}
	tokenAccount.Balance -= fees
	tokenAccount.LockOut += fees
	a.Coins.Set(tokenAccount)
	a.Nonce = tx.GetNonce()
	a.Time = tx.GetTime()
	a.OutJournal.Add(blockHeight, param.Token, 0, fees, tx.GetNonce(), tx.GetTime())
	return nil
}

// Change of contract information
func (a *Account) fromContractV2Change(tx ITransaction, blockHeight uint64) error {
	fees := tx.GetFees()
	tokenAccount, ok := a.Coins.Get(param.Token.String())
	if !ok {
		return errors.New("account is not exist")
	}
	if tokenAccount.Balance < fees {
		return ErrNotEnoughFees
	}
	tokenAccount.Balance -= fees
	tokenAccount.LockOut += fees
	a.Coins.Set(tokenAccount)
	a.Nonce = tx.GetNonce()
	a.Time = tx.GetTime()
	a.OutJournal.Add(blockHeight, param.Token, 0, fees, tx.GetNonce(), tx.GetTime())
	return nil
}

// Change of contract information
func (a *Account) toContractChange(tx ITransaction, blockHeight uint64) error {
	txBody := tx.GetTxBody()
	amount := txBody.GetAmount()
	contract := txBody.GetContract()
	coinAccount, ok := a.Coins.Get(contract.String())
	if ok {
		coinAccount.LockIn += amount
	} else {
		coinAccount = &CoinAccount{
			Contract: contract.String(),
			Balance:  0,
			LockIn:   amount,
			LockOut:  0,
		}
	}

	a.Coins.Set(coinAccount)
	a.InJournal.Add(txBody.GetContract(), txBody.GetAmount(), blockHeight)
	return nil
}

func (a *Account) FeesChange(fees, blockHeight uint64) {
	if !a.IsExist() {
		a.Address = param.FeeAddress
	}
	coinAccount, ok := a.Coins.Get(param.Token.String())
	if ok {
		coinAccount.LockIn += fees
	} else {
		coinAccount = &CoinAccount{
			Contract: param.Token.String(),
			Balance:  0,
			LockOut:  0,
			LockIn:   fees,
		}
	}
	a.Coins.Set(coinAccount)
	a.InJournal.Add(param.Token, fees, blockHeight)
}

func (a *Account) ConsumptionChange(consumption, blockHeight uint64) {
	if !a.IsExist() {
		a.Address = param.EaterAddress
	}
	coinAccount, ok := a.Coins.Get(param.Token.String())
	if ok {
		coinAccount.LockIn += consumption
	} else {
		coinAccount = &CoinAccount{
			Contract: param.Token.String(),
			Balance:  0,
			LockOut:  0,
			LockIn:   consumption,
		}
	}
	a.Coins.Set(coinAccount)
	a.InJournal.Add(param.Token, consumption, blockHeight)
}

// To verify the transaction status, the nonce value of the transaction
// must be greater than the nonce value of the account of the transferring
// party.
func (a *Account) VerifyTxState(tx ITransaction) error {
	if !a.IsExist() {
		a.Address = tx.GetTxHead().From
	}

	/*	if tx.GetTime() <= a.Time {
			return ErrTime
		}
	*/
	if tx.GetNonce() <= a.Nonce {
		return ErrTxNonceRepeat
	}

	// The nonce value cannot be greater than the
	// maximum number of address transactions
	if tx.GetNonce() > a.Nonce+param.MaxAddressTxs {
		return ErrTooBigNonce
	}

	// Verify the balance of the token
	switch tx.GetTxType() {
	case Transfer_:
		if tx.GetTxBody().GetContract() == param.Token {
			return a.verifyTokenTxBalance(tx)
		} else {
			return a.verifyCoinTxBalance(tx)
		}
	case TransferV2_:
		if tx.GetTxBody().GetContract() == param.Token {
			return a.verifyTransferV2TokenBalance(tx)
		} else {
			return a.verifyCoinTxBalance(tx)
		}
	case Contract_:
		return a.verifyFees(tx)
	default:
		if tx.GetTxBody().GetAmount() != 0 {
			return ErrTxAmount
		}
		return a.verifyFees(tx)
	}
}

func (a *Account) verifyTransferV2TokenBalance(tx ITransaction) error {
	tokenAccount, ok := a.Coins.Get(param.Token.String())
	if !ok {
		return ErrNotEnoughBalance
	} else if tokenAccount.Balance < (tx.GetTxBody().GetAmount() + tx.GetFees()) {
		return ErrNotEnoughBalance
	}
	return nil
}

// Verify the account balance of the primary transaction, the transaction
// value and transaction fee cannot be greater than the balance.
func (a *Account) verifyTokenTxBalance(tx ITransaction) error {
	tokenAccount, ok := a.Coins.Get(param.Token.String())
	if !ok {
		return ErrNotEnoughBalance
	} else if tokenAccount.Balance < tx.GetTxBody().GetAmount() {
		return ErrNotEnoughBalance
	}
	return nil
}

// Verify the account balance of the secondary transaction, the transaction
// value cannot be greater than the balance.
func (a *Account) verifyCoinTxBalance(tx ITransaction) error {
	txBody := tx.GetTxBody()
	if err := a.verifyFees(tx); err != nil {
		return err
	}

	coinAccount, ok := a.Coins.Get(txBody.GetContract().String())
	if !ok {
		return ErrNotEnoughBalance
	} else if coinAccount.Balance < txBody.GetAmount() {
		return ErrNotEnoughBalance
	}
	return nil
}

// Verification fee
func (a *Account) verifyFees(tx ITransaction) error {
	tokenAccount, ok := a.Coins.Get(param.Token.String())
	if !ok {
		return ErrNotEnoughFees
	} else if tokenAccount.Balance < tx.GetFees() {
		return ErrNotEnoughFees
	}
	return nil
}

// The current nonce value of the block transaction must be the
// nonce + 1 of the sender's account.
func (a *Account) VerifyNonce(nonce uint64) error {
	if nonce != a.Nonce+1 {
		return fmt.Errorf("the nonce value must be %d", a.Nonce+1)
	}
	return nil
}

func (a *Account) GetAddress() hasharry.Address {
	return a.Address
}

func (a *Account) GetBalance(contract string) uint64 {
	coinAccount, ok := a.Coins.Get(contract)
	if ok {
		return coinAccount.Balance
	}
	return 0
}

func (a *Account) GetLockedIn(contract string) uint64 {
	coinAccount, ok := a.Coins.Get(contract)
	if ok {
		return coinAccount.LockOut
	}
	return 0
}

func (a *Account) GetLockedOut(contract string) uint64 {
	coinAccount, ok := a.Coins.Get(contract)
	if ok {
		return coinAccount.LockIn
	}
	return 0
}

func (a *Account) GetNonce() uint64 {
	return a.Nonce
}

func (a *Account) GetTime() uint64 {
	return a.Time
}

func (a *Account) GetConfirmedHeight() uint64 {
	return a.ConfirmedHeight
}

func (a *Account) GetConfirmedNonce() uint64 {
	return a.ConfirmedNonce
}

func (a *Account) GetConfirmedTime() uint64 {
	return a.ConfirmedTime
}

// Determine whether the account is in the initial state
func (a *Account) IsEmpty() bool {
	if a.Nonce != 0 {
		return false
	}
	if !a.InJournal.IsEmpty() {
		return false
	}
	if !a.OutJournal.IsEmpty() {
		return false
	}
	for _, coin := range *a.Coins {
		if coin.Balance != 0 || coin.LockOut != 0 || coin.LockIn != 0 {
			return false
		}
	}
	return true
}

type CoinAccount struct {
	Contract string
	Balance  uint64
	LockOut  uint64
	LockIn   uint64
}

// List of secondary accounts
type Coins []*CoinAccount

func (c *Coins) Get(contract string) (*CoinAccount, bool) {
	for _, coin := range *c {
		if coin.Contract == contract {
			return coin, true
		}
	}
	return &CoinAccount{}, false
}

func (c *Coins) Set(newCoin *CoinAccount) {
	for i, coin := range *c {
		if coin.Contract == newCoin.Contract {
			(*c)[i] = newCoin
			return
		}
	}
	*c = append(*c, newCoin)
}

// Account transfer log
type outJournal struct {
	Outs *TxOutList
}

func newOutJournal() *outJournal {
	return &outJournal{Outs: &TxOutList{}}
}

func (j *outJournal) Add(height uint64, contract hasharry.Address, amount, fees, nonce, time uint64) {
	j.Outs.Set(&txOut{
		Contract: contract.String(),
		Amount:   amount,
		Fees:     fees,
		Nonce:    nonce,
		Time:     time,
		Height:   height,
	})
}

func (j *outJournal) Get(height uint64) *txOut {
	out, ok := j.Outs.Get(height)
	if ok {
		return out
	}
	return nil
}

func (j *outJournal) Remove(height uint64) uint64 {
	tx, _ := j.Outs.Get(height)
	j.Outs.Remove(height)
	return tx.Amount
}

func (j *outJournal) IsExist(height uint64) bool {
	for _, txOut := range *j.Outs {
		if txOut.Height >= height {
			return true
		}
	}
	return false
}

func (j *outJournal) GetJournalOuts(confirmedHeight uint64) []*txOut {
	txOuts := make([]*txOut, 0)
	for _, txOut := range *j.Outs {
		if txOut.Height <= confirmedHeight {
			txOuts = append(txOuts, txOut)
		}
	}
	return txOuts
}

func (j *outJournal) Amount() map[string]uint64 {
	amounts := map[string]uint64{}
	for _, txOut := range *j.Outs {
		_, ok := amounts[txOut.Contract]
		if ok {
			amounts[txOut.Contract] += txOut.Amount
		} else {
			amounts[txOut.Contract] = txOut.Amount
		}
	}
	return amounts
}

func (j *outJournal) IsEmpty() bool {
	if j.Outs == nil || len(*j.Outs) == 0 {
		return true
	}
	return false
}

type txOut struct {
	Contract string
	Amount   uint64
	Fees     uint64
	Nonce    uint64
	Time     uint64
	Height   uint64
}

type TxOutList []*txOut

func (t *TxOutList) Get(height uint64) (*txOut, bool) {
	for _, outIn := range *t {
		if outIn.Height == height {
			return outIn, true
		}
	}
	return &txOut{}, false
}

func (t *TxOutList) Set(txOut *txOut) {
	for i, out := range *t {
		if out.Height == txOut.Height && out.Contract == txOut.Contract {
			(*t)[i].Amount += txOut.Amount
			(*t)[i].Fees += txOut.Fees
			return
		}
	}
	*t = append(*t, txOut)
}

func (t *TxOutList) Remove(height uint64) {
	for i, out := range *t {
		if out.Height == height {
			*t = append((*t)[0:i], (*t)[i+1:]...)
			return
		}
	}
}

// Account transfer log
type inJournal struct {
	Ins *InList
}

func newInJournal() *inJournal {
	return &inJournal{Ins: &InList{}}
}

func (j *inJournal) Add(contract hasharry.Address, amount, height uint64) {
	in, ok := j.Ins.Get(height, contract.String())
	if ok {
		in.Amount += amount
	} else {
		in = &InAmount{}
		in.Amount = amount
		in.Height = height
		in.Contract = contract.String()
	}
	j.Ins.Set(in)
}

func (j *inJournal) Get(height uint64, contract string) *InAmount {
	txIn, ok := j.Ins.Get(height, contract)
	if ok {
		return txIn
	}
	return &InAmount{0, "", 0}
}

func (j *inJournal) IsExist(height uint64) bool {
	for _, in := range *j.Ins {
		if in.Height >= height {
			return true
		}
	}
	return false
}

func (j *inJournal) Remove(height uint64, contract string) *InAmount {
	return j.Ins.Remove(height, contract)
}

func (j *inJournal) GetJournalIns(confirmedHeight uint64) map[string]*InAmount {
	txIns := make(map[string]*InAmount)
	for _, in := range *j.Ins {
		if in.Height <= confirmedHeight {
			key := fmt.Sprintf("%s_%d", in.Contract, in.Height)
			txIns[key] = in
		}
	}
	return txIns
}

func (j *inJournal) IsEmpty() bool {
	if j.Ins == nil || len(*j.Ins) == 0 {
		return true
	}
	return false
}

type InAmount struct {
	Amount   uint64
	Contract string
	Height   uint64
}

type InList []*InAmount

func (o *InList) Get(height uint64, contract string) (*InAmount, bool) {
	for _, in := range *o {
		if in.Height == height && in.Contract == contract {
			return in, true
		}
	}
	return &InAmount{}, false
}

func (o *InList) Set(outAmount *InAmount) {
	for i, in := range *o {
		if in.Height == outAmount.Height && in.Contract == outAmount.Contract {
			(*o)[i] = outAmount
			return
		}
	}
	*o = append(*o, outAmount)
}

func (o *InList) Remove(height uint64, contract string) *InAmount {
	for i, in := range *o {
		if in.Height == height && in.Contract == contract {
			*o = append((*o)[0:i], (*o)[i+1:]...)
			return in
		}
	}
	return nil
}
