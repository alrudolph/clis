package main

import (
	"github.com/alrudolph/clis/src/sync-static-site-s3/cmd"
	_ "github.com/alrudolph/clis/src/sync-static-site-s3/cmd/config"
	_ "github.com/alrudolph/clis/src/sync-static-site-s3/cmd/setup"
)

func main() {
	cmd.Execute()
}
