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
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	testExecutor     Client
	testExecutorPath = "./executor_test.dat"
	testPasswd       = []byte("123456")
)

func TestMain(t *testing.M) {
	var err error
	testExecutor, err = Open(testExecutorPath)
	if err != nil {
		fmt.Printf("Open executor:%s error:%s\n", testExecutorPath, err)
		return
	}
	t.Run()
	os.Remove(testExecutorPath)
	os.Remove("ActorLog")
}

func TestClientNewAccount(t *testing.T) {
	accountNum := testExecutor.GetAccountNum()
	label1 := "t1"
	acc1, err := testExecutor.NewAccount(label1, keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, testPasswd)
	if err != nil {
		t.Errorf("TestClientNewAccount error:%s", err)
		return
	}
	label2 := "t2"
	acc2, err := testExecutor.NewAccount(label2, keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, testPasswd)
	if err != nil {
		t.Errorf("TestClientNewAccount error:%s", err)
		return
	}
	if accountNum+2 != testExecutor.GetAccountNum() {
		t.Errorf("TestClientNewAccount account num:%d != %d", testExecutor.GetAccountNum(), accountNum+2)
		return
	}
	accTmp, err := testExecutor.GetAccountByAddress(acc1.Address.ToBase58(), testPasswd)
	if err != nil {
		t.Errorf("TestClientNewAccount GetAccountByAddress:%s error:%s", acc1.Address.ToBase58(), err)
		return
	}
	if accTmp.Address.ToBase58() != acc1.Address.ToBase58() {
		t.Errorf("TestClientNewAccount by address address:%s != %s", accTmp.Address.ToBase58(), acc1.Address.ToBase58())
		return
	}
	accTmp, err = testExecutor.GetAccountByIndex(accountNum+1, testPasswd)
	if err != nil {
		t.Errorf("")
	}
	if accTmp.Address.ToBase58() != acc1.Address.ToBase58() {
		t.Errorf("TestClientNewAccount by index address:%s != %s", accTmp.Address.ToBase58(), acc1.Address.ToBase58())
		return
	}
	accTmp, err = testExecutor.GetAccountByLabel(label1, testPasswd)
	if err != nil {
		t.Errorf("TestClientNewAccount GetAccountByLabel:%s error:%s", label1, err)
		return
	}
	if accTmp.Address.ToBase58() != acc1.Address.ToBase58() {
		t.Errorf("TestClientNewAccount by label address:%s != %s", accTmp.Address.ToBase58(), acc1.Address.ToBase58())
		return
	}

	testExecutor2, err := Open(testExecutorPath)
	if err != nil {
		t.Errorf("NewAccount Open executor:%s error:%s", testExecutorPath, err)
		return
	}

	if testExecutor.GetAccountNum() != testExecutor2.GetAccountNum() {
		t.Errorf("TestClientNewAccount  AccountNum:%d != %d", testExecutor2.GetAccountNum(), testExecutor.GetAccountNum())
		return
	}

	accTmp, err = testExecutor2.GetAccountByLabel(label2, testPasswd)
	if err != nil {
		t.Errorf("TestClientNewAccount GetAccountByLabel:%s error:%s", label2, err)
		return
	}

	if accTmp.Address.ToBase58() != acc2.Address.ToBase58() {
		t.Errorf("TestClientNewAccount reopen address:%s != %s", accTmp.Address.ToBase58(), acc2.Address.ToBase58())
		return
	}

	_, err = testExecutor.NewAccount(label2, keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, testPasswd)
	if err == nil {
		t.Errorf("TestClientNewAccount new account with duplicate label:%s should failed", label2)
		return
	}
}

