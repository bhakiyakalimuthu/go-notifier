/*
Copyright © 2022
Author Bhakiyaraj Kalimuthu
Email bhakiya.kalimuthu@gmail.com
*/
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"go-notifier/internal"
	"log"
	"net/url"
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

const (
	workerPoolSize = 5     // default worker pool size
	env            = "dev" // development or production env
)

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{
		Use:   "notifier",
		Short: "message notifier",
		Long:  `message notifier can notify the message to the configured URL`,
		Run:   runRootCmd,
	}
	rootArgs struct {
		url      string        // url where notification to be sent
		interval time.Duration // interval in which notification to be sent
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
	root.StringVarP(&rootArgs.url, "url", "u", "", "URL to which notification to be sent")
	root.DurationVarP(&rootArgs.interval, "interval", "i", 100*time.Millisecond, "Notification interval")
	cobra.MarkFlagRequired(root, "url")
}

func runRootCmd(cmd *cobra.Command, args []string) {
	// logger setup
	l := loggerSetup()

	// init clock
	clock := internal.NewClock()
	defer func() {
		l.Info("Time taken to complete", zap.Duration("time_taken", <-clock.Since()))
	}()

	// validate url and fail early
	if !isValidURL(rootArgs.url) {
		cmd.Help()
		os.Exit(1)
	}

	// producer channel
	pChan := make(chan string, 1)
	// consumer channel
	cChan := make(chan string, workerPoolSize)

	// create http client
	httpClient := internal.NewHttpClient(l, rootArgs.url)

	// create notifier
	notifier := internal.NewNotifier(l, httpClient, rootArgs.interval, pChan, cChan)

	// setup cancellation context and wait group
	// root background with cancellation support
	ctx, cancel := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)

	// start notifier and pass the cancellation ctx
	go notifier.Start(ctx)

	// start workers and add worker pool
	wg.Add(workerPoolSize)
	for i := 1; i <= workerPoolSize; i++ {
		go notifier.Process(wg, i)
	}

	doneCh := make(chan os.Signal, 1)

	// user input
	go func(cancelFunc context.CancelFunc, doneCh chan os.Signal) {
		// new buffer io scanner to get user input
		scanner := bufio.NewScanner(os.Stdin)
		var msg string
		for scanner.Scan() {
			msg = scanner.Text()
			pChan <- msg // send in data to producer channel
			l.Debug(msg)
		}
		// bufio.Scanner has max buffer size 64*1024 bytes which means
		// in case file has any line greater than the size of 64*1024,
		// then it will throw error
		// Note: buffer limit can be increased by using scanner.Buffer
		// just for the simplicity bufio.Scanner default is used
		if err := scanner.Err(); err != nil {

			l.Fatal("line length exceeded the bufio scanner max buffer size of 64*1024", zap.Error(err))
			os.Exit(1)
		}

		<-time.Tick(time.Second * 1) // wait for all the workers to finish up
		// once file read is completed, send quit system call
		// exit the program
		doneCh <- syscall.SIGQUIT
	}(cancel, doneCh)

	// handle manual interruption
	signal.Notify(doneCh, syscall.SIGINT, syscall.SIGTERM)

	switch <-doneCh { // blocks here until interrupted
	case syscall.SIGINT, syscall.SIGTERM:
		l.Warn("CTRL-C received.Terminating......")
	default:
		l.Warn("file read is completed,exiting......")
	}
	signal.Stop(doneCh)

	// handle shut down
	cancel() // cancel context
	// even if cancellation received, current running job will be not be interrupted until it completes
	wg.Wait() // wait for the workers to be completed
	l.Warn("All jobs are done, shutting down")

}

// loggerSetup setup zap logger
func loggerSetup() *zap.Logger {
	if env == "prod" {
		logger, err := zap.NewProduction()
		if err != nil {
			log.Fatalf("failed to create zap logger : %v", err)
		}
		logger.Info("logger setup done")
		return logger
	}

	// setup dev logger to show different colors
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

func isValidURL(URL string) bool {
	if rootArgs.url == "" {
		fmt.Println("Error: url field is empty")
		return false
	}

	// parse url if valid
	_, err := url.ParseRequestURI(URL)
	if err != nil {
		fmt.Printf("Error: invalid url %v", err)
		return false
	}
	return true
}
