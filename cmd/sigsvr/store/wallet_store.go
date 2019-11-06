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
package store

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dnaproject2/DNA/account"
	"github.com/dnaproject2/DNA/core/types"
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"sync"
)

type ExecutorStore struct {
	ExecutorName     string
	ExecutorVersion  string
	ExecutorScrypt   *keypair.ScryptParam
	ExecutorExtra    string
	path             string
	db               *leveldb.DB
	nextAccountIndex uint32
	lock             sync.RWMutex
}

func NewExecutorStore(path string) (*ExecutorStore, error) {
	lvlOpts := &opt.Options{
		NoSync: false,
		Filter: filter.NewBloomFilter(10),
	}
	db, err := leveldb.OpenFile(path, lvlOpts)
	if err != nil {
		return nil, err
	}
	executorStore := &ExecutorStore{
		path: path,
		db:   db,
	}

	init, err := executorStore.isInit()
	if err != nil {
		return nil, err
	}
	if !init {
		executorStore.ExecutorName = DEFAULT_WALLET_NAME
		err = executorStore.setExecutorName(executorStore.ExecutorName)
		if err != nil {
			return nil, fmt.Errorf("setExecutorName error:%s", err)
		}
		executorStore.ExecutorVersion = WALLET_VERSION
		err = executorStore.setExecutorVersion(executorStore.ExecutorVersion)
		if err != nil {
			return nil, fmt.Errorf("setExecutorVersion error:%s", err)
		}
		executorStore.ExecutorScrypt = keypair.GetScryptParameters()
		err = executorStore.setExecutorScrypt(executorStore.ExecutorScrypt)
		if err != nil {
			return nil, fmt.Errorf("setExecutorScrypt error:%s", err)
		}
		executorStore.ExecutorExtra = ""
		err = executorStore.setExecutorExtra(executorStore.ExecutorExtra)
		if err != nil {
			return nil, fmt.Errorf("setExecutorExtra error:%s", err)
		}
		err = executorStore.init()
		if err != nil {
			return nil, fmt.Errorf("init error:%s", err)
		}
		return executorStore, nil
	}
	nextAccountIndex, err := executorStore.getNextAccountIndex()
	if err != nil {
		return nil, fmt.Errorf("getNextAccountIndex error:%s", err)
	}
	executorName, err := executorStore.getExecutorName()
	if err != nil {
		return nil, fmt.Errorf("getExecutorName error:%s", err)
	}
	executorVersion, err := executorStore.getExecutorVersion()
	if err != nil {
		return nil, fmt.Errorf("getExecutorVersion error:%s", err)
	}
	executorScrypt, err := executorStore.getExecutorScrypt()
	if err != nil {
		return nil, fmt.Errorf("getExecutorScrypt error:%s", err)
	}
	executorExtra, err := executorStore.getExecutorExtra()
	if err != nil {
		return nil, fmt.Errorf("getExecutorExtra error: %v", err)
	}
	executorStore.nextAccountIndex = nextAccountIndex
	executorStore.ExecutorScrypt = executorScrypt
	executorStore.ExecutorName = executorName
	executorStore.ExecutorVersion = executorVersion
	executorStore.ExecutorExtra = executorExtra
	return executorStore, nil
}

