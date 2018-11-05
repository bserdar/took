package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/crypta"
	"github.com/bserdar/took/crypta/rpc"
)

var decryptDur time.Duration

func init() {
	RootCmd.AddCommand(encryptCmd)
	RootCmd.AddCommand(decryptCmd)

	decryptCmd.Flags().DurationVarP(&decryptDur, "timeout", "t", cfg.DefaultEncTimeout, "Timeout duration. Decryption will stop after this much idle time. Default 10 minutes. 0m means never")
}

func askAndConfirmPwd() string {
top:
	pwd := cfg.AskPasswordWithPrompt("Configuration/Token encryption password: ")
	if len(pwd) == 0 {
		return ""
	}
	if cfg.AskPasswordWithPrompt("Confirm password: ") != pwd {
		fmt.Println("Passwords do not match!")
		goto top
	}
	return pwd
}

var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt the configuration file",
	Long: `Set a password to encrypt the user configuration file.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(cfg.UserCfg.AuthKey) > 0 {
			if !firstRunRan {
				fmt.Printf("Configuration file is already encrypted\n")
			}
			return
		}
		ans := cfg.Ask(`This operation will encrypt the user configuration file using a password.
There is no way to rever this operation. Do you want to continue(y/N)?`)
		if ans != "y" {
			return
		}

		pwd := askAndConfirmPwd()
		if len(pwd) == 0 {
			return
		}
		srv, err := crypta.InitServer(pwd)
		if err != nil {
			log.Fatal(err)
		}
		rp := crypta.NewRequestProcessor(srv, nil)

		cfg.UserCfg.AuthKey, err = srv.GetAuthKey()
		if err != nil {
			log.Fatal(err)
		}

		enc := func(in interface{}) string {
			in = cfg.ConvertMap(in)
			doc, err := json.Marshal(in)
			if err != nil {
				log.Fatal(err)
			}
			var rsp crypta.DataResponse
			err = rp.Encrypt(crypta.DataRequest{Data: string(doc)}, &rsp)
			if err != nil {
				log.Fatal(err)
			}
			return rsp.Data
		}

		out := make(map[string]cfg.Remote)
		for k, v := range cfg.UserCfg.Remotes {
			if v.Configuration != nil {
				v.ECfg = enc(v.Configuration)
				v.Configuration = nil
			}
			if v.Data != nil {
				v.EData = enc(v.Data)
				v.Data = nil
			}
			out[k] = v
		}
		cfg.UserCfg.Remotes = out
		WriteUserConfig()
	}}

var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt the configuration file",
	Long: `Enter the password to decrypt the configuration file.
`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(cfg.UserCfg.AuthKey) == 0 {
			fmt.Printf("Configuration file is not encrypted\n")
			return
		}
		if len(args) == 1 && args[0] == "x" {
			reader := bufio.NewReader(os.Stdin)
			s, e := reader.ReadString('\n')
			if e != nil {
				log.Fatal(e)
			}
			if s[len(s)-1] == '\n' {
				s = s[:len(s)-1]
			}
			socketName, e := homedir.Expand("~/.took.s")
			if e != nil {
				log.Fatal(e)
			}
			os.Remove(socketName)
			e = rpc.Server(socketName, s, cfg.UserCfg.AuthKey, decryptDur)
			if e != nil {
				log.Fatal(e)
			}
		} else {
			cfg.AskPasswordStartDecrypt(decryptDur)
		}
	}}
