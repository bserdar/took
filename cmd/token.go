package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
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
	RootCmd.AddCommand(TokenCmd)
	TokenCmd.Flags().BoolVarP(&forceNew, "force-new", "f", false, "Force new token")
	TokenCmd.Flags().BoolVarP(&forceRenew, "renew", "r", false, "Force token renewal")
	if cfg.InsecureAllowed() {
		TokenCmd.Flags().BoolVarP(&proto.InsecureTLS, "insecure", "k", false, "Insecure TLS (do not validate certificates)")
	}
	TokenCmd.Flags().BoolVarP(&writeHeader, "header", "e", false, "Write HTTP header, Authorization: Bearer <token>")
}

// TokenCmd is the took token command
var TokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Get token <config name> [username] [password]",
	Long:  `Get a token for a config, renew if necessary`,
	Args:  cobra.RangeArgs(1, 3),
	Run: func(cmd *cobra.Command, args []string) {
		cfg.DecryptUserConfig()
		userRemote, uok := cfg.UserCfg.Remotes[args[0]]
		commonRemote, cok := cfg.CommonCfg.Remotes[args[0]]
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
		protocol.SetCfg(userRemote, commonRemote)
		// userCfg := protocol.GetConfigInstance()
		// defaults := protocol.GetConfigDefaultsInstance()
		// if uok {
		// 	cfg.Decode(userRemote.Configuration, userCfg)
		// 	log.Debugf("User cfg: %v", userCfg)
		// }
		// if cok {
		// 	cfg.Decode(commonRemote.Configuration, defaults)
		// 	log.Debugf("Defaults: %v", defaults)
		// }
		// data := protocol.GetDataInstance()
		// if uok {
		// 	cfg.Decode(userRemote.Data, data)
		// }
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
		cfg.UserCfg.Remotes[args[0]] = cfg.Remote{Type: userRemote.Type, Configuration: userRemote.Configuration,
			Data: data}
		WriteUserConfig()
	}}
