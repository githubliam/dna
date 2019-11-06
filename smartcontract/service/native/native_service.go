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

package native

import (
	"fmt"

	"github.com/dnaproject2/DNA/common"
	"github.com/dnaproject2/DNA/core/types"
	"github.com/dnaproject2/DNA/errors"
	"github.com/dnaproject2/DNA/smartcontract/context"
	"github.com/dnaproject2/DNA/smartcontract/event"
	"github.com/dnaproject2/DNA/smartcontract/states"
	sstates "github.com/dnaproject2/DNA/smartcontract/states"
	"github.com/dnaproject2/DNA/smartcontract/storage"
)

type (
	Handler         func(native *NativeService) ([]byte, error)
	RegisterService func(native *NativeService)
)

var (
	Contracts = make(map[common.Address]RegisterService)
)

// Native service struct
// Invoke a native smart contract, new a native service
type NativeService struct {
	CacheDB       *storage.CacheDB
	ServiceMap    map[string]Handler
	Notifications []*event.NotifyEventInfo
	InvokeParam   sstates.ContractInvokeParam
	Input         []byte
	Tx            *types.Transaction
	Height        uint32
	Time          uint32
	BlockHash     common.Uint256
	ContextRef    context.ContextRef
}

func (this *NativeService) Register(methodName string, handler Handler) {
	this.ServiceMap[methodName] = handler
}

func (this *NativeService) Invoke() (interface{}, error) {
	contract := this.InvokeParam
	services, ok := Contracts[contract.Address]
	if !ok {
		return false, fmt.Errorf("Native contract address %x haven't been registered.", contract.Address)
	}
	services(this)
	service, ok := this.ServiceMap[contract.Method]
	if !ok {
		return false, fmt.Errorf("Native contract %x doesn't support this function %s.",
			contract.Address, contract.Method)
	}
	args := this.Input
	this.Input = contract.Args
	this.ContextRef.PushContext(&context.Context{ContractAddress: contract.Address})
	notifications := this.Notifications
	this.Notifications = []*event.NotifyEventInfo{}
	result, err := service(this)
	if err != nil {
		return result, errors.NewDetailErr(err, errors.ErrNoCode, "[Invoke] Native serivce function execute error!")
	}
	this.ContextRef.PopContext()
	this.ContextRef.PushNotifications(this.Notifications)
	this.Notifications = notifications
	this.Input = args
	return result, nil
}

func (this *NativeService) NativeCall(address common.Address, method string, args []byte) (interface{}, error) {
	c := states.ContractInvokeParam{
		Address: address,
		Method:  method,
		Args:    args,
	}
	this.InvokeParam = c
	return this.Invoke()
}
