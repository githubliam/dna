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
package ontid

import (
	"encoding/hex"
	"errors"

	com "github.com/dnaproject2/DNA/common"
	"github.com/dnaproject2/DNA/core/states"
	"github.com/dnaproject2/DNA/core/types"
	"github.com/dnaproject2/DNA/smartcontract/service/native"
	"github.com/dnaproject2/DNA/smartcontract/service/native/utils"
	"github.com/ontio/ontology-crypto/keypair"
)

func checkIDExistence(srvc *native.NativeService, encID []byte) bool {
	val, err := srvc.CacheDB.Get(encID)
	if err == nil {
		val, err := states.GetValueFromRawStorageItem(val)
		if err == nil {
			if len(val) > 0 && val[0] == flag_exist {
				return true
			}
		}
	}
	return false
}

const (
	flag_exist = 0x01

	FIELD_VERSION byte = 0
	FLAG_VERSION  byte = 0x01

	FIELD_PK       byte = 1
	FIELD_ATTR     byte = 2
	FIELD_RECOVERY byte = 3
)

func encodeID(id []byte) ([]byte, error) {
	length := len(id)
	if length == 0 || length > 255 {
		return nil, errors.New("encode ID error: invalid ID length")
	}
	//enc := []byte{byte(length)}
	enc := append(utils.OntIDContractAddress[:], byte(length))
	enc = append(enc, id...)
	return enc, nil
}

func decodeID(data []byte) ([]byte, error) {
	prefix := len(utils.OntIDContractAddress)
	size := len(data)
	if size < prefix || size != int(data[prefix])+1+prefix {
		return nil, errors.New("decode ID error: invalid data length")
	}
	return data[prefix+1:], nil
}

func setRecovery(srvc *native.NativeService, encID []byte, recovery com.Address) error {
	key := append(encID, FIELD_RECOVERY)
	val := states.StorageItem{Value: recovery[:]}
	srvc.CacheDB.Put(key, val.ToArray())
	return nil
}

func getRecovery(srvc *native.NativeService, encID []byte) ([]byte, error) {
	key := append(encID, FIELD_RECOVERY)
	item, err := utils.GetStorageItem(srvc, key)
	if err != nil {
		return nil, errors.New("get recovery error: " + err.Error())
	} else if item == nil {
		return nil, nil
	}
	return item.Value, nil
}

func checkWitness(srvc *native.NativeService, key []byte) error {
	// try as if key is a public key
	pk, err := keypair.DeserializePublicKey(key)
	if err == nil {
		addr := types.AddressFromPubKey(pk)
		if srvc.ContextRef.CheckWitness(addr) {
			return nil
		}
	}

	// try as if key is an address
	addr, err := com.AddressParseFromBytes(key)
	if srvc.ContextRef.CheckWitness(addr) {
		return nil
	}

	return errors.New("check witness failed, " + hex.EncodeToString(key))
}
