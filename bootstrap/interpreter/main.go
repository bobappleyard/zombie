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
	registerBuiltins(env)
	pkg := Pkg{
		owner: env,
		path:  "<main>",
		defs:  map[string]any{},
	}
	err = pkg.Import("zombie")
	if err != nil {
		log.Fatal(err)
	}
	err = pkg.evalFile(src)
	if err != nil {
		log.Fatal(err)
	}
}
