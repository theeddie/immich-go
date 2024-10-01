package fshelper

import (
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
)

//  GlobWalkFS creates a FS that limits the WalkDir function to the
//  list of files that match the glob expression, and cheats to
//  matches *.XMP files in all circumstances
//  It implements ReadDir and Stat to filter the file list

type GlobWalkFS struct {
    rootFS fs.FS
    dir    string
    parts  []string
}

func NewGlobWalkFS(pattern string) (fs.FS, error) {
    dir, magic := FixedPathAndMagic(pattern)
    if magic == "" {
        s, err := os.Stat(dir)
        if err != nil {
            return nil, err
        }

        if !s.IsDir() {
            magic = strings.ToLower(filepath.Base(dir))
            dir = filepath.Dir(dir)
            if dir == "" {
                dir, _ = os.Getwd()
            }
            return &GlobWalkFS{
                rootFS: NewFSWithName(os.DirFS(dir), filepath.Base(dir)),
            }, nil
        }
    }
    return nil, fmt.Errorf("invalid pattern")
}

// WalkDir function to skip specific directories and files
func WalkDir(root string, ignoreFolders map[string]bool) error {
    return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        // Skip directories in the ignore list
        if d.IsDir() && ignoreFolders[d.Name()] {
            fmt.Printf("Skipping directory: %s\n", path)
            return filepath.SkipDir
        }

        // Skip specific files like .DS_Store and SYNOFILE_THUMB_*.*
        if !d.IsDir() && (d.Name() == ".DS_Store" || strings.HasPrefix(d.Name(), "SYNOFILE_THUMB_")) {
            fmt.Printf("Skipping file: %s\n", path)
            return nil
        }

        // Process other files
        if !d.IsDir() {
            fmt.Printf("Scanning file: %s\n", path)
        }

        return nil
    })
}

func (f *GlobWalkFS) Open(name string) (fs.File, error) {
    // Original Open function logic
    return f.rootFS.Open(name)
}

func (f *GlobWalkFS) Stat(name string) (fs.FileInfo, error) {
    // Original Stat function logic
    return os.Stat(filepath.Join(f.dir, name))
}

func (f *GlobWalkFS) ReadDir(name string) ([]fs.DirEntry, error) {
    // Original ReadDir function logic
    return fs.ReadDir(f.rootFS, name)
}
