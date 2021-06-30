package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

// exit code on interrupt
const exitCodeInterrupt = 2

// default global timeout value
var timeout = flag.Uint("timeout", 120, "timeout in seconds")

func main() {
	flag.Parse()
	// handle ^C interrupt signals gracefully
	ctx, cancel := context.WithDeadline(context.Background(),
		time.Now().Add(time.Duration(*timeout)*time.Second))
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	go func() {
		select {
		case <-signalChan:
			log.Println("exiting. ^C again to force.")
			cancel()
		case <-ctx.Done():
		}
		<-signalChan
		os.Exit(exitCodeInterrupt)
	}()
	// get Lagoon project/environment build variables from process environment
	buildVars, err := LagoonBuildVars()
	if err != nil {
		log.Fatalf("couldn't get Lagoon build variables: %v", err)
	}
	// get variables from any configured secret stores
	merged, err := MergeSecrets([]SecretStore{
		&MockBackend{ctx: ctx},
		&AWSSecretsManager{ctx: ctx},
		&GoogleSecretManager{ctx: ctx},
	}, buildVars)
	if err != nil {
		log.Fatalf("couldn't merge secrets: %v", err)
	}
	// print JSON-formatted variables to stdout
	output, err := json.Marshal(merged)
	if err != nil {
		log.Fatalf("couldn't marshal JSON: %v", err)
	}
	fmt.Println(string(output))
}
