package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bserdar/took/proto"
)

var forceNew bool
var forceRenew bool

func init() {
	rootCmd.AddCommand(tokenCmd)
	tokenCmd.Flags().BoolVarP(&forceNew, "force-new", "f", false, "Force new token")
	tokenCmd.Flags().BoolVarP(&forceRenew, "renew", "r", false, "Force token renewal")
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Get token",
	Long:  `Get a token, renew if necessary`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		t := viper.GetString(fmt.Sprintf("remotes.%s.type", args[0]))
		if t == "" {
			fmt.Printf("Cannot find %s\n", args[0])
			os.Exit(1)
		}
		protocol := proto.Get(t)
		if protocol == nil {
			fmt.Printf("Cannot find protocol %s\n", t)
			os.Exit(1)
		}
		cfg := protocol.GetConfigInstance()
		err := viper.UnmarshalKey(fmt.Sprintf("remotes.%s.cfg", args[0]), cfg)
		if err != nil {
			panic(err)
		}
		data := protocol.GetDataInstance()
		err = viper.UnmarshalKey(fmt.Sprintf("remotes.%s.data", args[0]), &data)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		opt := proto.UseDefault
		if forceNew {
			opt = proto.UseReAuth
		} else if forceRenew {
			opt = proto.UseRefresh
		}
		s, err := protocol.GetToken(opt)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		fmt.Println(s)
		viper.Set(fmt.Sprintf("remotes.%s.data", args[0]), data)
		e := viper.WriteConfig()
		if e != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
	}}
