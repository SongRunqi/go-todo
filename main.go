package main

import (
	"os"
	"fmt"
)

func main() {
	args := os.Args
	if (len(args) < 1) {
		fmt.Printf("args must be exits")
	}

	fmt.Printf("current arg[0] is %s", args[0])

}
