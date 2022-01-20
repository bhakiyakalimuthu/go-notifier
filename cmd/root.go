/*
Copyright © 2022
<Bhakiyaraj Kalimuthu>
*/
package cmd

import (
	"bufio"
	"context"
	"go-notifier/internal"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const workerPoolSize = 5

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{
		Use:   "notifier",
		Short: "message notifier",
		Long:  `message notifier can notify the message to the configured URL`,
		Run:   runRootCmd,
	}
	rootArgs struct {
		url      string
		interval time.Duration
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
	root := rootCmd.Flags()
	root.StringVarP(&rootArgs.url, "url", "u", "localhost:8080", "URL to which notification to be sent")
	root.DurationVarP(&rootArgs.interval, "interval", "i", 1*time.Second, "Notification interval")
}

func runRootCmd(cmd *cobra.Command, args []string) {
	// logger setup
	l := loggerSetup()
	// producer channel
	pChan := make(chan string, 1)
	// consumerChannel
	cChan := make(chan string, workerPoolSize)

	// create notifier
	notifier := internal.NewNotifier(l, rootArgs.url, rootArgs.interval, pChan, cChan)

	// setup cancellation context and wait group
	ctx, cancel := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)

	// start notifier and pass the cancellation ctx
	go notifier.Start(ctx)

	// start workers and add worker pool
	wg.Add(workerPoolSize)
	for i := 0; i < workerPoolSize; i++ {
		go notifier.Process(wg, i)
	}

	doneCh := make(chan os.Signal, 1)
	signal.Notify(doneCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-doneCh // block here
		signal.Stop(doneCh)
		l.Warn("CTRL-C received.Terminating......")

		// handle shut down
		cancel()   // cancel context
		wg.Wait()  // wait for the workers to be completed
		os.Exit(0) // Exit
	}()

	scanner := bufio.NewScanner(os.Stdin)
	var text string
	for scanner.Scan() {
		text = scanner.Text()
		pChan <- text
		l.Debug(text)
	}
	// bufio.Scanner has max buffer size 64*1024 bytes which means
	// in case file has any line greater than the size of 64*1024,
	// then it will throw error
	if err := scanner.Err(); err != nil {

		l.Fatal("line length exceeded the bufio scanner max buffer size of 64*1024", zap.Error(err))
	}
}
func loggerSetup() *zap.Logger {
	//if cfg.Env == config.EnvProd {
	//	logger, err := zap.NewProduction()
	//	if err != nil {
	//		log.Fatalf("failed to create zap logger : %v", err)
	//	}
	//	return logger
	//}

	// setup dev logger logger to show different colors
	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	log := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zapcore.AddSync(colorable.NewColorableStdout()),
		zapcore.InfoLevel,
	))
	log.Info("logger setup done")
	return log
}
