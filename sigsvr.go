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
package main

import (
	"github.com/dnaproject2/DNA/cmd"
	"github.com/dnaproject2/DNA/cmd/abi"
	cmdsvr "github.com/dnaproject2/DNA/cmd/sigsvr"
	clisvrcom "github.com/dnaproject2/DNA/cmd/sigsvr/common"
	"github.com/dnaproject2/DNA/cmd/sigsvr/store"
	"github.com/dnaproject2/DNA/cmd/utils"
	"github.com/dnaproject2/DNA/common/config"
	"github.com/dnaproject2/DNA/common/log"
	"github.com/urfave/cli"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func setupSigSvr() *cli.App {
	app := cli.NewApp()
	app.Usage = "DNA Sig server"
	app.Action = startSigSvr
	app.Version = config.Version
	app.Copyright = "Copyright in 2018 The DNA Authors"
	app.Flags = []cli.Flag{
		utils.LogLevelFlag,
		utils.CliExecutorDirFlag,
		//cli setting
		utils.CliAddressFlag,
		utils.CliRpcPortFlag,
		utils.CliABIPathFlag,
	}
	app.Commands = []cli.Command{
		cmdsvr.ImportExecutorCommand,
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}
	return app
}

func startSigSvr(ctx *cli.Context) {
	logLevel := ctx.GlobalInt(utils.GetFlagName(utils.LogLevelFlag))
	log.InitLog(logLevel, log.PATH, log.Stdout)

	executorDirPath := ctx.String(utils.GetFlagName(utils.CliExecutorDirFlag))
	if executorDirPath == "" {
		log.Errorf("Please using --executordir flag to specific executor saving path")
		return
	}

	executorStore, err := store.NewExecutorStore(executorDirPath)
	if err != nil {
		log.Errorf("NewExecutorStore error:%s", err)
		return
	}
	clisvrcom.DefExecutorStore = executorStore

	accountNum, err := executorStore.GetAccountNumber()
	if err != nil {
		log.Errorf("GetAccountNumber error:%s", err)
		return
	}
	log.Infof("Load executor data success. Account number:%d", accountNum)

	rpcAddress := ctx.String(utils.GetFlagName(utils.CliAddressFlag))
	rpcPort := ctx.Uint(utils.GetFlagName(utils.CliRpcPortFlag))
	if rpcPort == 0 {
		log.Errorf("Please using sig server port by --%s flag", utils.GetFlagName(utils.CliRpcPortFlag))
		return
	}
	go cmdsvr.DefCliRpcSvr.Start(rpcAddress, rpcPort)

	abiPath := ctx.GlobalString(utils.GetFlagName(utils.CliABIPathFlag))
	abi.DefAbiMgr.Init(abiPath)

	log.Infof("Sig server init success")
	log.Infof("Sig server listing on: %s:%d", rpcAddress, rpcPort)

	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			log.Infof("Sig server received exit signal:%v.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}

func main() {
	if err := setupSigSvr().Run(os.Args); err != nil {
		cmd.PrintErrorMsg(err.Error())
		os.Exit(1)
	}
}
