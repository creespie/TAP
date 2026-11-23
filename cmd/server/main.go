package main

import (
	"flag"
	"fmt"
)

func main() {
	port := flag.Int("port", 4242, "TCP pport to listen on")
	world := flag.String("world", "world.json", "World file used to build the server's world")
	seed := flag.Int64("seed", 1, "Seed to determine luck and random rolls inside the game")

	flag.Parse()
	fmt.Printf("%d %s %d\n", *port, *world, *seed)
}
