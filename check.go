package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/ivpusic/grpool"
	"github.com/spf13/afero"
)

func check(dir string, update bool) error {
	fs := afero.NewOsFs()
	files, err := afero.ReadDir(fs, dir)
	if err != nil {
		return err
	}

	pool := grpool.NewPool(runtime.GOMAXPROCS(0), 100)
	defer pool.Release()
	pool.WaitCount(len(files))

	for _, file := range files {
		f := file
		pool.JobQueue <- func() {
			defer pool.JobDone()
			if err := checkCRC(fs, dir, f, update); err != nil {
				log.Fatal(err)
			}
		}
	}

	pool.WaitAll()
	return nil
}

func checkCRC(fs afero.Fs, dir string, file os.FileInfo, update bool) error {
	crcHashBytes, err := extractHash(file.Name())
	if err != nil {
		return err
	}
	if crcHashBytes == nil {
		return nil
	}
	crcCalcBytes, err := calculateHash(fs, path.Join(dir, file.Name()))
	if err != nil {
		return err
	}

	result := color.RedString("MISMATCH")
	if bytes.Equal(crcHashBytes, crcCalcBytes) {
		result = color.GreenString("OK")
	} else if update {
		if err := renameFileHash(fs, dir, file, crcHashBytes, crcCalcBytes); err != nil {
			return err
		}
		result = color.YellowString("UPDATED")
	}

	fmt.Printf("%s - %s\n", file.Name(), result)
	return nil
}

func extractHash(name string) ([]byte, error) {
	regexMatch := crcRegex.FindStringSubmatch(name)
	if len(regexMatch) != 2 {
		return nil, nil
	}
	return hex.DecodeString(regexMatch[1])
}

func calculateHash(fs afero.Fs, name string) ([]byte, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	hasher := crc32.NewIEEE()
	if _, err := io.Copy(hasher, f); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

func renameFileHash(fs afero.Fs, dir string, file os.FileInfo, crcHashBytes, crcCalcBytes []byte) error {
	crcHash := strings.ToUpper(hex.EncodeToString(crcHashBytes))
	crcCalc := strings.ToUpper(hex.EncodeToString(crcCalcBytes))
	newName := strings.Replace(file.Name(), "["+crcHash+"]", "["+crcCalc+"]", 1)
	return fs.Rename(path.Join(dir, file.Name()), path.Join(dir, newName))
}
