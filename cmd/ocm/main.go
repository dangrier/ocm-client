package main

import (
	_ "time/tzdata" // embed IANA timezone data so Australia/Brisbane works anywhere

	"github.com/dangrier/ocm-client/cli"
)

func main() {
	cli.Execute()
}
