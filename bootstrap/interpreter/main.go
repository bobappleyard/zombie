package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) < 1 {
		log.Fatal("usage: interpreter file")
	}
	src, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	env := newEnv(os.Getenv("ZOMBIE_PATH"))
	pkg := Pkg{
		owner: env,
	}
	err = pkg.evalFile(src)
	if err != nil {
		log.Fatal(err)
	}
}
