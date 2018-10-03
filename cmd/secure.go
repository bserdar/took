package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/crypta"
	"github.com/bserdar/took/crypta/rpc"
	"github.com/bserdar/took/proto"
)

func init() {
	RootCmd.AddCommand(encryptCmd)
	RootCmd.AddCommand(decryptCmd)
}

var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt the configuration file",
	Long: `Set a password to encrypt the user configuration file.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(cfg.UserCfg.AuthKey) > 0 {
			fmt.Printf("Configuration file is already encrypted\n")
			return
		}
		ans := proto.Ask(`This operation will encrypt the user configuration file using a password.
There is no way to rever this operation. Do you want to continue(y/N)?`)
		if ans != "y" {
			return
		}
		pwd := proto.AskPasswordWithPrompt("Password: ")
		if len(pwd) == 0 {
			return
		}
		if proto.AskPasswordWithPrompt("Confirm password: ") != pwd {
			panic("Passwords do not match")
		}

		srv, err := crypta.InitServer(pwd)
		if err != nil {
			panic(err)
		}
		rp := crypta.NewRequestProcessor(srv, nil)

		cfg.UserCfg.AuthKey, err = srv.GetAuthKey()
		if err != nil {
			panic(err)
		}

		enc := func(in interface{}) string {
			in = cfg.ConvertMap(in)
			doc, err := json.Marshal(in)
			if err != nil {
				panic(err)
			}
			var rsp crypta.DataResponse
			err = rp.Encrypt(crypta.DataRequest{Data: string(doc)}, &rsp)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Enc: In: %s out: %s\n", doc, rsp.Data)
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
				panic(e)
			}
			if s[len(s)-1] == '\n' {
				s = s[:len(s)-1]
			}
			socketName, e := homedir.Expand("~/.took.s")
			if e != nil {
				panic(e)
			}
			os.Remove(socketName)
			e = rpc.RPCServer(socketName, s, cfg.UserCfg.AuthKey, 10*time.Minute)
			if e != nil {
				panic(e)
			}
		} else {
			pwd := proto.AskPasswordWithPrompt("Password: ")
			_, err := crypta.NewServer(pwd, cfg.UserCfg.AuthKey)
			if err != nil {
				panic(err)
			}
			cmd := exec.Command(os.Args[0], "decrypt", "x")
			wr, _ := cmd.StdinPipe()
			cmd.Start()
			wr.Write([]byte(pwd))
			wr.Write([]byte("\n"))
		}
	}}
