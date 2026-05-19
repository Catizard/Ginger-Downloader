package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rotisserie/eris"
)

type Locale int

const (
	ENGLISH_LOCALE = iota
	CHINESE_LOCALE = iota
)

type field struct {
	Name    string
	Comment map[Locale]string
}

// fields indicates the config options that should be generated and
// the comment along with it. It's possible that the comment is empty,
// the caller side must handle this case.
var fields []field = []field{
	{
		Name: "ClientType",
		Comment: map[Locale]string{
			ENGLISH_LOCALE: "Your game client. Should be either LR2 or Beatoraja",
		},
	},
	{
		Name: "GameInstallationPath",
		Comment: map[Locale]string{
			ENGLISH_LOCALE: `Your game installation path, which should point to the directory. 
For Raja users, it should be the directory that contains the 'beatoraja.jar'
For LR2 users, it should be the directory that contains the 'LR2Body.exe'`,
		},
	},
	{
		Name: "DownloadDirectory",
		Comment: map[Locale]string{
			ENGLISH_LOCALE: "Where you want to download to. Must be a directory.",
		},
	},
}

// CreateTemplateConfig creates a template config file.
//
//  1. locale decides the comment's locale
//  2. ignoreExistedFile when flagged, override the existed file if it existed.
//     Otherwise, if the file is existed, this function will return an os.ErrExist.
//
// Please remember this function won't create a full config file that shows
// all the possible options. For example, it will only creates a config file
// that has "GameInstallationPath" while not "SongDataFilePath", the latter
// one is an advanced feature that should be only used when needed.
func CreateTemplateConfig(locale Locale, ignoreExistedFile bool) error {
	if _, err := os.Stat(configFilePath); err == nil {
		if !ignoreExistedFile {
			return os.ErrExist
		}
		if err := os.Remove(configFilePath); err != nil {
			return eris.Wrap(err, "create template config: trying to remove the existed config file")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return eris.Wrap(err, "create template config: stat returns a bad error")
	}

	var sb strings.Builder
	sb.WriteString("{\n")
	for i, field := range fields {
		if field.Comment != nil {
			if comment, ok := field.Comment[locale]; ok {
				scanner := bufio.NewScanner(strings.NewReader(comment))
				for scanner.Scan() {
					line := scanner.Text()
					fmt.Fprintf(&sb, "	// %s\n", line)
				}
			}
		}
		fmt.Fprintf(&sb, "	\"%s\": \"\"", field.Name)
		if i+1 != len(fields) {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("}")

	f, err := os.Create(configFilePath)
	if err != nil {
		return eris.Wrap(err, "create template config: create file")
	}
	if _, err := io.WriteString(f, sb.String()); err != nil {
		return eris.Wrap(err, "create template config: write file")
	}

	return nil
}
