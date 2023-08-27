package assets

import (
	"context"
	"fmt"
	"immich-go/immich/metadata"
	"io/fs"
	"path"
	"strings"
)

type LocalAssetBrowser struct {
	fs.FS
	albums map[string]string
}

func BrowseLocalAssets(fsys fs.FS) *LocalAssetBrowser {
	return &LocalAssetBrowser{
		FS: fsys,
	}
}

func (fsys LocalAssetBrowser) Stat(name string) (fs.FileInfo, error) {
	return fs.Stat(fsys.FS, name)
}

func (fsys LocalAssetBrowser) Browse(ctx context.Context) chan *LocalAssetFile {
	fileChan := make(chan *LocalAssetFile)
	// Browse all given FS to collect the list of files
	go func(ctx context.Context) {
		defer close(fileChan)
		err := fs.WalkDir(fsys, ".",
			func(name string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				// Check if the context has been cancelled
				select {
				case <-ctx.Done():
					// If the context has been cancelled, return immediately
					return ctx.Err()
				default:
				}
				if d.IsDir() {
					return nil
				}
				ext := strings.ToLower(path.Ext(name))
				switch ext {
				case ".jpg", "jpeg", ".png", ".mp4", ".heic", ".mov", ".gif":

					s, err := d.Info()

					f := LocalAssetFile{
						FSys:     fsys,
						FileName: name,
						Title:    name,
						size:     int(s.Size()),
						Err:      err,
					}
					if fsys.albums[path.Dir(name)] != "" {
						f.AddAlbum(fsys.albums[path.Dir(name)])
					}

					_, err = fs.Stat(fsys, name+".xmp")
					if err == nil {
						f.SideCar = &metadata.SideCarMetadata{
							FileName: name + ".xmp",
							OnFSsys:  true,
						}
					}

					// Check if the context has been cancelled before sending the file
					select {
					case <-ctx.Done():
						// If the context has been cancelled, return immediately
						return ctx.Err()
					case fileChan <- &f:
					}
					err = nil
				}
				return err
			})
		if err != nil {
			// Check if the context has been cancelled before sending the error
			select {
			case <-ctx.Done():
				// If the context has been cancelled, return immediately
				return
			case fileChan <- &LocalAssetFile{
				Err: err,
			}:
			}
		}

	}(ctx)

	return fileChan
}

func (fsys LocalAssetBrowser) BrowseAlbums(ctx context.Context) error {
	fsys.albums = map[string]string{}
	err := fs.WalkDir(fsys, ".",
		func(name string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Check if the context has been cancelled
			select {
			case <-ctx.Done():
				// If the context has been cancelled, return immediately
				return ctx.Err()
			default:
			}
			if name != "." && d.IsDir() {
				fsys.albums[name] = fsys.albums[path.Base(name)]
				return nil
			}
			return nil
		})

	return err
}

func ReadLocalAsset(fsys fs.FS, name string) (*LocalAssetFile, error) {
	ext := strings.ToLower(path.Ext(name))
	switch ext {
	case ".jpg", "jpeg", ".png", ".mp4", ".heic", ".mov", ".gif":
		s, err := fs.Stat(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("can't read asset: %w", err)
		}
		f := LocalAssetFile{
			FSys:     fsys,
			FileName: name,
			Title:    name,
			size:     int(s.Size()),
		}
		_, err = fs.Stat(fsys, name+".xmp")
		if err == nil {
			f.SideCar = &metadata.SideCarMetadata{
				FileName: name + ".xmp",
				OnFSsys:  true,
			}
		}
		return &f, nil
	default:
		return nil, fmt.Errorf("%q not supported: %q", ext, name)
	}
}
