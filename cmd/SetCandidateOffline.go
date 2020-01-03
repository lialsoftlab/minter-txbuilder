/*
Copyright © 2020 Aleksey V. Litvinov

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/MinterTeam/minter-go-sdk/api"
	"github.com/MinterTeam/minter-go-sdk/transaction"
	"github.com/MinterTeam/minter-go-sdk/wallet"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

// SetCandidateOfflineCmd represents the SetCandidateOffline command
var SetCandidateOfflineCmd = &cobra.Command{
	Use:   "SetCandidateOffline",
	Short: "Build a transaction for setting candidate into offline mode",
	//	Long: `A longer description that spans multiple lines and likely contains examples
	//and usage of using your command. For example:
	//
	//Cobra is a CLI library for Go that empowers applications.
	//This application is a tool to generate the needed files
	//to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("SetCandidateOffline called")
		//for _, url := range viper.GetStringSlice("common.api_nodes") {
		//	fmt.Printf("\t%s\n", url)
		//}
		//fmt.Println(viper.GetString("tx_set_candidate_offline.validator_pubkey"))

		fmt.Print("Seed-phrase of owner wallet (hidden input): ")
		var mnemonic []byte
		var err error
		mnemonic, err = terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println("ERROR (", err, ")")
			fmt.Print("Seed-phrase of owner wallet: ")
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			text = strings.Replace(text, "\n", "", -1)
			text = strings.Replace(text, "\r", "", -1)
			mnemonic = []byte(text)
		}

		//mnemonic = []byte("aaaaa bbbbbb cccccc dddddd eeeeee ffffff hhhh iiiii hhhhhh jjj kkkkkk lllll")

		seed, _ := wallet.Seed(string(mnemonic))

		wlt, err := wallet.NewWallet(seed)
		exitIfItIsError(err)

		nonce, err := getNonce(viper.GetStringSlice("common.api_nodes"), wlt.Address())
		exitIfItIsError(err)

		validator_pubkey := viper.GetString("tx_set_candidate_offline.validator_pubkey")

		offData := transaction.NewSetCandidateOffData().MustSetPubKey(validator_pubkey)
		txOff, err := transaction.NewBuilder(getNetworkChain()).NewTransaction(offData)
		exitIfItIsError(err)

		txOff.SetNonce(nonce).SetGasCoin(getCoin()).SetGasPrice(50)
		signedTxOff, err := txOff.Sign(wlt.PrivateKey())
		exitIfItIsError(err)

		tx_enc, _ := signedTxOff.Encode()

		fmt.Println("Validator pubkey:", validator_pubkey)
		fmt.Println("Owner pubkey:", wlt.PublicKey())
		fmt.Println("TX OFF:", tx_enc)

		f, err := os.Create(fmt.Sprintf("tx_off_%s.yaml", validator_pubkey))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer exitIfItIsError(f.Close())

		_, err = f.WriteString(fmt.Sprintf(
			"validator_pubkey: %s\nowner_pubkey: %s\ntx_off: %s\n",
			validator_pubkey, wlt.PublicKey(), tx_enc))
		exitIfItIsError(err)
	},
}

func init() {
	rootCmd.AddCommand(SetCandidateOfflineCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// SetCandidateOfflineCmd.PersistentFlags().String("foo", "", "A help for foo")

	//Cobra supports local flags which will only run when this command
	//is called directly, e.g.:
	SetCandidateOfflineCmd.Flags().StringP("validator-pubkey", "k", "", "public key of validator node")
	_ = viper.BindPFlag("tx_set_candidate_offline.validator_pubkey", SetCandidateOfflineCmd.Flags().Lookup("validator-pubkey"))
}

func getNonce(api_nodes []string, address string) (nonce uint64, err error) {
	for _, api_node := range api_nodes {
		nonce, err = api.NewApi(api_node).Nonce(address)
		if err == nil {
			return
		}
		fmt.Println(err)
	}

	return 0, errors.New(fmt.Sprintf("Can't get nonce for %s", address))
}

func exitIfItIsError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(10)
	}
}

func getNetworkChain() transaction.ChainID {
	if viper.GetBool("common.use_testnet") {
		return transaction.TestNetChainID
	} else {
		return transaction.MainNetChainID
	}
}

func getCoin() string {
	if viper.GetBool("common.use_testnet") {
		return "MNT"
	} else {
		return "BIP"
	}
}
