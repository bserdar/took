package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/crypta"
)

var firstRunRan bool

func firstRun() {
	firstRunRan = true
	fmt.Println(`A new took.yaml file is created to store your API credentials and tokens.
You have the option to keep these credentials and tokens as plain text, however password
encryption will prevent accidental disclosure.`)
again:
	ans := cfg.Ask("Do you want to encrypt the configuration? (Y/n) ")
	if ans == "y" || ans == "Y" || ans == "" {
		pwd := askAndConfirmPwd()
		if len(pwd) == 0 {
			return
		}
		srv, err := crypta.InitServer(pwd)
		if err != nil {
			log.Fatal(err)
		}
		cfg.UserCfg.AuthKey, err = srv.GetAuthKey()
		if err != nil {
			log.Fatal(err)
		}
		WriteUserConfig()
		cfg.StartDecrypt(pwd, 10*time.Minute)

	} else if ans == "n" || ans == "N" {
	} else {
		goto again
	}
}
