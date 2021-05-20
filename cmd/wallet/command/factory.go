package command

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/uworldao/UWORLD/common/hasharry"
	"github.com/uworldao/UWORLD/core/types"
	"github.com/uworldao/UWORLD/rpc/rpctypes"
	"github.com/uworldao/UWORLD/ut/transaction"
	"strconv"
)

func init() {
	factoryCmds := []*cobra.Command{
		CreateFactoryCmd,
		SetFactoryAdminCmd,
		SetFactoryFeeToCmd,
		CreatePairCmd,
	}
	RootCmd.AddCommand(factoryCmds...)
	RootSubCmdGroups["factory"] = factoryCmds

}

var CreateFactoryCmd = &cobra.Command{
	Use:     "CreateFactory {from} {admin} {feeTo} {password} {nonce}; Create a decentralized factory;",
	Aliases: []string{"CreateFactory", "createfactory", "cf", "CF"},
	Short:   "CreateFactory {from} {admin} {feeTo} {password} {nonce}; Create a decentralized factory;",
	Example: `
	CreateFactory 3ajDJUnMYDyzXLwefRfNp7yLcdmg3ULb9ndQ 3ajNkh7yVYkETL9JKvGx3aL2YVNrqksjCUUE 3ajNkh7yVYkETL9JKvGx3aL2YVNrqksjCUUE 123456
		OR
	CreateFactory 3ajDJUnMYDyzXLwefRfNp7yLcdmg3ULb9ndQ 3ajNkh7yVYkETL9JKvGx3aL2YVNrqksjCUUE 3ajNkh7yVYkETL9JKvGx3aL2YVNrqksjCUUE 123456 1
	`,
	Args: cobra.MinimumNArgs(3),
	Run:  CreateFactory,
}

func CreateFactory(cmd *cobra.Command, args []string) {
	var passwd []byte
	var err error
	if len(args) > 3 {
		passwd = []byte(args[3])
	} else {
		fmt.Println("please input password：")
		passwd, err = readPassWd()
		if err != nil {
			log.Error(cmd.Use+" err: ", fmt.Errorf("read password failed! %s", err.Error()))
			return
		}
	}
	privKey, err := ReadAddrPrivate(getAddJsonPath(args[0]), passwd)
	if err != nil {
		log.Error(cmd.Use+" err: ", fmt.Errorf("wrong password"))
		return
	}
	resp, err := GetAccountByRpc(args[0])
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	if resp.Code != 0 {
		log.Errorf(cmd.Use+" err: code %d, message: %s", resp.Code, resp.Err)
		return
	}
	var account *rpctypes.Account
	if err := json.Unmarshal(resp.Result, &account); err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}

	tx, err := parseCEParams(args, account.Nonce+1)
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}

	if !signTx(cmd, tx, privKey.Private) {
		log.Error(cmd.Use+" err: ", errors.New("signature failure"))
		return
	}

	rs, err := sendTx(cmd, tx)
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
	} else if rs.Code != 0 {
		log.Errorf(cmd.Use+" err: code %d, message: %s", rs.Code, rs.Err)
	} else {
		fmt.Println()
		fmt.Println(string(rs.Result))
	}
}

func parseCEParams(args []string, nonce uint64) (*types.Transaction, error) {
	var err error
	from := hasharry.StringToAddress(args[0])
	admin := args[1]
	feeTo := args[2]
	if len(args) > 4 {
		nonce, err = strconv.ParseUint(args[4], 10, 64)
		if err != nil {
			return nil, errors.New("wrong nonce")
		}
	}
	tx, err := transaction.NewFactory(Net, from.String(), admin, feeTo, nonce, "")
	if err != nil {
		return nil, err
	}
	return tx, nil
}

var SetFactoryAdminCmd = &cobra.Command{
	Use:     "SetFactoryAdmin {from} {factory} {admin} {password} {nonce}; Set factory feeTo setter;",
	Aliases: []string{"setfactoryadmin", "sfa", "SFA"},
	Short:   "SetFactoryAdmin {from} {factory} {admin} {password} {nonce}; Set factory feeTo setter;",
	Example: `
	SetFactoryAdmin UWDGLmQMfEeF6Fh8CGztrSktnHVpCxLiheYw UWTfBGxDMZX19vjnacXVkP51min9EjhYq43W UWDGLmQMfEeF6Fh8CGztrSktnHVpCxLiheYw 123456
		OR
	SetFactoryAdmin UWDGLmQMfEeF6Fh8CGztrSktnHVpCxLiheYw UWTfBGxDMZX19vjnacXVkP51min9EjhYq43W UWDGLmQMfEeF6Fh8CGztrSktnHVpCxLiheYw 123456 1
	`,
	Args: cobra.MinimumNArgs(3),
	Run:  SetFactoryAdmin,
}

