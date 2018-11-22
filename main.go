package main

import (
	"fmt"
	"os"
	"os/signal"
)

func main() {
	stress := New()
	stress.ConfigWithArgs()
	stress.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	recover := make(chan string, 1)

	for {
		select {
		case <-interrupt:
			fmt.Print("Pausing, ")
			if stress.isPause {
				stress.Stop()
			} else {
				stress.Pause()
				go func() {
					fmt.Println("press Ctrl+C exit or Enter continue...")
					var input string
					fmt.Scanln(&input)
					recover <- input
				}()
			}
		case <-recover:
			stress.Recover()
		case <-stress.exit:
			return
		}
	}
}
