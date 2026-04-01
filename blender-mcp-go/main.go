package main

import "os"

func main() {
	cfg := LoadConfig()
	serve(os.Stdin, os.Stdout, cfg)
}