func TestClientDeleteAccount(t *testing.T) {
	accountNum := testExecutor.GetAccountNum()
	accSize := 10
	for i := 0; i < accSize; i++ {
		_, err := testExecutor.NewAccount("", keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, testPasswd)
		if err != nil {
			t.Errorf("TestClientDeleteAccount NewAccount error:%s", err)
			return
		}
	}
	delIndex := accountNum + 3
	delAcc, err := testExecutor.GetAccountByIndex(delIndex, testPasswd)
	if err != nil {
		t.Errorf("TestClientDeleteAccount GetAccountByIndex:%d error:%s", delIndex, err)
		return
	}
	if delAcc == nil {
		t.Errorf("TestClientDeleteAccount cannot getaccount by index:%d", delIndex)
		return
	}

	accountNum += accSize
	delAccTmp, err := testExecutor.DeleteAccount(delAcc.Address.ToBase58(), testPasswd)
	if err != nil {
		t.Errorf("TestClientDeleteAccount DeleteAccount error:%s", err)
		return
	}
	if delAcc.Address.ToBase58() != delAccTmp.Address.ToBase58() {
		t.Errorf("TestClientDeleteAccount Account address %s != %s", delAcc.Address.ToBase58(), delAccTmp.Address.ToBase58())
		return
	}
	if testExecutor.GetAccountNum() != accountNum-1 {
		t.Errorf("TestClientDeleteAccount AccountNum:%d != %d", testExecutor.GetAccountNum(), accountNum-1)
		return
	}
	accTmp, err := testExecutor.GetAccountByAddress(delAcc.Address.ToBase58(), testPasswd)
	if err != nil {
		t.Errorf("TestClientDeleteAccount GetAccountByAddress:%s error:%s", delAcc.Address.ToBase58(), err)
		return
	}
	if accTmp != nil {
		t.Errorf("TestClientDeleteAccount GetAccountByAddress:%s should return nil", delAcc.Address.ToBase58())
		return
	}
}

func TestClientSetLabel(t *testing.T) {
	accountNum := testExecutor.GetAccountNum()
	accountSize := 10
	if accountNum < accountSize {
		for i := accountSize - accountNum; i > accountNum; i-- {
			_, err := testExecutor.NewAccount("", keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, testPasswd)
			if err != nil {
				t.Errorf("TestClientSetLabel NewAccount error:%s", err)
				return
			}
		}
	}
	testAccIndex := 5
	testAcc := testExecutor.GetAccountMetadataByIndex(testAccIndex)
	oldLabel := testAcc.Label
	newLabel := fmt.Sprintf("%s-%d", oldLabel, testAccIndex)

	accountNum = testExecutor.GetAccountNum()
	err := testExecutor.SetLabel(testAcc.Address, newLabel)
	if err != nil {
		t.Errorf("TestClientSetLabel SetLabel error:%s", err)
		return
	}

	if testExecutor.GetAccountNum() != accountNum {
		t.Errorf("TestClientSetLabel account num %d != %d", testExecutor.GetAccountNum(), accountNum)
		return
	}

	accTmp, err := testExecutor.GetAccountByLabel(newLabel, testPasswd)
	if err != nil {
		t.Errorf("TestClientSetLabel GetAccountByLabel:%s error:%s", newLabel, err)
		return
	}
	if accTmp == nil {
		t.Errorf("TestClientSetLabel cannot get account by label:%s", newLabel)
		return
	}

	if accTmp.Address.ToBase58() != testAcc.Address {
		t.Errorf("TestClientSetLabel address:%s != %s", accTmp.Address.ToBase58(), testAcc.Address)
		return
	}

	accTmp, err = testExecutor.GetAccountByLabel(oldLabel, testPasswd)
	if err != nil {
		t.Errorf("TestClientSetLabel GetAccountByLabel:%s error:%s", oldLabel, err)
		return
	}
	if accTmp != nil {
		t.Errorf("TestClientSetLabel GetAccountByLabel:%s should return nil", oldLabel)
		return
	}
}

