package java

import (
	"archive/zip"
	"errors"
	"path/filepath"

	"github.com/restartfu/portmypack/portmypack/image"
	"github.com/restartfu/portmypack/portmypack/internal/fsutil"
	"github.com/restartfu/portmypack/portmypack/internal/logger"
)

type ResourcePack struct {
	Name string
	r    *zip.Reader

	Skies []image.Texture

	PackIcon  image.Texture
	Icons     image.Texture
	Particles image.Texture

	Items  []image.Texture
	Blocks []image.Texture
	Armors []image.Texture
}

func NewResourcePack(filePath string) (ResourcePack, error) {
	pck := ResourcePack{}
	pck.Name = filepath.Base(filePath)

	r, err := fsutil.OpenZip(filePath)
	if err != nil {
		return ResourcePack{}, errors.New("error opening java pack zip: " + err.Error())
	}

	path, found := fsutil.FindDirectory(r, "world0")
	if found {
		pck.Skies, err = findSkies(r, path)
		if err != nil {
			logger.Debugf("error trying to find skies: %s", err)
		}
	}

	pck.PackIcon, err = image.NewTextureFS(r, "pack.png", false)
	if err != nil {
		logger.Debugf("error trying to find icon: %s", err)
	}

	texturesPath, found := fsutil.FindDirectory(r, "textures")
	if !found {
		return ResourcePack{}, errors.New("could not find textures path, stopping progress.")
	}

	pck.Icons, err = image.NewTextureFS(r, texturesPath+"/gui/icons.png", true)
	if err != nil {
		logger.Debugf("error trying to find icon: %s", err)
	}
	pck.Particles, err = image.NewTextureFS(r, texturesPath+"/particle/particles.png", false)
	if err != nil {
		logger.Debugf("error trying to find icon: %s", err)
	}

	// Minecraft renamed these folders from the plural "items"/"blocks" to the
	// singular "item"/"block" in the 1.13 flattening. Most packs in the wild
	// today use the new names, but some legacy packs still use the old ones,
	// so check both and merge whatever is found instead of hardcoding one.
	pck.Items, err = directoryTexturesAny(r, texturesPath, []string{"item", "items"}, true)
	if err != nil {
		logger.Debugf("error trying to find items: %s", err)
	}
	pck.Blocks, err = directoryTexturesAny(r, texturesPath, []string{"block", "blocks"}, false)
	if err != nil {
		logger.Debugf("error trying to find blocks: %s", err)
	}
	pck.Armors, err = image.DirectoryTexturesFS(r, texturesPath+"/models/armor", false)
	if err != nil {
		logger.Debugf("error trying to find armor: %s", err)
	}

	pck.r = r
	return pck, nil
}

// directoryTexturesAny loads textures from the first of several candidate
// subdirectory names that exists (or merges all that do), since Minecraft's
// texture folder naming has changed across versions.
func directoryTexturesAny(r *zip.Reader, texturesPath string, names []string, alphaFix bool) ([]image.Texture, error) {
	var all []image.Texture
	var lastErr error
	for _, name := range names {
		textures, err := image.DirectoryTexturesFS(r, texturesPath+"/"+name, alphaFix)
		if err != nil {
			lastErr = err
			continue
		}
		all = append(all, textures...)
	}
	if len(all) == 0 && lastErr != nil {
		return nil, lastErr
	}
	return all, nil
}
