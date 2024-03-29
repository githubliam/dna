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

package neovm

import (
	"bytes"
	"fmt"

	scommon "github.com/dnaproject2/DNA/common"
	"github.com/dnaproject2/DNA/common/log"
	"github.com/dnaproject2/DNA/core/signature"
	"github.com/dnaproject2/DNA/core/store"
	"github.com/dnaproject2/DNA/core/types"
	"github.com/dnaproject2/DNA/errors"
	"github.com/dnaproject2/DNA/smartcontract/context"
	"github.com/dnaproject2/DNA/smartcontract/event"
	"github.com/dnaproject2/DNA/smartcontract/storage"
	vm "github.com/dnaproject2/DNA/vm/neovm"
	ntypes "github.com/dnaproject2/DNA/vm/neovm/types"
	"github.com/ontio/ontology-crypto/keypair"
)

var (
	// Register all service for smart contract execute
	ServiceMap = map[string]Service{
		ATTRIBUTE_GETUSAGE_NAME:              {Execute: AttributeGetUsage, Validator: validatorAttribute},
		ATTRIBUTE_GETDATA_NAME:               {Execute: AttributeGetData, Validator: validatorAttribute},
		BLOCK_GETTRANSACTIONCOUNT_NAME:       {Execute: BlockGetTransactionCount, Validator: validatorBlock},
		BLOCK_GETTRANSACTIONS_NAME:           {Execute: BlockGetTransactions, Validator: validatorBlock},
		BLOCK_GETTRANSACTION_NAME:            {Execute: BlockGetTransaction, Validator: validatorBlockTransaction},
		BLOCKCHAIN_GETHEIGHT_NAME:            {Execute: BlockChainGetHeight},
		BLOCKCHAIN_GETHEADER_NAME:            {Execute: BlockChainGetHeader, Validator: validatorBlockChainHeader},
		BLOCKCHAIN_GETBLOCK_NAME:             {Execute: BlockChainGetBlock, Validator: validatorBlockChainBlock},
		BLOCKCHAIN_GETTRANSACTION_NAME:       {Execute: BlockChainGetTransaction, Validator: validatorBlockChainTransaction},
		BLOCKCHAIN_GETCONTRACT_NAME:          {Execute: BlockChainGetContract, Validator: validatorBlockChainContract},
		BLOCKCHAIN_GETTRANSACTIONHEIGHT_NAME: {Execute: BlockChainGetTransactionHeight},
		HEADER_GETINDEX_NAME:                 {Execute: HeaderGetIndex, Validator: validatorHeader},
		HEADER_GETHASH_NAME:                  {Execute: HeaderGetHash, Validator: validatorHeader},
		HEADER_GETVERSION_NAME:               {Execute: HeaderGetVersion, Validator: validatorHeader},
		HEADER_GETPREVHASH_NAME:              {Execute: HeaderGetPrevHash, Validator: validatorHeader},
		HEADER_GETTIMESTAMP_NAME:             {Execute: HeaderGetTimestamp, Validator: validatorHeader},
		HEADER_GETCONSENSUSDATA_NAME:         {Execute: HeaderGetConsensusData, Validator: validatorHeader},
		HEADER_GETNEXTCONSENSUS_NAME:         {Execute: HeaderGetNextConsensus, Validator: validatorHeader},
		HEADER_GETMERKLEROOT_NAME:            {Execute: HeaderGetMerkleRoot, Validator: validatorHeader},
		TRANSACTION_GETHASH_NAME:             {Execute: TransactionGetHash, Validator: validatorTransaction},
		TRANSACTION_GETTYPE_NAME:             {Execute: TransactionGetType, Validator: validatorTransaction},
		TRANSACTION_GETATTRIBUTES_NAME:       {Execute: TransactionGetAttributes, Validator: validatorTransaction},
		CONTRACT_CREATE_NAME:                 {Execute: ContractCreate},
		CONTRACT_MIGRATE_NAME:                {Execute: ContractMigrate},
		CONTRACT_GETSTORAGECONTEXT_NAME:      {Execute: ContractGetStorageContext},
		CONTRACT_DESTROY_NAME:                {Execute: ContractDestory},
		CONTRACT_GETSCRIPT_NAME:              {Execute: ContractGetCode, Validator: validatorGetCode},
		RUNTIME_GETTIME_NAME:                 {Execute: RuntimeGetTime},
		RUNTIME_CHECKWITNESS_NAME:            {Execute: RuntimeCheckWitness, Validator: validatorCheckWitness},
		RUNTIME_NOTIFY_NAME:                  {Execute: RuntimeNotify, Validator: validatorNotify},
		RUNTIME_LOG_NAME:                     {Execute: RuntimeLog, Validator: validatorLog},
		RUNTIME_GETTRIGGER_NAME:              {Execute: RuntimeGetTrigger},
		RUNTIME_SERIALIZE_NAME:               {Execute: RuntimeSerialize, Validator: validatorSerialize},
		RUNTIME_DESERIALIZE_NAME:             {Execute: RuntimeDeserialize, Validator: validatorDeserialize},
		NATIVE_INVOKE_NAME:                   {Execute: NativeInvoke},
		STORAGE_GET_NAME:                     {Execute: StorageGet},
		STORAGE_PUT_NAME:                     {Execute: StoragePut},
		STORAGE_DELETE_NAME:                  {Execute: StorageDelete},
		STORAGE_GETCONTEXT_NAME:              {Execute: StorageGetContext},
		STORAGE_GETREADONLYCONTEXT_NAME:      {Execute: StorageGetReadOnlyContext},
		STORAGECONTEXT_ASREADONLY_NAME:       {Execute: StorageContextAsReadOnly, Validator: validatorContextAsReadOnly},
		GETSCRIPTCONTAINER_NAME:              {Execute: GetCodeContainer},
		GETEXECUTINGSCRIPTHASH_NAME:          {Execute: GetExecutingAddress},
		GETCALLINGSCRIPTHASH_NAME:            {Execute: GetCallingAddress},
		GETENTRYSCRIPTHASH_NAME:              {Execute: GetEntryAddress},

		RUNTIME_BASE58TOADDRESS_NAME:     {Execute: RuntimeBase58ToAddress},
		RUNTIME_ADDRESSTOBASE58_NAME:     {Execute: RuntimeAddressToBase58},
		RUNTIME_GETCURRENTBLOCKHASH_NAME: {Execute: RuntimeGetCurrentBlockHash},
	}
)

