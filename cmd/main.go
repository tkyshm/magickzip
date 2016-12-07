package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/tkyshm/magickzip"
	//"github.com/tkyshm/magickzip/magick"
	"gopkg.in/gographics/imagick.v2/imagick"
)

var (
	conf     string
	inputDir string
	config   *magickzip.Config
)

func main() {
	flag.StringVar(&conf, "conf", "~/.config/magickzip/conf.yml", "conf file")
	flag.StringVar(&inputDir, "input-dir", "./input", "input directory has images")
	flag.Parse()

	var err error
	config, err = magickzip.LoadConfig(conf)
	if err != nil {
		panic(err)
	}

	// 1. make base structure
	makeStructure(config.Structure, ".")

	// 2. modulating files
	for dir, m := range config.Modulate {
		// TODO: dirの中の画像をmの値でmodulate
		log.Printf("directory:%s, modulate:%d%%", dir, m)
	}

	// 3. resize files
	startResize(config.Resize)
	for dir, params := range config.Resize {
		log.Printf("directory:%s, params:%v", dir, params)
	}

	imagick.Initialize()
	defer imagick.Terminate()
}

func makeStructure(list map[interface{}]interface{}, parent string) {
	for dir, content := range list {
		// chain path
		path := fmt.Sprintf("%s/%s", parent, dir.(string))
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			log.Println("make directory:", dir)
			err := os.Mkdir(path, 0755)
			if err != nil {
				panic(err)
			}
		}

		switch t := content.(type) {
		case []interface{}: // files
			for _, r := range content.([]interface{}) {
				rpath := fmt.Sprintf("%s/%s", inputDir, r)
				log.Println(rpath)
				files, err := filepath.Glob(rpath)
				if err != nil {
					continue
				}

				// file copy
				for _, srcFile := range files {
					// Source
					src, err := os.Open(srcFile)
					if err != nil {
						log.Println("[error] ", err)
						continue
					}

					// Destiation
					dstFile := fmt.Sprintf("%s/%s", path, filepath.Base(srcFile))
					dst, err := os.Create(dstFile)
					if err != nil {
						log.Println("[error] ", err)
						continue
					}

					// Copy file
					_, err = io.Copy(dst, src)
					if err != nil {
						log.Println("[error] ", err)
						continue
					}

					src.Close()
					dst.Close()
				}
			}
		case map[interface{}]interface{}: // directory
			makeStructure(content.(map[interface{}]interface{}), path)
		default: // unexpected type
			log.Printf("[error] unexpected type: %T", t)
			log.Println("[error] content: ", content)
		}
	}
}

func startResize(list map[interface{}]interface{}) {
	for dir, params := range list {
		for subdir, size := range params.(map[interface{}]interface{}) {
			// check resize params struct
			switch t := size.(type) {
			case map[interface{}]interface{}:
				files, err := filepath.Glob("./" + config.Root + "/*" + dir.(string))
				if err != nil {
					log.Println("[error] ", err)
					continue
				}
				// all paths
				for _, path := range files {
					dstPath := fmt.Sprintf("%s/%s", path, subdir)

					// mkdir if not exist directory
					if _, err := os.Stat(dstPath); os.IsNotExist(err) {
						err := os.Mkdir(dstPath, 0755)
						if err != nil {
							log.Println("[error] ", err)
							continue
						}
					}

					// find image files
					imgs, err := filepath.Glob(path + "/*")
					if err != nil {
						log.Println("[error] ", err)
						continue
					}

					// resize
					for _, img := range imgs {
						// TODO: resize
						//h := size.(map[interface{}]interface{})[string("height")]
						//w := size.(map[interface{}]interface{})[string("width")]
						// only file
						fi, _ := os.Stat(img)
						// skip directory
						if fi.IsDir() {
							continue
						}
						// dst: path+subdir, src: path/**
						fmt.Println(img)
					}
				}

			default:
				log.Printf("[error] unexpected type: %T", t)
				log.Println("[error] content: ", size)
			}
		}
		log.Printf("directory:%s, params:%+v", dir, params)
	}
}
