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
package handlers

import (
	"encoding/json"
	clisvrcom "github.com/dnaproject2/DNA/cmd/sigsvr/common"
	"github.com/dnaproject2/DNA/common/log"
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
)

type CreateAccountReq struct {
}

type CreateAccountRsp struct {
	Account string `json:"account"`
}

func CreateAccount(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	pwd := req.Pwd
	if pwd == "" {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		resp.ErrorInfo = "pwd cannot empty"
		return
	}
	accData, err := clisvrcom.DefExecutorStore.NewAccountData(keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, []byte(pwd))
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		resp.ErrorInfo = "create executor failed"
		log.Errorf("CreateAccount Qid:%s NewAccountData error:%s", req.Qid, err)
		return
	}
	_, err = clisvrcom.DefExecutorStore.AddAccountData(accData)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		resp.ErrorInfo = "create executor failed"
		log.Errorf("CreateAccount Qid:%s AddAccountData error:%s", req.Qid, err)
		return
	}
	resp.Result = &CreateAccountRsp{
		Account: accData.Address,
	}

	data, _ := json.Marshal(accData)
	log.Infof("[CreateAccount]%s", data)
}
