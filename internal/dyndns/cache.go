package dyndns

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/lukasdietrich/dynflare/internal/config"
)

type cache struct {
	foldername string
}

func newCache() (*cache, error) {
	userCache, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("could not resolve user cache folder: %w", err)
	}

	foldername := filepath.Join(userCache, "dynflare")
	return &cache{foldername: foldername}, os.MkdirAll(foldername, 0700)
}

func (c *cache) read(domain config.Domain) (string, error) {
	b, err := ioutil.ReadFile(c.deriveFilename(domain))
	if os.IsNotExist(err) {
		return "", nil
	}

	return strings.TrimSpace(string(b)), err
}

func (c *cache) write(domain config.Domain, addr string) error {
	return ioutil.WriteFile(c.deriveFilename(domain), []byte(addr), 0600)
}

func (c *cache) deriveFilename(domain config.Domain) string {
	return filepath.Join(
		c.foldername,
		fmt.Sprintf("%s.%s.txt", url.PathEscape(domain.Name), domain.Kind),
	)
}
