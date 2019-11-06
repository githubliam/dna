/*
 * Copyright (C) 2018 The DNA Authors
 * This file is part of The DNA library.
 *
 * The DNA is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The DNA is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The DNA.  If not, see <http://www.gnu.org/licenses/>.
 */

package common

import (
	"fmt"
	"github.com/dnaproject2/DNA/account"
	"github.com/dnaproject2/DNA/cmd/utils"
	"github.com/dnaproject2/DNA/common"
	"github.com/dnaproject2/DNA/common/config"
	"github.com/dnaproject2/DNA/common/password"
	"github.com/urfave/cli"
	"strconv"
)

func GetPasswd(ctx *cli.Context) ([]byte, error) {
	var passwd []byte
	var err error
	if ctx.IsSet(utils.GetFlagName(utils.AccountPassFlag)) {
		passwd = []byte(ctx.String(utils.GetFlagName(utils.AccountPassFlag)))
	} else {
		passwd, err = password.GetAccountPassword()
		if err != nil {
			return nil, fmt.Errorf("Input password error:%s", err)
		}
	}
	return passwd, nil
}

func OpenExecutor(ctx *cli.Context) (account.Client, error) {
	executorFile := ctx.String(utils.GetFlagName(utils.ExecutorFileFlag))
	if executorFile == "" {
		executorFile = config.DEFAULT_WALLET_FILE_NAME
	}
	if !common.FileExisted(executorFile) {
		return nil, fmt.Errorf("cannot find executor file:%s", executorFile)
	}
	executor, err := account.Open(executorFile)
	if err != nil {
		return nil, err
	}
	return executor, nil
}

func GetAccountMulti(executor account.Client, passwd []byte, accAddr string) (*account.Account, error) {
	//Address maybe address in base58, label or index
	if accAddr == "" {
		defAcc, err := executor.GetDefaultAccount(passwd)
		if err != nil {
			return nil, err
		}
		return defAcc, nil
	}
	acc, err := executor.GetAccountByAddress(accAddr, passwd)
	if err != nil {
		return nil, fmt.Errorf("getAccountByAddress:%s error:%s", accAddr, err)
	}
	if acc != nil {
		return acc, nil
	}
	acc, err = executor.GetAccountByLabel(accAddr, passwd)
	if err != nil {
		return nil, fmt.Errorf("getAccountByLabel:%s error:%s", accAddr, err)
	}
	if acc != nil {
		return acc, nil
	}
	index, err := strconv.ParseInt(accAddr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("cannot get account by:%s", accAddr)
	}
	acc, err = executor.GetAccountByIndex(int(index), passwd)
	if err != nil {
		return nil, fmt.Errorf("getAccountByIndex:%d error:%s", index, err)
	}
	if acc != nil {
		return acc, nil
	}
	return nil, fmt.Errorf("cannot get account by:%s", accAddr)
}

func GetAccountMetadataMulti(executor account.Client, accAddr string) *account.AccountMetadata {
	//Address maybe address in base58, label or index
	if accAddr == "" {
		fmt.Printf("Using default account:%s\n", accAddr)
		return executor.GetDefaultAccountMetadata()
	}
	acc := executor.GetAccountMetadataByAddress(accAddr)
	if acc != nil {
		return acc
	}
	acc = executor.GetAccountMetadataByLabel(accAddr)
	if acc != nil {
		return acc
	}
	index, err := strconv.ParseInt(accAddr, 10, 32)
	if err != nil {
		return nil
	}
	return executor.GetAccountMetadataByIndex(int(index))
}

func GetAccount(ctx *cli.Context, address ...string) (*account.Account, error) {
	executor, err := OpenExecutor(ctx)
	if err != nil {
		return nil, err
	}
	passwd, err := GetPasswd(ctx)
	if err != nil {
		return nil, err
	}
	defer ClearPasswd(passwd)
	accAddr := ""
	if len(address) > 0 {
		accAddr = address[0]
	} else {
		accAddr = ctx.String(utils.GetFlagName(utils.AccountAddressFlag))
	}
	return GetAccountMulti(executor, passwd, accAddr)
}

func IsBase58Address(address string) bool {
	if address == "" {
		return false
	}
	_, err := common.AddressFromBase58(address)
	return err == nil
}

//ParseAddress return base58 address from base58, label or index
func ParseAddress(address string, ctx *cli.Context) (string, error) {
	if IsBase58Address(address) {
		return address, nil
	}
	executor, err := OpenExecutor(ctx)
	if err != nil {
		return "", err
	}
	acc := executor.GetAccountMetadataByLabel(address)
	if acc != nil {
		return acc.Address, nil
	}
	index, err := strconv.ParseInt(address, 10, 32)
	if err != nil {
		return "", fmt.Errorf("cannot get account by:%s", address)
	}
	acc = executor.GetAccountMetadataByIndex(int(index))
	if acc != nil {
		return acc.Address, nil
	}
	return "", fmt.Errorf("cannot get account by:%s", address)
}

func ClearPasswd(passwd []byte) {
	size := len(passwd)
	for i := 0; i < size; i++ {
		passwd[i] = 0
	}
}
