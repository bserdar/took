package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"

	"github.com/bserdar/took/cfg"
	"github.com/bserdar/took/proto"
)

var forceNew bool
var forceRenew bool
var writeHeader bool
var userName string

var insecureTLS bool

func init() {
	rootCmd.AddCommand(tokenCmd)
	tokenCmd.Flags().BoolVarP(&forceNew, "force-new", "f", false, "Force new token")
	tokenCmd.Flags().BoolVarP(&forceRenew, "renew", "r", false, "Force token renewal")
	tokenCmd.Flags().BoolVarP(&proto.InsecureTLS, "insecure", "k", false, "Insecure TLS (do not validate certificates)")
	tokenCmd.Flags().BoolVarP(&writeHeader, "header", "e", false, "Write HTTP header, Authorization: Bearer <token>")
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Get token <config name> [username] [password]",
	Long:  `Get a token for a config, renew if necessary`,
	Args:  cobra.RangeArgs(1, 3),
	Run: func(cmd *cobra.Command, args []string) {
		userRemote, uok := UserCfg.Remotes[args[0]]
		commonRemote, cok := CommonCfg.Remotes[args[0]]
		if !uok && !cok {
			log.Fatalf("Cannot find %s\n", args[0])
		}
		t := userRemote.Type
		if len(t) == 0 {
			t = commonRemote.Type
		}
		if len(t) == 0 {
			log.Fatalf("Invalid configuration: no type for %s", args[0])
		}
		protocol := proto.Get(t)
		if protocol == nil {
			fmt.Printf("Cannot find protocol %s\n", t)
			os.Exit(1)
		}
		userCfg := protocol.GetConfigInstance()
		defaults := protocol.GetConfigDefaultsInstance()
		if uok {
			err := mapstructure.Decode(userRemote.Configuration, userCfg)
			if err != nil {
				log.Fatalf("Error reading configuration: %s", err)
			}
		}
		if cok {
			err := mapstructure.Decode(commonRemote.Configuration, defaults)
			if err != nil {
				log.Fatalf("Error reading common configuration: %s", err)
			}
		}
		data := protocol.GetDataInstance()
		if uok {
			err := mapstructure.Decode(userRemote.Data, data)
			if err != nil {
				log.Fatalf("Error reading data: %s", err)
			}
		}
		opt := proto.UseDefault
		if forceNew {
			opt = proto.UseReAuth
		} else if forceRenew {
			opt = proto.UseRefresh
		}
		out := proto.OutputToken
		if writeHeader {
			out = proto.OutputHeader
		}
		userName := ""
		password := ""
		if len(args) > 1 {
			userName = args[1]
		}
		if len(args) > 2 {
			password = args[2]
		}
		s, data, err := protocol.GetToken(proto.TokenRequest{Refresh: opt, Out: out, Username: userName, Password: password})
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		fmt.Println(s)
		UserCfg.Remotes[args[0]] = cfg.Remote{Type: userRemote.Type, Configuration: userRemote.Configuration,
			Data: data}
		writeUserConfig()
	}}
