package main

import (
	"os"
	"io"
	"strings"
	"archive/tar"
	"compress/gzip"
	"path/filepath"
	"fmt"
)

func main() {

	args := os.Args[1:]

	if (len(args) < 2 || len(args) > 2) {
		fmt.Println("2 arguments required: <source tar.gz> <target dir>")
		return
	}

	// Grab arguments
	input := args[0]
	target := args[1]

	// Get the name of the tar file
	inputTar := input[:strings.LastIndex(input, ".")]

	// Unzip the tar file
	err := ungzip(input, inputTar)
	if (err != nil) {
		fmt.Printf("GZip error: %v\n", err)
		return
	}

	// Explode the tar file
	err = untar(inputTar, target)
	if (err != nil) {
		fmt.Printf("Tar error: %v\n", err)
		return
	}

	// Remove the tar file
	os.Remove(inputTar)
}


func ungzip(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	target = filepath.Join(target, archive.Name)
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, archive)
	return err
}

func untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()

		// Check for directory
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		// Check for symlink
		if header.Typeflag == tar.TypeSymlink {
			os.Symlink(header.Linkname, path)
		}

		// Handle regular files
		if header.Typeflag == tar.TypeReg {
			file, err := os.OpenFile(path, os.O_CREATE | os.O_TRUNC | os.O_WRONLY, info.Mode())
			if err != nil {
				return err
			}
			_, err = io.Copy(file, tarReader)
			if err != nil {
				return err
			}

			file.Close()
		}
	}

	return nil
}