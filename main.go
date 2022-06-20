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

	err = compress("./examples/job2", outFile)
	if err != nil {
		panic("Error compressing archive")
	}

    err = uncompress("archive.tar.gz", "/tmp")
	if err != nil {
		panic("Error unpacking archive")
	}

	fmt.Println("Archiving and file compression completed.")
}

func compress(src string, buffer io.Writer) error {

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

func uncompress(tarball, dst string) error {

    reader, err := os.Open(tarball)
    if err != nil {
        return err
    }
    defer reader.Close()

	// ungzip
	zr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}

	// untar
	tr := tar.NewReader(zr)

	// uncompress each element
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		// target := 
		//
		// // validate name against path traversal
		// if !validRelPath(header.Name) {
		// 	return fmt.Errorf("tar contained invalid name error %q\n", target)
		// }

		// add dst + re-format slashes according to system
        target := filepath.Join(dst, header.Name)
		// if no join is needed, replace with ToSlash:
		// target = filepath.ToSlash(header.Name)

		// check the type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it (with 0755 permission)
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		// if it's a file create it (with same permission)
		case tar.TypeReg:
			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			// copy over contents
			if _, err := io.Copy(fileToWrite, tr); err != nil {
				return err
			}
			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			fileToWrite.Close()
		}
	}

    return nil
}
