package main

import (
	"fmt"
	log "seelog"
)

func log_open(log_file string) {
	fmt.Println("Date time format")

	//logger, err := log.LoggerFromConfigAsBytes([]byte(testConfig))
	logger, err := log.LoggerFromConfigAsFile("tls_benchmark.xml")

	if err != nil {
		fmt.Println(err)
	}

	loggerErr := log.ReplaceLogger(logger)

	if loggerErr != nil {
		fmt.Println(loggerErr)
	}

	/* Usage: */
	/*	log.Trace("Test message!")
		log.Info("Hello from Seelog!")
		log.Error("Error msg!")
		log.Warn("Warn msg!")
		log.Critical("Critical msg!") */
}
