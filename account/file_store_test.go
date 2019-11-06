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
package account

import (
	"encoding/hex"
	"os"
	"sort"
	"testing"

	"github.com/dnaproject2/DNA/core/types"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/stretchr/testify/assert"
)

func genAccountData() (*AccountData, *keypair.ProtectedKey) {
	var acc = new(AccountData)
	prvkey, pubkey, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	ta := types.AddressFromPubKey(pubkey)
	address := ta.ToBase58()
	password := []byte("123456")
	prvSectet, _ := keypair.EncryptPrivateKey(prvkey, address, password)
	acc.SetKeyPair(prvSectet)
	acc.SigSch = "SHA256withECDSA"
	acc.PubKey = hex.EncodeToString(keypair.SerializePublicKey(pubkey))
	return acc, prvSectet
}

func TestAccountData(t *testing.T) {
	acc, prvSectet := genAccountData()
	assert.NotNil(t, acc)
	assert.Equal(t, acc.Address, acc.ProtectedKey.Address)
	assert.Equal(t, prvSectet, acc.GetKeyPair())
}

func TestExecutorSave(t *testing.T) {
	executorFile := "w.data"
	defer func() {
		os.Remove(executorFile)
		os.RemoveAll("Log/")
	}()

	executor := NewExecutorData()
	size := 10
	for i := 0; i < size; i++ {
		acc, _ := genAccountData()
		executor.AddAccount(acc)
		err := executor.Save(executorFile)
		if err != nil {
			t.Errorf("Save error:%s", err)
			return
		}
	}

	executor2 := NewExecutorData()
	err := executor2.Load(executorFile)
	if err != nil {
		t.Errorf("Load error:%s", err)
		return
	}

	assert.Equal(t, len(executor2.Accounts), len(executor.Accounts))
}

func TestExecutorDel(t *testing.T) {
	executor := NewExecutorData()
	size := 10
	accList := make([]string, 0, size)
	for i := 0; i < size; i++ {
		acc, _ := genAccountData()
		executor.AddAccount(acc)
		accList = append(accList, acc.Address)
	}
	sort.Strings(accList)
	for _, address := range accList {
		executor.DelAccount(address)
		_, index := executor.GetAccountByAddress(address)
		if !assert.Equal(t, -1, index) {
			return
		}
	}
}
