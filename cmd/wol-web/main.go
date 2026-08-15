// WOL Web - 远程开机 (Wake-on-LAN) Web 服务入口。
package main

import (
	"log"

	"github.com/632-8nm/remote-wakeup/internal/config"
	"github.com/632-8nm/remote-wakeup/internal/web"
)

func main() {
	// Optionally load a dotenv file (secrets stay out of the repo).
	config.LoadEnv(".env")

	cfg := config.Default()
	cfg.Resolve()

	srv := web.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
