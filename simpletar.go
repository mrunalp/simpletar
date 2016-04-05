package simpletar

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
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

		// handle block/char devices
		st, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("could not convert to syscall.Stat_t")
		}
		if st.Mode&syscall.S_IFBLK != 0 || st.Mode&syscall.S_IFCHR != 0 {
			th.Devmajor = int64(major(uint64(st.Rdev)))
			th.Devminor = int64(minor(uint64(st.Rdev)))
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

func major(device uint64) uint64 {
	return (device >> 8) & 0xfff
}

func minor(device uint64) uint64 {
	return (device & 0xff) | ((device >> 12) & 0xfff00)
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
	df, err := os.Open(src)
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
				return err
			}

		case tar.TypeBlock:
			mode := uint32(fi.Mode() & 07777)
			mode |= syscall.S_IFBLK
			if err := syscall.Mknod(path, mode, int(mkdev(th.Devmajor, th.Devminor))); err != nil {
				return err
			}

		case tar.TypeChar:
			mode := uint32(fi.Mode() & 07777)
			mode |= syscall.S_IFCHR
			if err := syscall.Mknod(path, mode, int(mkdev(th.Devmajor, th.Devminor))); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported type")
		}

	}
	return nil
}

func mkdev(major int64, minor int64) uint32 {
	return uint32(((minor & 0xfff00) << 12) | ((major & 0xfff) << 8) | (minor & 0xff))
}
