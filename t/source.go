package main

import "os"

var globalstring string = "hello stderr\n"

var globalint int = 30
var globalint2 int = 0

func main() {
	print(globalstring)

	var localstring1 string
	var localstring2 string

	print("hello string literal\n")
	localstring1 = "i am a local 1\n"
	localstring2 = "i m local2\n"
	print(localstring1)
	print(localstring2)
	globalstring = "globalstring changed\n"
	print(globalstring)
	var locali3 int
	locali3 = 10
	globalint2 = 2
	os.Exit(globalint + globalint2 + locali3)
}
