package main

import (
    "os"
    "io"
    "fmt"
    "archive/tar"
    "compress/gzip"
	"path/filepath"
)

func main() {

	outFile, err := os.Create("archive.tar.gz")
	if err != nil {
		panic("Error creating archive file")
	}
	defer outFile.Close()

	err = createTarAndGz("./examples/job2", outFile)

	if err != nil {
		panic("Error creating archive file.")
	}

	fmt.Println("Archiving and file compression completed.")
}

func createTarAndGz(src string, buffer io.Writer) error {

	gzipWriter := gzip.NewWriter(buffer)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	filepath.Walk(src, addToTar(tarWriter))

	return nil
}

func addToTar(tarWriter *tar.Writer) func(string, os.FileInfo, error) error {
    return func(file string, fi os.FileInfo, err error) error {

        header, err := tar.FileInfoHeader(fi, file)
        if err != nil {
            return err
        }

        header.Name = filepath.ToSlash(file)
        if err := tarWriter.WriteHeader(header); err != nil {
            return err
        }

        if !fi.IsDir() {
            data, err := os.Open(file)
            if err != nil {
                return err
            }
            if _, err := io.Copy(tarWriter, data); err != nil {
                return err
            }
        }
        return nil
    }
}
