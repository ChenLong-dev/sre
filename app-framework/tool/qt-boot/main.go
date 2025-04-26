package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"gitlab.shanhai.int/sre/app-framework/tool/qt-boot/internal/project"
)

func main() {
	app := &cli.App{
		// 注册子命令
		Commands: []*cli.Command{
			project.GenProjectCmd,
		},
		Usage: `qt-boot a toolbox for app-framework`,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}

}