var (
	ERR_CHECK_STACK_SIZE  = errors.NewErr("[NeoVmService] vm execution exceeded the max stack size!")
	ERR_EXECUTE_CODE      = errors.NewErr("[NeoVmService] vm execution code was invalid!")
	ERR_GAS_INSUFFICIENT  = errors.NewErr("[NeoVmService] insufficient gas for transaction!")
	VM_EXEC_STEP_EXCEED   = errors.NewErr("[NeoVmService] vm execution exceeded the step limit!")
	CONTRACT_NOT_EXIST    = errors.NewErr("[NeoVmService] the given contract does not exist!")
	DEPLOYCODE_TYPE_ERROR = errors.NewErr("[NeoVmService] deploy code type error!")
	VM_EXEC_FAULT         = errors.NewErr("[NeoVmService] vm execution encountered a state fault!")
)

var (
	BYTE_ZERO_20 = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

type (
	Execute   func(service *NeoVmService, engine *vm.ExecutionEngine) error
	Validator func(engine *vm.ExecutionEngine) error
)

type Service struct {
	Execute   Execute
	Validator Validator
}

// NeoVmService is a struct for smart contract provide interop service
type NeoVmService struct {
	Store         store.LedgerStore
	CacheDB       *storage.CacheDB
	ContextRef    context.ContextRef
	Notifications []*event.NotifyEventInfo
	Code          []byte
	Tx            *types.Transaction
	Time          uint32
	Height        uint32
	BlockHash     scommon.Uint256
	Engine        *vm.ExecutionEngine
	PreExec       bool
}

// Invoke a smart contract
func (this *NeoVmService) Invoke() (interface{}, error) {
	if len(this.Code) == 0 {
		return nil, ERR_EXECUTE_CODE
	}
	this.ContextRef.PushContext(&context.Context{ContractAddress: scommon.AddressFromVmCode(this.Code), Code: this.Code})
	this.Engine.PushContext(vm.NewExecutionContext(this.Engine, this.Code))
	for {
		//check the execution step count
		if this.PreExec && !this.ContextRef.CheckExecStep() {
			return nil, VM_EXEC_STEP_EXCEED
		}
		if len(this.Engine.Contexts) == 0 || this.Engine.Context == nil {
			break
		}
		if this.Engine.Context.GetInstructionPointer() >= len(this.Engine.Context.Code) {
			break
		}
		if err := this.Engine.ExecuteCode(); err != nil {
			return nil, err
		}
		if this.Engine.Context.GetInstructionPointer() < len(this.Engine.Context.Code) {
			if ok := checkStackSize(this.Engine); !ok {
				return nil, ERR_CHECK_STACK_SIZE
			}
		}
		if this.Engine.OpCode >= vm.PUSHBYTES1 && this.Engine.OpCode <= vm.PUSHBYTES75 {
			if !this.ContextRef.CheckUseGas(OPCODE_GAS) {
				return nil, ERR_GAS_INSUFFICIENT
			}
		} else {
			if err := this.Engine.ValidateOp(); err != nil {
				return nil, err
			}
			price, err := GasPrice(this.Engine, this.Engine.OpExec.Name)
			if err != nil {
				return nil, err
			}
			if !this.ContextRef.CheckUseGas(price) {
				return nil, ERR_GAS_INSUFFICIENT
			}
		}
		switch this.Engine.OpCode {
		case vm.VERIFY:
			if vm.EvaluationStackCount(this.Engine) < 3 {
				return nil, errors.NewErr("[VERIFY] too few input parameters")
			}
			pubKey, err := vm.PopByteArray(this.Engine)
			if err != nil {
				return nil, err
			}
			key, err := keypair.DeserializePublicKey(pubKey)
			if err != nil {
				return nil, err
			}
			sig, err := vm.PopByteArray(this.Engine)
			if err != nil {
				return nil, err
			}
			data, err := vm.PopByteArray(this.Engine)
			if err != nil {
				return nil, err
			}
			if err := signature.Verify(key, data, sig); err != nil {
				vm.PushData(this.Engine, false)
			} else {
				vm.PushData(this.Engine, true)
			}
		case vm.SYSCALL:
			if err := this.SystemCall(this.Engine); err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] service system call error!")
			}
		case vm.APPCALL:
			address, err := this.Engine.Context.OpReader.ReadBytes(20)
			if err != nil {
				return nil, fmt.Errorf("[Appcall] read contract address error: %v", err)
			}
			if bytes.Compare(address, BYTE_ZERO_20) == 0 {
				if vm.EvaluationStackCount(this.Engine) < 1 {
					return nil, fmt.Errorf("[Appcall] too few input parameters: %d", vm.EvaluationStackCount(this.Engine))
				}
				address, err = vm.PopByteArray(this.Engine)
				if err != nil {
					return nil, fmt.Errorf("[Appcall] pop contract address error: %v", err)
				}
				if len(address) != 20 {
					return nil, fmt.Errorf("[Appcall] pop contract address len != 20: %x", address)
				}
			}
			addr, err := scommon.AddressParseFromBytes(address)
			if err != nil {
				return nil, err
			}
			code, err := this.getContract(addr)
			if err != nil {
				return nil, err
			}
			service, err := this.ContextRef.NewExecuteEngine(code)
			if err != nil {
				return nil, err
			}
			this.Engine.EvaluationStack.CopyTo(service.(*NeoVmService).Engine.EvaluationStack)
			result, err := service.Invoke()
			if err != nil {
				return nil, err
			}
			if result != nil {
				vm.PushData(this.Engine, result)
			}
		default:
			if err := this.Engine.StepInto(); err != nil {
				return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[NeoVmService] vm execution error!")
			}
			if this.Engine.State == vm.FAULT {
				return nil, VM_EXEC_FAULT
			}
		}
	}
	this.ContextRef.PopContext()
	this.ContextRef.PushNotifications(this.Notifications)
	if this.Engine.EvaluationStack.Count() != 0 {
		return this.Engine.EvaluationStack.Peek(0), nil
	}
	return nil, nil
}

