package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"runtime"

	"github.com/fatih/color"
	"github.com/ivpusic/grpool"
)

var crcRegex = regexp.MustCompile(`\[([A-Fa-f0-9]{8})]`)
var rootDir string

func init() {
	flag.StringVar(&rootDir, "dir", "", "Directory to scan for files")
	flag.Usage = func() {
		fmt.Printf("%s - Extracts the CRC value from file names and validates the files' integrity."+
			"\n\nOptions:\n\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if len(rootDir) == 0 {
		var err error
		rootDir, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}

	files, err := ioutil.ReadDir(rootDir)
	if err != nil {
		log.Fatal(err)
	}

	pool := grpool.NewPool(runtime.GOMAXPROCS(0), 100)
	defer pool.Release()
	pool.WaitCount(len(files))

	for _, file := range files {
		f := file
		pool.JobQueue <- func() {
			defer pool.JobDone()
			checkCRC(rootDir, f)
		}
	}

	pool.WaitAll()
}

func checkCRC(dir string, file os.FileInfo) {
	regexMatch := crcRegex.FindStringSubmatch(file.Name())
	if len(regexMatch) != 2 {
		return
	}
	crcHash, err := hex.DecodeString(regexMatch[1])
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Open(path.Join(dir, file.Name()))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	hasher := crc32.NewIEEE()
	io.Copy(hasher, f)

	crcCalc := hasher.Sum(nil)
	result := color.RedString("MISMATCH")
	if bytes.Equal(crcHash, crcCalc) {
		result = color.GreenString("OK")
	}

	fmt.Printf("%s - %s\n", file.Name(), result)
}
