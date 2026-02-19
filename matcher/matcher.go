package matcher

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const (
	SimilarityThreshold = 0.6
)

type Matcher struct {
	rdb *redis.Client
}

func NewMatcher(rdb *redis.Client) *Matcher {
	return &Matcher{rdb: rdb}
}

func (m *Matcher) StoreFingerprint(ctx context.Context, songName string, hashes map[uint64]struct{}) error {
	key := "song:" + songName + ":hashes"

	for h := range hashes {
		hashStr := strconv.FormatUint(h, 10)
		if err := m.rdb.SAdd(ctx, key, hashStr).Err(); err != nil {
			return err
		}
	}

	if err := m.rdb.SAdd(ctx, "songs", songName).Err(); err != nil {
		return err
	}

	return nil
}

func (m *Matcher) FindMatch(ctx context.Context, queryHashes map[uint64]struct{}) (string, float64, error) {
	songs, err := m.rdb.SMembers(ctx, "songs").Result()
	if err != nil {
		return "", 0, err
	}

	if len(songs) == 0 {
		return "", 0, nil
	}

	queryCount := float64(len(queryHashes))
	if queryCount == 0 {
		return "", 0, nil
	}

	var bestMatch string
	var bestSimilarity float64

	for _, song := range songs {
		key := "song:" + song + ":hashes"
		storedHashes, err := m.rdb.SMembers(ctx, key).Result()
		if err != nil {
			continue
		}

		matchingCount := 0
		for _, hStr := range storedHashes {
			h, err := strconv.ParseUint(hStr, 10, 64)
			if err != nil {
				continue
			}
			if _, ok := queryHashes[h]; ok {
				matchingCount++
			}
		}

		similarity := float64(matchingCount) / queryCount

		if similarity >= SimilarityThreshold && similarity > bestSimilarity {
			bestSimilarity = similarity
			bestMatch = song
		}
	}

	return bestMatch, bestSimilarity, nil
}

func (m *Matcher) DeleteSong(ctx context.Context, songName string) error {
	key := "song:" + songName + ":hashes"

	if err := m.rdb.Del(ctx, key).Err(); err != nil {
		return err
	}

	if err := m.rdb.SRem(ctx, "songs", songName).Err(); err != nil {
		return err
	}

	return nil
}

func (m *Matcher) ListSongs(ctx context.Context) ([]string, error) {
	return m.rdb.SMembers(ctx, "songs").Result()
}