func TestClientSetDefault(t *testing.T) {
	accountNum := testExecutor.GetAccountNum()
	accountSize := 10
	if accountNum < accountSize {
		for i := accountSize - accountNum; i > accountNum; i-- {
			_, err := testExecutor.NewAccount("", keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, testPasswd)
			if err != nil {
				t.Errorf("TestClientSetDefault NewAccount error:%s", err)
				return
			}
		}
	}
	testAccIndex := 5
	testAcc, err := testExecutor.GetAccountByIndex(testAccIndex, testPasswd)
	if err != nil {
		t.Errorf("TestClientSetDefault GetAccountByIndex:%d error:%s", testAccIndex, err)
		return
	}

	oldDefAcc, err := testExecutor.GetDefaultAccount(testPasswd)
	if err != nil {
		t.Errorf("TestClientSetDefault GetDefaultAccount error:%s", err)
		return
	}
	if oldDefAcc == nil {
		t.Errorf("TestClientSetDefault GetDefaultAccount return nil")
		return
	}

	err = testExecutor.SetDefaultAccount(testAcc.Address.ToBase58())
	if err != nil {
		t.Errorf("TestClientSetDefault SetDefaultAccount error:%s", err)
		return
	}

	defAcc, err := testExecutor.GetDefaultAccount(testPasswd)
	if err != nil {
		t.Errorf("TestClientSetDefault GetDefaultAccount error:%s", err)
		return
	}
	if defAcc == nil {
		t.Errorf("TestClientSetDefault GetDefaultAccount return nil")
		return
	}

	if defAcc.Address.ToBase58() != testAcc.Address.ToBase58() {
		t.Errorf("TestClientSetDefault address %s != %s", defAcc.Address.ToBase58(), testAcc.Address.ToBase58())
		return
	}

	accTmp := testExecutor.GetAccountMetadataByAddress(oldDefAcc.Address.ToBase58())
	if accTmp.IsDefault {
		t.Errorf("TestClientSetDefault address:%s should not default account", accTmp.Address)
		return
	}

	accTmp = testExecutor.GetAccountMetadataByAddress(testAcc.Address.ToBase58())
	if !accTmp.IsDefault {
		t.Errorf("TestClientSetDefault address:%s should be default account", accTmp.Address)
		return
	}
}

func TestImportAccount(t *testing.T) {
	executorPath2 := "tmp.dat"
	executor2, err := NewClientImpl(executorPath2)
	if err != nil {
		t.Errorf("TestImportAccount NewClientImpl error:%s", err)
		return
	}
	defer os.Remove(executorPath2)

	acc1, err := executor2.NewAccount("", keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, testPasswd)
	if err != nil {
		t.Errorf("TestImportAccount NewAccount error:%s", err)
		return
	}
	accMetadata := executor2.GetAccountMetadataByAddress(acc1.Address.ToBase58())
	if accMetadata == nil {
		t.Errorf("TestImportAccount GetAccountMetadataByAddress:%s return nil", acc1.Address.ToBase58())
		return
	}
	err = testExecutor.ImportAccount(accMetadata)
	if err != nil {
		t.Errorf("TestImportAccount ImportAccount error:%s", err)
		return
	}

	acc, err := testExecutor.GetAccountByAddress(accMetadata.Address, testPasswd)
	if err != nil {
		t.Errorf("TestImportAccount GetAccountByAddress error:%s", err)
		return
	}
	if acc == nil {
		t.Errorf("TestImportAccount failed, GetAccountByAddress return nil after import")
		return
	}
	assert.Equal(t, acc.Address.ToBase58() == acc1.Address.ToBase58(), true)
}

func TestCheckSigScheme(t *testing.T) {
	testClient, _ := NewClientImpl("")

	assert.Equal(t, testClient.checkSigScheme("ECDSA", "SHA224withECDSA"), true)
	assert.Equal(t, testClient.checkSigScheme("ECDSA", "SM3withSM2"), false)
	assert.Equal(t, testClient.checkSigScheme("SM2", "SM3withSM2"), true)
	assert.Equal(t, testClient.checkSigScheme("SM2", "SHA224withECDSA"), false)
	assert.Equal(t, testClient.checkSigScheme("Ed25519", "SHA512withEdDSA"), true)
	assert.Equal(t, testClient.checkSigScheme("Ed25519", "SHA224withECDSA"), false)
}
