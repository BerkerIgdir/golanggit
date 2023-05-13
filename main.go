package main

import "golang/gitproject/fileop"

const (
	berkersDir = "/Users/berker.igdir/check24"
)

// TO DO: Add a channel for successfull and failed builds respectively
func main() {
	fileop.Run(berkersDir)
}
