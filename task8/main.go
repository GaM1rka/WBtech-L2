package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/beevik/ntp"
)

func main() {
	writer := bufio.NewWriter(os.Stderr)

	t, err := ntp.Time("0.beevik-ntp.pool.ntp.org")
	if err != nil {
		writer.Write([]byte(err.Error()))
		writer.Flush()
		os.Exit(1)
	}

	fmt.Println(t)
}
