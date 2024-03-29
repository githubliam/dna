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
package smartcontract

import (
	"fmt"

	"github.com/dnaproject2/DNA/common"
	"github.com/dnaproject2/DNA/common/log"
	"github.com/dnaproject2/DNA/core/store"
	ctypes "github.com/dnaproject2/DNA/core/types"
	"github.com/dnaproject2/DNA/smartcontract/context"
	"github.com/dnaproject2/DNA/smartcontract/event"
	"github.com/dnaproject2/DNA/smartcontract/service/native"
	"github.com/dnaproject2/DNA/smartcontract/service/neovm"
	"github.com/dnaproject2/DNA/smartcontract/storage"
	vm "github.com/dnaproject2/DNA/vm/neovm"
)

const (
	MAX_EXECUTE_ENGINE = 1024
)

// SmartContract describe smart contract execute engine
type SmartContract struct {
	Contexts      []*context.Context // all execute smart contract context
	CacheDB       *storage.CacheDB   // state cache
	Store         store.LedgerStore  // ledger store
	Config        *Config
	Notifications []*event.NotifyEventInfo // all execute smart contract event notify info
	Gas           uint64
	ExecStep      int
	PreExec       bool
}

// Config describe smart contract need parameters configuration
type Config struct {
	Time      uint32              // current block timestamp
	Height    uint32              // current block height
	BlockHash common.Uint256      // current block hash
	Tx        *ctypes.Transaction // current transaction
}

// PushContext push current context to smart contract
func (this *SmartContract) PushContext(context *context.Context) {
	this.Contexts = append(this.Contexts, context)
}

// CurrentContext return smart contract current context
func (this *SmartContract) CurrentContext() *context.Context {
	if len(this.Contexts) < 1 {
		return nil
	}
	return this.Contexts[len(this.Contexts)-1]
}

// CallingContext return smart contract caller context
func (this *SmartContract) CallingContext() *context.Context {
	if len(this.Contexts) < 2 {
		return nil
	}
	return this.Contexts[len(this.Contexts)-2]
}

// EntryContext return smart contract entry entrance context
func (this *SmartContract) EntryContext() *context.Context {
	if len(this.Contexts) < 1 {
		return nil
	}
	return this.Contexts[0]
}

// PopContext pop smart contract current context
func (this *SmartContract) PopContext() {
	if len(this.Contexts) > 1 {
		this.Contexts = this.Contexts[:len(this.Contexts)-1]
	}
}

// PushNotifications push smart contract event info
func (this *SmartContract) PushNotifications(notifications []*event.NotifyEventInfo) {
	this.Notifications = append(this.Notifications, notifications...)
}

func (this *SmartContract) CheckExecStep() bool {
	if this.ExecStep >= neovm.VM_STEP_LIMIT {
		return false
	}
	this.ExecStep += 1
	return true
}

func (this *SmartContract) CheckUseGas(gas uint64) bool {
	if this.Gas < gas {
		return false
	}
	this.Gas -= gas
	return true
}

func (this *SmartContract) checkContexts() bool {
	if len(this.Contexts) > MAX_EXECUTE_ENGINE {
		return false
	}
	return true
}

// Execute is smart contract execute manager
// According different vm type to launch different service
func (this *SmartContract) NewExecuteEngine(code []byte) (context.Engine, error) {
	if !this.checkContexts() {
		return nil, fmt.Errorf("%s", "engine over max limit!")
	}
	service := &neovm.NeoVmService{
		Store:      this.Store,
		CacheDB:    this.CacheDB,
		ContextRef: this,
		Code:       code,
		Tx:         this.Config.Tx,
		Time:       this.Config.Time,
		Height:     this.Config.Height,
		BlockHash:  this.Config.BlockHash,
		Engine:     vm.NewExecutionEngine(),
		PreExec:    this.PreExec,
	}
	return service, nil
}

func (this *SmartContract) NewNativeService() (*native.NativeService, error) {
	if !this.checkContexts() {
		return nil, fmt.Errorf("%s", "engine over max limit!")
	}
	service := &native.NativeService{
		CacheDB:    this.CacheDB,
		ContextRef: this,
		Tx:         this.Config.Tx,
		Time:       this.Config.Time,
		Height:     this.Config.Height,
		BlockHash:  this.Config.BlockHash,
		ServiceMap: make(map[string]native.Handler),
	}
	return service, nil
}

// CheckWitness check whether authorization correct
// If address is executor address, check whether in the signature addressed list
// Else check whether address is calling contract address
// Param address: executor address or contract address
func (this *SmartContract) CheckWitness(address common.Address) bool {
	if this.checkAccountAddress(address) || this.checkContractAddress(address) {
		return true
	}
	return false
}

func (this *SmartContract) checkAccountAddress(address common.Address) bool {
	addresses, err := this.Config.Tx.GetSignatureAddresses()
	if err != nil {
		log.Errorf("get signature address error:%v", err)
		return false
	}
	for _, v := range addresses {
		if v == address {
			return true
		}
	}
	return false
}

func (this *SmartContract) checkContractAddress(address common.Address) bool {
	if this.CallingContext() != nil && this.CallingContext().ContractAddress == address {
		return true
	}
	return false
}
