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

package types

import (
	"io"

	"github.com/dnaproject2/DNA/common"
	comm "github.com/dnaproject2/DNA/p2pserver/common"
)

type DataReq struct {
	DataType common.InventoryType
	Hash     common.Uint256
}

//Serialize message payload
func (this DataReq) Serialization(sink *common.ZeroCopySink) {
	sink.WriteByte(byte(this.DataType))
	sink.WriteHash(this.Hash)
}

func (this *DataReq) CmdType() string {
	return comm.GET_DATA_TYPE
}

//Deserialize message payload
func (this *DataReq) Deserialization(source *common.ZeroCopySource) error {
	ty, eof := source.NextByte()
	this.DataType = common.InventoryType(ty)

	this.Hash, eof = source.NextHash()
	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}
