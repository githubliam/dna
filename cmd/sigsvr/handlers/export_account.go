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
	"fmt"
	"github.com/dnaproject2/DNA/account"
	clisvrcom "github.com/dnaproject2/DNA/cmd/sigsvr/common"
	"github.com/dnaproject2/DNA/common"
	"github.com/dnaproject2/DNA/common/log"
	"io/ioutil"
	"strings"
	"time"
)

type ExportAccountReq struct {
	ExecutorPath string `json:"executor_path"`
}

type ExportAccountResp struct {
	ExecutorFile  string `json:"executor_file"`
	AccountNumber int    `json:"account_num"`
}

func ExportAccount(req *clisvrcom.CliRpcRequest, resp *clisvrcom.CliRpcResponse) {
	expReq := &ExportAccountReq{}
	err := json.Unmarshal(req.Params, expReq)
	if err != nil {
		resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
		log.Infof("ExportAccount Qid:%s json.Unmarshal ExportAccountReq error:%s", req.Qid, err)
		return
	}
	executorPath := expReq.ExecutorPath
	if executorPath != "" {
		if !common.FileExisted(executorPath) {
			resp.ErrorCode = clisvrcom.CLIERR_INVALID_PARAMS
			resp.ErrorInfo = "executor path doesn't exist"
			return
		}
	} else {
		executorPath = "./"
	}

	executorStore := clisvrcom.DefExecutorStore
	executorData := &account.ExecutorData{
		Name:     executorStore.ExecutorName,
		Version:  executorStore.ExecutorVersion,
		Scrypt:   executorStore.ExecutorScrypt,
		Accounts: make([]*account.AccountData, 0),
		Extra:    executorStore.ExecutorExtra,
	}

	accountCount := executorStore.GetNextAccountIndex()
	for i := uint32(0); i < accountCount; i++ {
		accData, err := executorStore.GetAccountDataByIndex(i)
		if err != nil {
			log.Errorf("ExportAccount Qid:%s GetAccountDataByIndex:%d error:%s\n", req.Qid, i, err)
			continue
		}
		if accData == nil {
			continue
		}
		executorData.Accounts = append(executorData.Accounts, accData)
	}

	data, err := json.Marshal(executorData)
	if err != nil {
		log.Errorf("ExportAccount Qid:%s json.Marshal ExecutorData error:%s\n", req.Qid, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}

	executorFile := fmt.Sprintf("%s/executor_%s.dat", strings.TrimRight(executorPath, "/"), time.Now().Format("2006_01_02_15_04_05"))
	err = ioutil.WriteFile(executorFile, data, 0666)
	if err != nil {
		log.Errorf("ExportAccount Qid:%s write file:%s error:%s", req.Qid, executorFile, err)
		resp.ErrorCode = clisvrcom.CLIERR_INTERNAL_ERR
		return
	}

	resp.Result = &ExportAccountResp{
		ExecutorFile:  executorFile,
		AccountNumber: len(executorData.Accounts),
	}
	log.Infof("ExportAccount Qid:%s success executor file:%s", req.Qid, executorFile)
}
