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
	"strings"

	"github.com/fatih/color"
	"github.com/ivpusic/grpool"
)

var crcRegex = regexp.MustCompile(`\[([A-Fa-f0-9]{8})]`)

var (
	rootDir    string
	updateHash bool
)

func init() {
	flag.StringVar(&rootDir, "dir", "", "Directory to scan for files")
	flag.BoolVar(&updateHash, "update", false, "Update hash in the filename if it's a mismatch")
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
	crcHashBytes, err := hex.DecodeString(regexMatch[1])
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

	crcCalcBytes := hasher.Sum(nil)
	result := color.RedString("MISMATCH")
	if bytes.Equal(crcHashBytes, crcCalcBytes) {
		result = color.GreenString("OK")
	} else if updateHash {
		result = color.YellowString("UPDATED")
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
		if err := renameFileHash(dir, file, crcHashBytes, crcCalcBytes); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("%s - %s\n", file.Name(), result)
}

func renameFileHash(dir string, file os.FileInfo, crcHashBytes, crcCalcBytes []byte) error {
	crcHash := strings.ToUpper(hex.EncodeToString(crcHashBytes))
	crcCalc := strings.ToUpper(hex.EncodeToString(crcCalcBytes))
	newName := strings.Replace(file.Name(), crcHash, crcCalc, 1)
	return os.Rename(path.Join(dir, file.Name()), path.Join(dir, newName))
}
