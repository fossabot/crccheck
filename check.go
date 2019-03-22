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
	"regexp"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/ivpusic/grpool"
	"github.com/spf13/afero"
)

// crcRegex is a regex to find the CRC32-Hash within a string.
var crcRegex = regexp.MustCompile(`\[([A-Fa-f0-9]{8})]`)

// check reads the CRC32-Hash from all files in the given dir (excluding sub-folders) and validates the hashes against
// the file contents. It optionally updates the hashes in case of a mismatch.
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

// checkCRC reads the CRC32-Hash from a given file and compares it against the hash of the file content. The result is
// printed to os.Stdout. If update is true, the file will be renamed to include the new hash in case of a mismatch.
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

// extractHash extracts the CRC32-Hash from the given file name (or any string) in the format [XXXXXXXX], where X is
// a hex character so any of 0-9, a-f or A-F. In case string doesn't contain a hash, a nil byte slice is returned
// which NOT an error.
func extractHash(name string) ([]byte, error) {
	regexMatch := crcRegex.FindStringSubmatch(name)
	if len(regexMatch) != 2 {
		return nil, nil
	}
	return hex.DecodeString(regexMatch[1])
}

// calculateHash reads the content of the given file from the file system and computes a new CRC32-Hash.
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

// renameFileHash replaces the hash of a file with the new hash by replacing the hash value in the file name,
// effectively renaming the file.
func renameFileHash(fs afero.Fs, dir string, file os.FileInfo, crcHashBytes, crcCalcBytes []byte) error {
	crcHash := strings.ToUpper(hex.EncodeToString(crcHashBytes))
	crcCalc := strings.ToUpper(hex.EncodeToString(crcCalcBytes))
	newName := strings.Replace(file.Name(), "["+crcHash+"]", "["+crcCalc+"]", 1)
	return fs.Rename(path.Join(dir, file.Name()), path.Join(dir, newName))
}
