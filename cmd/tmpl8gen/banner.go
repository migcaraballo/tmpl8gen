package main

import (
	"fmt"
)

func printBanner() {
	fmt.Println(banner())
	fmt.Println("ver: 1.0.0")
	fmt.Println("enter 'q' to quit")
	fmt.Println()
}

func banner() string {
	return `
 _____           _ ___    ___          
|_   _| __  _ __| ( _ )  / __|___ _ _  
  | || '  \| '_ \ / _ \ | (_ / -_) ' \ 
  |_||_|_|_| .__/_\___/  \___\___|_||_|
           |_|                         
`
}