func SetFactoryAdmin(cmd *cobra.Command, args []string) {
	var passwd []byte
	var err error
	if len(args) > 3 {
		passwd = []byte(args[3])
	} else {
		fmt.Println("please input password：")
		passwd, err = readPassWd()
		if err != nil {
			log.Error(cmd.Use+" err: ", fmt.Errorf("read password failed! %s", err.Error()))
			return
		}
	}
	privKey, err := ReadAddrPrivate(getAddJsonPath(args[0]), passwd)
	if err != nil {
		log.Error(cmd.Use+" err: ", fmt.Errorf("wrong password"))
		return
	}
	resp, err := GetAccountByRpc(args[0])
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	if resp.Code != 0 {
		log.Errorf(cmd.Use+" err: code %d, message: %s", resp.Code, resp.Err)
		return
	}
	var account *rpctypes.Account
	if err := json.Unmarshal(resp.Result, &account); err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}

	tx, err := parseSEFTSParams(args, account.Nonce+1)
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}

	if !signTx(cmd, tx, privKey.Private) {
		log.Error(cmd.Use+" err: ", errors.New("signature failure"))
		return
	}

	rs, err := sendTx(cmd, tx)
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
	} else if rs.Code != 0 {
		log.Errorf(cmd.Use+" err: code %d, message: %s", rs.Code, rs.Err)
	} else {
		fmt.Println()
		fmt.Println(string(rs.Result))
	}
}

func parseSEFTSParams(args []string, nonce uint64) (*types.Transaction, error) {
	var err error
	from := args[0]
	factory := args[1]
	admin := args[2]
	if len(args) > 4 {
		nonce, err = strconv.ParseUint(args[4], 10, 64)
		if err != nil {
			return nil, errors.New("wrong nonce")
		}
	}
	tx, err := transaction.NewSetAdmin(from, factory, admin, nonce, "")
	if err != nil {
		return nil, err
	}
	return tx, nil
}

var SetFactoryFeeToCmd = &cobra.Command{
	Use:     "SetFactoryFeeTo {from} {factory} {feeTo} {password} {nonce}; Set factory feeTo;",
	Aliases: []string{"setfactoryfeeto", "sfft", "SFFT"},
	Short:   "SetFactoryFeeTo {from} {factory} {feeTo} {password} {nonce}; Set factory feeTo;",
	Example: `
	SetFactoryFeeTo UWDGLmQMfEeF6Fh8CGztrSktnHVpCxLiheYw UWTfBGxDMZX19vjnacXVkP51min9EjhYq43W UWDGLmQMfEeF6Fh8CGztrSktnHVpCxLiheYw 123456
		OR
	SetFactoryFeeTo UWDGLmQMfEeF6Fh8CGztrSktnHVpCxLiheYw UWTfBGxDMZX19vjnacXVkP51min9EjhYq43W UWDGLmQMfEeF6Fh8CGztrSktnHVpCxLiheYw 123456 1
	`,
	Args: cobra.MinimumNArgs(3),
	Run:  SetFactoryFeeTo,
}

func SetFactoryFeeTo(cmd *cobra.Command, args []string) {
	var passwd []byte
	var err error
	if len(args) > 3 {
		passwd = []byte(args[3])
	} else {
		fmt.Println("please input password：")
		passwd, err = readPassWd()
		if err != nil {
			log.Error(cmd.Use+" err: ", fmt.Errorf("read password failed! %s", err.Error()))
			return
		}
	}
	privKey, err := ReadAddrPrivate(getAddJsonPath(args[0]), passwd)
	if err != nil {
		log.Error(cmd.Use+" err: ", fmt.Errorf("wrong password"))
		return
	}
	resp, err := GetAccountByRpc(args[0])
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	if resp.Code != 0 {
		log.Errorf(cmd.Use+" err: code %d, message: %s", resp.Code, resp.Err)
		return
	}
	var account *rpctypes.Account
	if err := json.Unmarshal(resp.Result, &account); err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}

	tx, err := parseSEFTParams(args, account.Nonce+1)
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}

	if !signTx(cmd, tx, privKey.Private) {
		log.Error(cmd.Use+" err: ", errors.New("signature failure"))
		return
	}

	rs, err := sendTx(cmd, tx)
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
	} else if rs.Code != 0 {
		log.Errorf(cmd.Use+" err: code %d, message: %s", rs.Code, rs.Err)
	} else {
		fmt.Println()
		fmt.Println(string(rs.Result))
	}
}

