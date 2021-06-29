package main

import (
	"os"

	"github.com/line/lfb-sdk/server"
	svrcmd "github.com/line/lfb-sdk/server/cmd"

	"github.com/line/lfb/app"
	"github.com/line/lfb/cmd/lfb/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)

		default:
			os.Exit(1)
		}
	}
}
