/*
Copyright © 2022
<Bhakiyaraj Kalimuthu>
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{
		Use:   "notifier",
		Short: "A message notifier",
		Long:  `A message notifier`,
		Run:   runRootCmd,
	}
	rootArgs struct {
		url      string
		interval int
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	doneCh := make(chan os.Signal, 1)
	signal.Notify(doneCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-doneCh
		signal.Stop(doneCh)
		fmt.Println("CTRL-C received.Terminating...")
		os.Exit(0)
	}()
	root := rootCmd.Flags()
	root.StringVar(&rootArgs.url, "url", "localhost:8080", "URL to which notification to be sent")
	root.IntVar(&rootArgs.interval, "interval", 5, "Notification interval in seconds")
}

func runRootCmd(cmd *cobra.Command, args []string) {
	//	scanner := bufio.NewScanner(os.Stdin)
	//	var text string
	//	for text != "q" {
	//		fmt.Println("Enter your text:")
	//		scanner.Scan()
	//		text = scanner.Text()
	//		fmt.Println(text)
	//	}
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Errorf("failed to read %v", err)
	}
	fmt.Println(string(bytes))
	// send data
	// send interval
	// send url
}
