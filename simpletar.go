package simpletar

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func tarHelper(dir string, tw *tar.Writer) error {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip the source directory
		if path == dir && info.IsDir() {
			return nil
		}

		// get the relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}

		// check if it is a symbolic link
		linkPath := ""
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			linkPath, err = os.Readlink(path)
		}

		th, err := tar.FileInfoHeader(info, linkPath)
		if err != nil {
			return err
		}

		// fix up the name in the header to be relative
		th.Name = relPath
		if info.IsDir() {
			th.Name = th.Name + "/"
		}

		// write the tar header
		if err := tw.WriteHeader(th); err != nil {
			return err
		}

		// copy the file contents
		if th.Typeflag == tar.TypeReg {
			fp, err := os.Open(path)
			if err != nil {
				return err
			}
			defer fp.Close()
			_, err = io.Copy(tw, fp)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return nil
}

// Tar tars the src directory and outputs the file at specified dest
func Tar(src string, dest string) error {
	fInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !fInfo.IsDir() {
		return fmt.Errorf("source path expected to be a directory: %v", src)
	}

	f, err := os.Create(dest)
	if err != nil {
		return nil
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	defer tw.Close()

	return tarHelper(src, tw)
}

// Untar untars the source file to dest directory
func Untar(src string, dest string) error {
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()
	tr := tar.NewReader(df)

	for {
		th, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		th.Name = filepath.Clean(th.Name)
		path := filepath.Join(dest, th.Name)

		fi := th.FileInfo()
		switch th.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(path, fi.Mode()); err != nil {
				return err
			}

		case tar.TypeReg, tar.TypeRegA:
			nf, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, fi.Mode())
			if err != nil {
				return err
			}
			_, err = io.Copy(nf, tr)
			nf.Close()
			if err != nil {
				return err
			}

		case tar.TypeSymlink:
			if err := os.Symlink(th.Linkname, path); err != nil {
				return nil
			}

		default:
			return fmt.Errorf("unsupported type")
		}

	}
	return nil
}
