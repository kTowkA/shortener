package main

import "os"

func main() {
	os.Exit(1) // want "not recommended function"
	otherFunc()
}
func otherFunc() {
	os.Exit(2)
}