// SystemCall provide register service for smart contract to interaction with blockchain
func (this *NeoVmService) SystemCall(engine *vm.ExecutionEngine) error {
	serviceName, err := engine.Context.OpReader.ReadVarString(vm.MAX_BYTEARRAY_SIZE)
	if err != nil {
		return err
	}
	service, ok := ServiceMap[serviceName]
	if !ok {
		return errors.NewErr(fmt.Sprintf("[SystemCall] the given service is not supported: %s", serviceName))
	}
	if service.Validator != nil {
		if err := service.Validator(engine); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[SystemCall] there was a service validator error!")
		}
	}
	price, err := GasPrice(engine, serviceName)
	if err != nil {
		return err
	}
	if !this.ContextRef.CheckUseGas(price) {
		return ERR_GAS_INSUFFICIENT
	}
	if err := service.Execute(this, engine); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[SystemCall] service execution error!")
	}
	return nil
}

func (this *NeoVmService) getContract(address scommon.Address) ([]byte, error) {
	dep, err := this.CacheDB.GetContract(address)
	if err != nil {
		return nil, errors.NewErr("[getContract] get contract context error!")
	}
	log.Debugf("invoke contract address: %s", address.ToHexString())
	if dep == nil {
		return nil, CONTRACT_NOT_EXIST
	}
	return dep.Code, nil
}

func checkStackSize(engine *vm.ExecutionEngine) bool {
	size := 0
	if engine.OpCode < vm.PUSH16 {
		size = 1
	} else {
		switch engine.OpCode {
		case vm.DEPTH, vm.DUP, vm.OVER, vm.TUCK:
			size = 1
		case vm.UNPACK:
			if engine.EvaluationStack.Count() == 0 {
				return false
			}
			item := vm.PeekStackItem(engine)
			if a, ok := item.(*ntypes.Array); ok {
				size = a.Count()
			}
			if a, ok := item.(*ntypes.Struct); ok {
				size = a.Count()
			}
		}
	}
	size += engine.EvaluationStack.Count() + engine.AltStack.Count()
	if size > DUPLICATE_STACK_SIZE {
		return false
	}
	return true
}