func (this *ExecutorStore) isInit() (bool, error) {
	data, err := this.db.Get(GetExecutorInitKey(), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	if string(data) != WALLET_INIT_DATA {
		return false, fmt.Errorf("init not success")
	}
	return true, nil
}

func (this *ExecutorStore) init() error {
	return this.db.Put(GetExecutorInitKey(), []byte(WALLET_INIT_DATA), nil)
}

func (this *ExecutorStore) setExecutorVersion(version string) error {
	return this.db.Put(GetExecutorVersionKey(), []byte(version), nil)
}

func (this *ExecutorStore) getExecutorVersion() (string, error) {
	data, err := this.db.Get(GetExecutorVersionKey(), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (this *ExecutorStore) setExecutorName(name string) error {
	return this.db.Put(GetExecutorNameKey(), []byte(name), nil)
}

func (this *ExecutorStore) getExecutorName() (string, error) {
	data, err := this.db.Get(GetExecutorNameKey(), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (this *ExecutorStore) setExecutorScrypt(scrypt *keypair.ScryptParam) error {
	data, err := json.Marshal(scrypt)
	if err != nil {
		return err
	}
	return this.db.Put(GetExecutorScryptKey(), data, nil)
}

func (this *ExecutorStore) getExecutorScrypt() (*keypair.ScryptParam, error) {
	data, err := this.db.Get(GetExecutorScryptKey(), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	scypt := &keypair.ScryptParam{}
	err = json.Unmarshal(data, scypt)
	if err != nil {
		return nil, err
	}
	return scypt, nil
}

func (this *ExecutorStore) setExecutorExtra(extra string) error {
	return this.db.Put(GetExecutorExtraKey(), []byte(extra), nil)
}

func (this *ExecutorStore) getExecutorExtra() (string, error) {
	data, err := this.db.Get(GetExecutorExtraKey(), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (this *ExecutorStore) GetNextAccountIndex() uint32 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.nextAccountIndex
}

func (this *ExecutorStore) GetAccountByAddress(address string, passwd []byte) (*account.Account, error) {
	accData, err := this.GetAccountDataByAddress(address)
	if err != nil {
		return nil, err
	}
	if accData == nil {
		return nil, nil
	}
	privateKey, err := keypair.DecryptWithCustomScrypt(&accData.ProtectedKey, passwd, this.ExecutorScrypt)
	if err != nil {
		return nil, fmt.Errorf("decrypt PrivateKey error:%s", err)
	}
	publicKey := privateKey.Public()
	addr := types.AddressFromPubKey(publicKey)
	scheme, err := s.GetScheme(accData.SigSch)
	if err != nil {
		return nil, fmt.Errorf("signature scheme error:%s", err)
	}
	return &account.Account{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    addr,
		SigScheme:  scheme,
	}, nil
}

func (this *ExecutorStore) NewAccountData(typeCode keypair.KeyType, curveCode byte, sigScheme s.SignatureScheme, passwd []byte) (*account.AccountData, error) {
	if len(passwd) == 0 {
		return nil, fmt.Errorf("password cannot empty")
	}
	prvkey, pubkey, err := keypair.GenerateKeyPair(typeCode, curveCode)
	if err != nil {
		return nil, fmt.Errorf("generateKeyPair error:%s", err)
	}
	address := types.AddressFromPubKey(pubkey)
	addressBase58 := address.ToBase58()
	prvSecret, err := keypair.EncryptWithCustomScrypt(prvkey, addressBase58, passwd, this.ExecutorScrypt)
	if err != nil {
		return nil, fmt.Errorf("encryptPrivateKey error:%s", err)
	}
	accData := &account.AccountData{}
	accData.SetKeyPair(prvSecret)
	accData.SigSch = sigScheme.Name()
	accData.PubKey = hex.EncodeToString(keypair.SerializePublicKey(pubkey))

	return accData, nil
}

func (this *ExecutorStore) AddAccountData(accData *account.AccountData) (bool, error) {
	isExist, err := this.IsAccountExist(accData.Address)
	if err != nil {
		return false, err
	}

	this.lock.Lock()
	defer this.lock.Unlock()

	accountNum, err := this.GetAccountNumber()
	if err != nil {
		return false, fmt.Errorf("GetAccountNumber error:%s", err)
	}
	if accountNum == 0 {
		accData.IsDefault = true
	} else {
		accData.IsDefault = false
	}
	data, err := json.Marshal(accData)
	if err != nil {
		return false, err
	}

	batch := &leveldb.Batch{}
	//Put account
	batch.Put(GetAccountKey(accData.Address), data)

	nextIndex := this.nextAccountIndex
	if !isExist {
		accountIndex := nextIndex
		//Put account index
		batch.Put(GetAccountIndexKey(accountIndex), []byte(accData.Address))

		nextIndex++
		data = make([]byte, 4, 4)
		binary.LittleEndian.PutUint32(data, nextIndex)

		//Put next account index
		batch.Put(GetNextAccountIndexKey(), data)

		accountNum++
		binary.LittleEndian.PutUint32(data, accountNum)

		//Put account number
		batch.Put(GetExecutorAccountNumberKey(), data)
	}

	err = this.db.Write(batch, nil)
	if err != nil {
		return false, err
	}
	this.nextAccountIndex = nextIndex

	isAdd := !isExist
	return isAdd, nil
}

func (this *ExecutorStore) getNextAccountIndex() (uint32, error) {
	data, err := this.db.Get(GetNextAccountIndexKey(), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}
	return binary.LittleEndian.Uint32(data), nil
}

func (this *ExecutorStore) GetAccountDataByAddress(address string) (*account.AccountData, error) {
	data, err := this.db.Get(GetAccountKey(address), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	accData := &account.AccountData{}
	err = json.Unmarshal(data, accData)
	if err != nil {
		return nil, err
	}
	return accData, nil
}

func (this *ExecutorStore) IsAccountExist(address string) (bool, error) {
	data, err := this.db.Get(GetAccountKey(address), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return len(data) != 0, nil
}

func (this *ExecutorStore) GetAccountDataByIndex(index uint32) (*account.AccountData, error) {
	address, err := this.GetAccountAddress(index)
	if err != nil {
		return nil, err
	}
	if address == "" {
		return nil, nil
	}
	return this.GetAccountDataByAddress(address)
}

func (this *ExecutorStore) GetAccountAddress(index uint32) (string, error) {
	data, err := this.db.Get(GetAccountIndexKey(index), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (this *ExecutorStore) setAccountNumber(number uint32) error {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data, number)
	return this.db.Put(GetExecutorAccountNumberKey(), data, nil)
}

func (this *ExecutorStore) GetAccountNumber() (uint32, error) {
	data, err := this.db.Get(GetExecutorAccountNumberKey(), nil)
	if err == nil {
		return binary.LittleEndian.Uint32(data), nil
	}
	if err != leveldb.ErrNotFound {
		return 0, err
	}
	//Keep downward compatible
	nextIndex, err := this.getNextAccountIndex()
	if err != nil {
		return 0, fmt.Errorf("getNextAccountIndex error:%s", err)
	}
	if nextIndex == 0 {
		return 0, nil
	}
	addresses := make(map[string]string, 0)
	for i := uint32(0); i < nextIndex; i++ {
		address, err := this.GetAccountAddress(i)
		if err != nil {
			return 0, fmt.Errorf("GetAccountAddress Index:%d error:%s", i, err)
		}
		if address == "" {
			continue
		}
		addresses[address] = ""
	}
	accNum := uint32(len(addresses))
	err = this.setAccountNumber(accNum)
	if err != nil {
		return 0, fmt.Errorf("setAccountNumber error")
	}
	return accNum, nil
}
