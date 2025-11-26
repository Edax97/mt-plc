package wailonServer

import (
	"encoding/gob"
	"os"
	"time"
)

type SentCache struct {
	sentMap  map[string]time.Time
	diskPath string
}

func NewSentCache(diskPath string) *SentCache {
	cache := &SentCache{
		diskPath: diskPath,
		sentMap:  make(map[string]time.Time),
	}
	cache.LoadCache()
	return cache
}

func (c *SentCache) SaveCache() {
	f, err := os.Create(c.diskPath)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	encoder := gob.NewEncoder(f)
	err = encoder.Encode(c.sentMap)
	if err != nil {
		return
	}

}

func (c *SentCache) LoadCache() {
	f, err := os.Open(c.diskPath)
	if err != nil {
		c.sentMap = make(map[string]time.Time)
		return
	}
	defer func() {
		_ = f.Close()
	}()

	var data map[string]time.Time
	decoder := gob.NewDecoder(f)
	err = decoder.Decode(&data)
	if err != nil {
		c.sentMap = make(map[string]time.Time)
		return
	}
	c.sentMap = data
}

func (c *SentCache) HasSent(imei string, t time.Time) bool {
    sent, ok := c.sentMap[imei]
    if !ok {
        return false
    }
    // Return true only when query time is strictly before the stored time.
    // If equal or after, consider it not sent to avoid duplicates at the same timestamp.
    return t.Before(sent)
}

func (c *SentCache) UpdateSent(imei string, sent time.Time) bool {
	c.sentMap[imei] = sent
	c.SaveCache()
	return true
}
