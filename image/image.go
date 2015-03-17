package image

import (
	"fmt"
	. "github.com/sigmonsays/cask/util"
	"path/filepath"
)

var SupportedFormats = map[string]bool{
	".tar.gz": true,
}

func LocateImage(lxcpath, image string) (string, error) {
	extlist := make([]string, 0)
	for ext, sup := range SupportedFormats {
		if sup == false {
			continue
		}
		extlist = append(extlist, ext)
		archive := filepath.Join(lxcpath, image) + ext
		if FileExists(archive) {
			return archive, nil
		}
	}
	return "", fmt.Errorf("image %s not found, tried %s extensions", image, extlist)
}