func parseSEFTParams(args []string, nonce uint64) (*types.Transaction, error) {
	var err error
	from := args[0]
	factory := args[1]
	feeTo := args[2]
	if len(args) > 4 {
		nonce, err = strconv.ParseUint(args[4], 10, 64)
		if err != nil {
			return nil, errors.New("wrong nonce")
		}
	}
	tx, err := transaction.NewSetFeeTo(from, factory, feeTo, nonce, "")
	if err != nil {
		return nil, err
	}
	return tx, nil
}

var CreatePairCmd = &cobra.Command{
	Use:     "CreatePair {from} {to} {factory} {tokenA} {amountADesired} {amountAmin} {tokenB} {amountBDesired} {amountBMin} {password} {nonce};Create a pair contract;",
	Aliases: []string{"createpair", "cp", "CP"},
	Short:   "CreatePair {from} {to} {factory} {tokenA} {amountADesired} {amountAmin} {tokenB} {amountBDesired} {amountBMin} {password} {nonce}; Create a pair contract;",
	Example: `
	CreatePair UWDGLmQMfEeF6Fh8CGztrSktnHVpCxLiheYw UWDGLmQMfEeF6Fh8CGztrSktnHVpCxLiheYw UWTfBGxDMZX19vjnacXVkP51min9EjhYq43W UWTXEqvUWik48uAHcJXZiyyWMy4GLtpGuttL 100 90 UWD 1 0.9 123456
		OR
	CreatePair UWDGLmQMfEeF6Fh8CGztrSktnHVpCxLiheYw UWDGLmQMfEeF6Fh8CGztrSktnHVpCxLiheYw UWTfBGxDMZX19vjnacXVkP51min9EjhYq43W UWTXEqvUWik48uAHcJXZiyyWMy4GLtpGuttL 100 90 UWD 1 0.9 123456 1
	`,
	Args: cobra.MinimumNArgs(9),
	Run:  CreatePair,
}

func CreatePair(cmd *cobra.Command, args []string) {
	var passwd []byte
	var err error
	if len(args) > 9 {
		passwd = []byte(args[9])
	} else {
		fmt.Println("please input password：")
		passwd, err = readPassWd()
		if err != nil {
			log.Error(cmd.Use+" err: ", fmt.Errorf("read password failed! %s", err.Error()))
			return
		}
	}
	privKey, err := ReadAddrPrivate(getAddJsonPath(args[0]), passwd)
	if err != nil {
		log.Error(cmd.Use+" err: ", fmt.Errorf("wrong password"))
		return
	}
	resp, err := GetAccountByRpc(args[0])
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}
	if resp.Code != 0 {
		log.Errorf(cmd.Use+" err: code %d, message: %s", resp.Code, resp.Err)
		return
	}
	var account *rpctypes.Account
	if err := json.Unmarshal(resp.Result, &account); err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}

	tx, err := parseCPParams(args, account.Nonce+1)
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
		return
	}

	if !signTx(cmd, tx, privKey.Private) {
		log.Error(cmd.Use+" err: ", errors.New("signature failure"))
		return
	}

	rs, err := sendTx(cmd, tx)
	if err != nil {
		log.Error(cmd.Use+" err: ", err)
	} else if rs.Code != 0 {
		log.Errorf(cmd.Use+" err: code %d, message: %s", rs.Code, rs.Err)
	} else {
		fmt.Println()
		fmt.Println(string(rs.Result))
	}
}

func parseCPParams(args []string, nonce uint64) (*types.Transaction, error) {
	var err error
	from := args[0]
	to := args[1]
	factory := args[2]
	tokenA := args[3]
	amountADesiredf, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		return nil, errors.New("wrong amountADesired")
	}
	amountADesired, _ := types.NewAmount(amountADesiredf)
	amountAMinf, err := strconv.ParseFloat(args[5], 64)
	if err != nil {
		return nil, errors.New("wrong amountAMin")
	}
	amountAMin, _ := types.NewAmount(amountAMinf)
	tokenB := args[6]
	amountBDesiredf, err := strconv.ParseFloat(args[7], 64)
	if err != nil {
		return nil, errors.New("wrong amountBDesired")
	}
	amountBDesired, _ := types.NewAmount(amountBDesiredf)
	amountBMinf, err := strconv.ParseFloat(args[8], 64)
	if err != nil {
		return nil, errors.New("wrong amountBMin")
	}
	amountBMin, _ := types.NewAmount(amountBMinf)
	if len(args) > 10 {
		nonce, err = strconv.ParseUint(args[10], 10, 64)
		if err != nil {
			return nil, errors.New("wrong nonce")
		}
	}
	tx, err := transaction.NewPairCreate(Net, from, to, factory, tokenA, tokenB, amountADesired, amountBDesired, amountAMin, amountBMin, nonce, "")
	if err != nil {
		return nil, err
	}
	return tx, nil
}
