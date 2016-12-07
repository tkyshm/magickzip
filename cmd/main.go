package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/tkyshm/magickzip"
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
		nextPath := fmt.Sprintf("%s/%s", parent, dir)
		_, err := os.Stat(nextPath)
		if os.IsNotExist(err) {
			log.Println("make directory:", dir)
			err := os.Mkdir(nextPath, 0755)
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
					dstFile := fmt.Sprintf("%s/%s", nextPath, basename)

					// resize
					if isResizeDir(dir.(string)) {
						var cols interface{}
						var rows interface{}
						var resizePath string
						// resize section children
						for subdir, children := range config.Resize[dir].(map[interface{}]interface{}) {
							// checks that subdir exist or not
							switch t := children.(type) {
							case int: // not exist subdir
								if subdir == "height" {
									rows = children
								}
								if subdir == "width" {
									cols = children
								}

								if cols != nil && rows != nil {
									log.Printf("height:%d, width:%d", cols, rows)
									log.Println("start resize: ", dstFile)
									err := resize(cols.(int), rows.(int), dstFile, srcFile)
									if err != nil {
										log.Println("[error] ", err)
									}
								}
							case map[interface{}]interface{}: // exist subdir
								// make directories for resize deeply
								cols, rows, resizePath = makeResizeDirp(nextPath, subdir.(string), children.(map[interface{}]interface{}))
								// failed to make directories
								if cols == nil || rows == nil {
									log.Println("[error] failed to make directories because resize yaml format is wrong")
									continue
								}

								log.Printf("height:%d, width:%d", cols, rows)
								dstFile = fmt.Sprintf("%s/%s", resizePath, basename)
								log.Println("start resize: ", dstFile)
								err := resize(cols.(int), rows.(int), dstFile, srcFile)
								if err != nil {
									log.Println("[error] ", err)
									continue
								}
							default:
								log.Printf("[error] unexpected type: %T", t)
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
			makeStructure(content.(map[interface{}]interface{}), nextPath)
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

func makeResizeDirp(nextPath, subdir string, children map[interface{}]interface{}) (cols, rows interface{}, resizePath string) {
	resizePath = fmt.Sprintf("%s/%s", nextPath, subdir)
	if _, err := os.Stat(resizePath); os.IsNotExist(err) {
		log.Print("[resize] make directory:", resizePath)
		err := os.Mkdir(resizePath, 0755)
		if err != nil {
			log.Println("[error] ", err)
			panic(err)
		}
	}

	cols = children["height"]
	rows = children["width"]
	if cols != nil && rows != nil {
		return cols, rows, resizePath
	}

	if rows == nil && cols == nil {
		if len(children) == 1 {
			for subsubdir, nextChildren := range children {
				return makeResizeDirp(resizePath, subsubdir.(string), nextChildren.(map[interface{}]interface{}))
			}
		} else {
			panic(fmt.Errorf("template 'resize' format is wrong"))
		}
	}

	return cols, rows, resizePath
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
