package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// VideoLinkKeyPrefix is the prefix to retrieve all VideoLink
	VideoLinkKeyPrefix = "VideoLink/value/"
)

// VideoLinkKey returns the store key to retrieve a VideoLink from the index fields
func VideoLinkKey(
	index string,
) []byte {
	var key []byte

	indexBytes := []byte(index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
