package scrape

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/goextension/log"
)

// DefaultOutputPath ...
var DefaultOutputPath = "image"

type OutputInfo struct {
	Name        string
	Skip        bool
	Force       bool
	OutputPath  string
	CopyInfo    bool
	InfoPath    string
	InfoName    string
	CopyPoster  bool
	PosterPath  string
	PosterName  string
	CopyThumb   bool
	ThumbPath   string
	ThumbName   string
	CopySample  bool
	SamplePath  string
	SampleName  string
	SampleFiles []string
	ImagePath   string
	InfoExt     string
}

func DefaultOutputOption() *OutputInfo {
	return &OutputInfo{
		Skip:       false,
		OutputPath: "",
		ImagePath:  "image",
		CopyInfo:   false,
		InfoPath:   "",
		InfoName:   "",
		InfoExt:    ".nfo",
		CopyPoster: true,
		PosterPath: "",
		PosterName: "poster",
		CopyThumb:  true,
		ThumbPath:  "",
		ThumbName:  "thumb",
		CopySample: false,
		SamplePath: "",
		SampleName: "sample",
	}
}

func copyCache(cache *Cache, msg *Content, sample bool, output string, force bool) (e error) {
	pid := filepath.Join(output, strings.ToUpper(msg.ID), "."+msg.From)
	e = copyFile(cache, msg.Poster, filepath.Join(pid, "poster"), force)
	if e != nil {
		return e
	}
	e = copyFile(cache, msg.Thumb, filepath.Join(pid, "thumb"), force)
	if e != nil {
		return e
	}
	for _, act := range msg.Actors {
		e = copyFile(cache, act.Image, filepath.Join(pid, ".actor", act.Name), force)
		if e != nil {
			return e
		}
	}
	if sample {
		for _, s := range msg.Sample {
			e = copyFile(cache, s.Image, filepath.Join(pid, ".sample", "sample"+"@"+strconv.Itoa(s.Index)), force)
			if e != nil {
				return e
			}
			e = copyFile(cache, s.Thumb, filepath.Join(pid, ".thumb", "thumb"+"@"+strconv.Itoa(s.Index)), force)
			if e != nil {
				return e
			}
		}
	}
	return nil
}

func copyInfo(msg *Content, path string, name string) error {
	inf := filepath.Join(path, name)
	_ = os.MkdirAll(filepath.Dir(inf), os.ModePerm)
	info, e := os.Stat(inf)
	if e != nil && !os.IsNotExist(e) {
		return e
	}
	if e == nil && info.Size() != 0 {
		return nil
	}
	bytes, e := json.MarshalIndent(msg, "", " ")
	if e != nil {
		return e
	}
	return ioutil.WriteFile(inf, bytes, 0755)
}

func copyFileWithInfo(cache *Cache, content Content, option *OutputInfo) error {
	var e error
	if option.Skip {
		return nil
	}
	if option.CopyInfo {
		if option.InfoName == "@" {
			option.InfoName = content.ID
		}
		e = copyInfo(&content, filepath.Join(option.OutputPath, option.InfoPath), option.InfoName+option.InfoExt)
		if e != nil {
			log.Errorw("OutputCallback", "error", e, "output", content.ID)
		}
	}

	if option.CopyPoster {
		option.PosterName = option.PosterName + Ext(content.Poster)
		if option.PosterPath == "" {
			option.PosterPath = filepath.Join(option.ImagePath, option.PosterPath)
		}
		path := filepath.Join(option.OutputPath, option.PosterPath, option.PosterName)
		if debug {
			log.Infow("CopyFile", "source", content.Poster, "path", path)
		}
		e = copyFile(cache, content.Poster, path, option.Force)
		if e != nil {
			log.Errorw("OutputCallback", "error", e, "output", content.ID)
		}
	}

	if option.CopyThumb {
		option.ThumbName = option.ThumbName + Ext(content.Thumb)
		if option.ThumbPath == "" {
			option.ThumbPath = filepath.Join(option.ImagePath, option.ThumbPath)
		}
		path := filepath.Join(option.OutputPath, option.ThumbPath, option.ThumbName)
		e = copyFile(cache, content.Thumb, path, option.Force)
		if e != nil {
			log.Errorw("OutputCallback", "error", e, "output", content.ID)
		}
	}

	if option.CopySample {
		if option.SamplePath == "" {
			option.SamplePath = filepath.Join(option.ImagePath, option.SamplePath)
		}
		for i, sample := range content.Sample {
			path := filepath.Join(option.OutputPath, option.SamplePath, option.SampleName+"@"+strconv.Itoa(i)+Ext(content.Sample[i].Image))
			option.SampleFiles = append(option.SampleFiles, path)
			e = copyFile(cache, sample.Image, path, option.Force)
			if e != nil {
				log.Errorw("OutputCallback", "error", e, "output", content.ID)
			}
		}
	}
	return e
}

func copyFile(cache *Cache, source, path string, force bool) error {
	if source == "" {
		return nil
	}

	if debug {
		log.Infow("CopyFile", "source", source, "path", path)
	}
	_ = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	info, e := os.Stat(path)

	var bys []byte
	if force {
		bys, e = cache.ForceGet(source)
	} else {
		if e != nil && !os.IsNotExist(e) {
			return e
		}
		if e == nil && info.Size() != 0 {
			return nil
		}
		bys, e = cache.GetBytes(source)
	}
	if e != nil {
		return e
	}
	return ioutil.WriteFile(path, bys, 0755)
}

func imageCache(cache *Cache, m Content, sample bool) (e error) {
	path := make(chan string, 2)
	go func(path chan<- string) {
		defer close(path)
		path <- m.Poster
		path <- m.Thumb
		for _, act := range m.Actors {
			path <- act.Image
		}
		if sample {
			for _, s := range m.Sample {
				path <- s.Image
				path <- s.Thumb
			}
		}
	}(path)

	for p := range path {
		if p != "" {
			_, err := cache.Get(p)
			if err != nil && !os.IsExist(err) {
				log.Error(err)
			}
		}
	}
	return nil
}

// TrimEnd ...
func TrimEnd(source string) string {
	return strings.Split(source, "?")[0]
}

// Ext ...
func Ext(source string) string {
	ext := filepath.Ext(TrimEnd(source))
	if debug {
		log.Infow("ext", "source", source, "ext", ext)
	}
	return ext
}
