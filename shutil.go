package shutil

import (
  "fmt"
  "io"
  "os"
  "path/filepath"
)


type SameFileError struct {
  Src string
  Dst string
}

func (e SameFileError) Error() string {
  return fmt.Sprintf("%s and %s are the same file", e.Src, e.Dst)
}

type SpecialFileError struct {
  File string
  FileInfo os.FileInfo
}

func (e SpecialFileError) Error() string {
  return fmt.Sprintf("`%s` is a named pipe", e.File)
}


func samefile(src string, dst string) bool {
  srcInfo, _ := os.Stat(src)
  dstInfo, _ := os.Stat(dst)
  return os.SameFile(srcInfo, dstInfo)
}

func specialfile(fi os.FileInfo) bool {
  return (fi.Mode() & os.ModeNamedPipe) == os.ModeNamedPipe
}

func IsSymlink(fi os.FileInfo) bool {
  return (fi.Mode() & os.ModeSymlink) == os.ModeSymlink
}


// Copy data from src to dst
//
// If followSymlinks is not set and src is a symbolic link, a
// new symlink will be created instead of copying the file it points
// to.
func CopyFile(src, dst string, followSymlinks bool) (error) {
  if samefile(src, dst) {
    return &SameFileError{src, dst}
  }

  // Make sure src exists and neither are special files
  srcStat, err := os.Lstat(src)
  if err != nil {
    return err
  }
  if specialfile(srcStat) {
    return &SpecialFileError{src, srcStat}
  }

  dstStat, err := os.Stat(dst)
  if err != nil && !os.IsNotExist(err) {
    return err
  } else if err == nil {
    if specialfile(dstStat) {
      return &SpecialFileError{dst, dstStat}
    }
  }

  // If we don't follow symlinks and it's a symlink, just link it and be done
  if !followSymlinks && IsSymlink(srcStat) {
    return os.Symlink(src, dst)
  }

  // If we are a symlink, follow it
  if IsSymlink(srcStat) {
    src, err = os.Readlink(src)
    if err != nil {
      return err
    }
    srcStat, err = os.Stat(src)
    if err != nil {
      return err
    }
  }

  // Do the actual copy
  fsrc, err := os.Open(src)
  if err != nil {
    return err
  }
  defer fsrc.Close()

  fdst, err := os.Create(dst)
  if err != nil {
    return err
  }
  defer fdst.Close()

  size, err := io.Copy(fdst, fsrc)
  if err != nil {
    return err
  }

  if size != srcStat.Size() {
    return fmt.Errorf("%s: %d/%d copied", src, size, srcStat.Size())
  }

  return nil
}


// Copy mode bits from src to dst.
//
// If followSymlinks is false, symlinks aren't followed if and only
// if both `src` and `dst` are symlinks. If `lchmod` isn't available
// and both are symlinks this does nothing. (I don't think lchmod is
// available in Go)
func CopyMode(src, dst string, followSymlinks bool) error {
  srcStat, err := os.Lstat(src)
  if err != nil {
    return err
  }

  dstStat, err := os.Lstat(dst)
  if err != nil {
    return err
  }

  // They are both symlinks and we can't change mode on symlinks.
  if !followSymlinks && IsSymlink(srcStat) && IsSymlink(dstStat) {
    return nil
  }

  // Atleast one is not a symlink, get the actual file stats
  srcStat, _ = os.Stat(src)
  err = os.Chmod(dst, srcStat.Mode())
  return err
}


// Copy data and mode bits ("cp src dst"). Return the file's destination.
//
// The destination may be a directory.
//
// If followSymlinks is false, symlinks won't be followed. This
// resembles GNU's "cp -P src dst".
//
// If source and destination are the same file, a SameFileError will be
// rased.
func Copy(src, dst string, followSymlinks bool) (string, error){
  dstInfo, err := os.Stat(dst)

  if err == nil && dstInfo.Mode().IsDir() {
    dst = filepath.Join(dst, filepath.Base(src))
  }

  if err != nil && !os.IsNotExist(err) {
    return dst, err
  }

  err = CopyFile(src, dst, followSymlinks)
  if err != nil {
    return dst, err
  }

  err = CopyMode(src, dst, followSymlinks)
  if err != nil {
    return dst, err
  }

  return dst, nil
}
