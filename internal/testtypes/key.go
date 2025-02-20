package testtypes

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
)

// Key is the unique record identifier. Records can be identified using a specified namespace,
// an optional set name, and a user defined key which must be unique within a set.
// Records can also be identified by namespace/digest which is the combination used
// on the server.
type Key struct {
	// namespace. Equivalent to database name.
	namespace string

	// Optional set name. Equivalent to database table.
	setName string

	// Unique server hash value generated from set name and user key.
	digest [20]byte

	// Original user key. This key is immediately converted to a hash digest.
	// This key is not used or returned by the server by default. If the user key needs
	// to persist on the server, use one of the following methods:
	//
	// Set "WritePolicy.sendKey" to true. In this case, the key will be sent to the server
	// for storage on writes and retrieved on multi-record scans and queries.
	// Explicitly store and retrieve the key in a bin.
	userKey Value
}

// Namespace returns key's namespace.
func (ky *Key) Namespace() string {
	return ky.namespace
}

// SetName returns key's set name.
func (ky *Key) SetName() string {
	return ky.setName
}

// Value returns key's value.
func (ky *Key) Value() Value {
	return ky.userKey
}

// SetValue sets the Key's value and recompute its digest without allocating new memory.
// This allows the keys to be reusable.
func (ky *Key) SetValue(val Value) error {
	ky.userKey = val
	return ky.computeDigest()
}

// Digest returns key digest.
func (ky *Key) Digest() []byte {
	return ky.digest[:]
}

// Equals uses key digests to compare key equality.
func (ky *Key) Equals(other *Key) bool {
	return bytes.Equal(ky.digest[:], other.digest[:])
}

// String implements Stringer interface and returns string representation of key.
func (ky *Key) String() string {
	if ky == nil {
		return ""
	}

	if ky.userKey != nil {
		return fmt.Sprintf("%s:%s:%s:%v", ky.namespace, ky.setName, ky.userKey.String(),
			fmt.Sprintf("% 02x", ky.digest[:]))
	}
	return fmt.Sprintf("%s:%s::%v", ky.namespace, ky.setName, fmt.Sprintf("% 02x", ky.digest[:]))
}

// NewKey initializes a key from namespace, optional set name and user key.
// The set name and user defined key are converted to a digest before sending to the server.
// The server handles record identifiers by digest only.
func NewKey(namespace string, setName string, key interface{}) (*Key, error) {
	newKey := &Key{
		namespace: namespace,
		setName:   setName,
		userKey:   NewValue(key),
	}

	if err := newKey.computeDigest(); err != nil {
		return nil, err
	}

	return newKey, nil
}

// NewKeyWithDigest initializes a key from namespace, optional set name and user key.
// The server handles record identifiers by digest only.
func NewKeyWithDigest(namespace string, setName string, key interface{}, digest []byte) (*Key, error) {
	newKey := &Key{
		namespace: namespace,
		setName:   setName,
		userKey:   NewValue(key),
	}

	if err := newKey.SetDigest(digest); err != nil {
		return nil, err
	}
	return newKey, nil
}

// SetDigest sets a custom hash
func (ky *Key) SetDigest(digest []byte) error {
	if len(digest) != 20 {
		return errors.New("digest is required to be exactly 20 bytes")
	}
	copy(ky.digest[:], digest)
	return nil
}

func (ky *Key) computeDigest() error {
	data := fmt.Sprintf("%s:%s", ky.setName, ky.userKey)
	h := sha256.New()
	h.Write([]byte(data))
	hash := h.Sum(nil)

	copy(ky.digest[:], hash[:20])
	return nil
}

// PartitionId returns the partition that the key belongs to.
func (ky *Key) PartitionId() int {
	return 1
}
