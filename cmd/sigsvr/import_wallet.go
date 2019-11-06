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
package sigsvr

import (
	"fmt"
	"github.com/dnaproject2/DNA/account"
	"github.com/dnaproject2/DNA/cmd"
	"github.com/dnaproject2/DNA/cmd/sigsvr/store"
	"github.com/dnaproject2/DNA/cmd/utils"
	"github.com/dnaproject2/DNA/common"
	"github.com/urfave/cli"
)

var ImportExecutorCommand = cli.Command{
	Name:      "import",
	Usage:     "Import accounts from a executor file",
	ArgsUsage: "",
	Action:    importExecutor,
	Flags: []cli.Flag{
		utils.CliExecutorDirFlag,
		utils.ExecutorFileFlag,
	},
	Description: "",
}

func importExecutor(ctx *cli.Context) error {
	executorDirPath := ctx.String(utils.GetFlagName(utils.CliExecutorDirFlag))
	executorFilePath := ctx.String(utils.GetFlagName(utils.ExecutorFileFlag))
	if executorDirPath == "" || executorFilePath == "" {
		cmd.PrintErrorMsg("Missing %s or %s flag.", utils.CliExecutorDirFlag.Name, utils.ExecutorFileFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	if !common.FileExisted(executorFilePath) {
		return fmt.Errorf("executor file:%s does not exist", executorFilePath)
	}
	executorStore, err := store.NewExecutorStore(executorDirPath)
	if err != nil {
		return fmt.Errorf("NewExecutorStore dir path:%s error:%s", executorDirPath, err)
	}
	executor, err := account.Open(executorFilePath)
	if err != nil {
		return fmt.Errorf("open executor:%s error:%s", executorFilePath, err)
	}
	executorData := executor.GetExecutorData()
	if *executorStore.ExecutorScrypt != *executorData.Scrypt {
		return fmt.Errorf("import account failed, executor scrypt:%+v != %+v", executorData.Scrypt, executorStore.ExecutorScrypt)
	}
	addNum := 0
	updateNum := 0
	for i := 0; i < len(executorData.Accounts); i++ {
		ok, err := executorStore.AddAccountData(executorData.Accounts[i])
		if err != nil {
			return fmt.Errorf("import account address:%s error:%s", executorData.Accounts[i].Address, err)
		}
		if ok {
			addNum++
		} else {
			updateNum++
		}
	}
	cmd.PrintInfoMsg("Import account success.")
	cmd.PrintInfoMsg("Total account number:%d", len(executorData.Accounts))
	cmd.PrintInfoMsg("Add account number:%d", addNum)
	cmd.PrintInfoMsg("Update account number:%d", updateNum)
	return nil
}
