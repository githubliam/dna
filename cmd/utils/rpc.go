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

package utils

import (
	"encoding/json"
	"fmt"
	"github.com/dnaproject2/DNA/common/config"
	rpcerr "github.com/dnaproject2/DNA/http/base/error"
	"io/ioutil"
	"net/http"
	"strings"
)

//JsonRpc version
const JSON_RPC_VERSION = "2.0"

const (
	ERROR_INVALID_PARAMS = rpcerr.INVALID_PARAMS
	ERROR_DNA_COMMON     = 10000
	ERROR_DNA_SUCCESS    = 0
)

type DNAError struct {
	ErrorCode int64
	Error     error
}

func NewDNAError(err error, errCode ...int64) *DNAError {
	ontErr := &DNAError{Error: err}
	if len(errCode) > 0 {
		ontErr.ErrorCode = errCode[0]
	} else {
		ontErr.ErrorCode = ERROR_DNA_COMMON
	}
	if err == nil {
		ontErr.ErrorCode = ERROR_DNA_SUCCESS
	}
	return ontErr
}

//JsonRpcRequest object in rpc
type JsonRpcRequest struct {
	Version string        `json:"jsonrpc"`
	Id      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

//JsonRpcResponse object response for JsonRpcRequest
type JsonRpcResponse struct {
	Error  int64           `json:"error"`
	Desc   string          `json:"desc"`
	Result json.RawMessage `json:"result"`
}

func sendRpcRequest(method string, params []interface{}) ([]byte, *DNAError) {
	rpcReq := &JsonRpcRequest{
		Version: JSON_RPC_VERSION,
		Id:      "cli",
		Method:  method,
		Params:  params,
	}
	data, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, NewDNAError(fmt.Errorf("JsonRpcRequest json.Marshal error:%s", err))
	}

	addr := fmt.Sprintf("http://localhost:%d", config.DefConfig.Rpc.HttpJsonPort)
	resp, err := http.Post(addr, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return nil, NewDNAError(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, NewDNAError(fmt.Errorf("read rpc response body error:%s", err))
	}
	rpcRsp := &JsonRpcResponse{}
	err = json.Unmarshal(body, rpcRsp)
	if err != nil {
		return nil, NewDNAError(fmt.Errorf("json.Unmarshal JsonRpcResponse:%s error:%s", body, err))
	}
	if rpcRsp.Error != 0 {
		return nil, NewDNAError(fmt.Errorf("\n %s ", string(body)), rpcRsp.Error)
	}
	return rpcRsp.Result, nil
}
