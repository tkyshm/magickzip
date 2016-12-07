package main

import (
	"flag"
	"fmt"

	"github.com/tkyshm/magickzip"
	"gopkg.in/gographics/imagick.v2/imagick"
)

var (
	conf string
)

func main() {
	flag.StringVar(&conf, "conf", "~/.config/magickzip/conf.yml", "conf file")
	flag.Parse()

	config, err := magickzip.LoadConfig(conf)
	if err != nil {
		panic(err)
	}

	fmt.Println(config)
	imagick.Initialize()
	defer imagick.Terminate()
}
