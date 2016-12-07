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
	template string
	inputDir string
	config   *magickzip.Config
)

func main() {
	flag.StringVar(&template, "template", "~/.config/magickzip/template.yml", "template file")
	flag.StringVar(&inputDir, "input-dir", "./input", "input directory has images")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var err error
	config, err = magickzip.LoadConfig(template)
	if err != nil {
		panic(err)
	}

	imagick.Initialize()
	defer imagick.Terminate()

	// 1. make base structure
	makeStructure(config.Structure, ".")
}

func makeStructure(list map[interface{}]interface{}, parent string) {
	for dir, content := range list {
		// chain path
		path := fmt.Sprintf("%s/%s", parent, dir)
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
					basename := filepath.Base(srcFile)
					dstFile := fmt.Sprintf("%s/%s", path, basename)

					// TODO: check resize
					if isResizeDir(dir.(string)) {
						for subdir, size := range config.Resize[dir].(map[interface{}]interface{}) {
							resizePath := fmt.Sprintf("%s/%s", path, subdir)
							if _, err := os.Stat(resizePath); os.IsNotExist(err) {
								err := os.Mkdir(resizePath, 0755)
								if err != nil {
									log.Println("[error] ", err)
									continue
								}
							}
							h := size.(map[interface{}]interface{})[string("height")]
							w := size.(map[interface{}]interface{})[string("width")]
							log.Printf("height:%d, width:%d", h, w)
							dstFile = fmt.Sprintf("%s/%s", resizePath, basename)
							log.Println("start resize: ", dstFile)
							err := resize(h.(int), w.(int), dstFile, srcFile)
							if err != nil {
								log.Println("[error] ", err)
								continue
							}
						}
					} else {
						dst, err := os.Create(dstFile)
						if err != nil {
							log.Println("[error] ", err)
							continue
						}

						_, err = io.Copy(dst, src)
						if err != nil {
							log.Println("[error] ", err)
							continue
						}
						dst.Close()
					}
					src.Close()

					if isModulateDir(dir.(string)) {
						log.Println("start modulate: ", dstFile)
						m := config.Modulate[dir]
						err := modulate(m.(int), dstFile, srcFile)
						if err != nil {
							log.Println("[error] ", err)
							continue
						}
					}

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

func isResizeDir(d string) bool {
	for dir := range config.Resize {
		if dir.(string) == d {
			return true
		}
	}
	return false
}

func isModulateDir(d string) bool {
	for dir := range config.Modulate {
		if dir.(string) == d {
			return true
		}
	}
	return false
}

func resize(cols, rows int, dstFile, srcFile string) error {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ReadImage(srcFile)
	if err != nil {
		return err
	}

	err = mw.AdaptiveResizeImage(uint(cols), uint(rows))
	if err != nil {
		return err
	}

	// write image
	err = mw.WriteImage(dstFile)
	if err != nil {
		return err
	}

	return nil
}

func modulate(mod int, dstFile, srcFile string) error {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ReadImage(srcFile)
	if err != nil {
		return err
	}

	err = mw.ModulateImage(float64(mod), 100, 100)
	if err != nil {
		return err
	}

	err = mw.WriteImage(dstFile)
	if err != nil {
		return err
	}

	return nil
}
