package mapper_test

import (
	"log"
	"testing"

	mapper "github.com/reugn/aerospike-mapper-go"
	"github.com/reugn/aerospike-mapper-go/internal/assert"
	"github.com/reugn/aerospike-mapper-go/internal/testtypes"
)

func TestMapper_Decode(t *testing.T) {
	record1, err := newTestRecord()
	assert.IsNil(t, err)

	tests := []struct {
		name   string
		record any
	}{
		{
			name:   "record",
			record: record1,
		},
		{
			name: "batchRecord",
			record: testtypes.BatchRead{
				BatchRecord: testtypes.BatchRecord{
					Record: record1,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var item testtypes.Item
			err = mapper.Decode(test.record, &item)
			assert.IsNil(t, err)

			// assert metadata fields
			assert.Equal(t, item.Namespace, "ns1")
			assert.Equal(t, item.SetName, "set1")
			assert.Equal(t, item.UserKey, "key1")
			assert.Equal(t, item.Digest[:], record1.Key.Digest())

			// assert bin fields
			assert.Equal(t, item.Label, "label1")
			assert.Equal(t, item.Length, 10)
			assert.Equal(t, item.Title, "title1")
			assert.Equal(t, item.Description, "")
			assert.Equal(t, item.IntList, []int{1, 2, 3})
			assert.Equal(t, item.Dict, map[string]int{"a": 1, "b": 2, "c": 3})
			assert.IsNil(t, item.Offset)
			assert.IsNil(t, item.Item2)
		})
	}
}

func TestMapper_DecodeNegative(t *testing.T) {
	tests := []struct {
		name     string
		record   any
		expected error
	}{
		{
			name:     "struct",
			record:   testtypes.Item1{},
			expected: mapper.ErrInvalidSource,
		},
		{
			name:     "slice",
			record:   []int{1, 2, 3},
			expected: mapper.ErrInvalidSourceType,
		},
		{
			name:     "int",
			record:   1,
			expected: mapper.ErrInvalidSourceType,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var item testtypes.Item
			err := mapper.Decode(test.record, &item)
			assert.ErrorIs(t, err, test.expected)
		})
	}
}

func TestMapper_Encode(t *testing.T) {
	record1, err := newTestRecord()
	assert.IsNil(t, err)

	var item testtypes.Item
	err = mapper.Decode(record1, &item)
	assert.IsNil(t, err)

	encoded, err := mapper.Encode(&item)
	assert.IsNil(t, err)

	log.Print(encoded)

	// assert metadata fields
	assert.Equal(t, encoded.Namespace, "ns1")
	assert.Equal(t, encoded.SetName, "set1")
	assert.Equal(t, encoded.UserKey, "key1")
	assert.Equal(t, encoded.Digest[:], record1.Key.Digest())

	// assert bin fields
	assert.IsNil(t, encoded.Bins["label"]) // omit tag
	assert.Equal(t, encoded.Bins["length"], 10)
	assert.Equal(t, encoded.Bins["title"], "title1")
	assert.IsNil(t, encoded.Bins["description"]) // omitempty tag
	assert.IsNil(t, encoded.Bins["offset"])      // omitempty tag
	assert.Equal(t, item.IntList, []int{1, 2, 3})
	assert.Equal(t, item.Dict, map[string]int{"a": 1, "b": 2, "c": 3})
}

func newTestRecord() (*testtypes.Record, error) {
	key1, err := testtypes.NewKey("ns1", "set1", "key1")
	if err != nil {
		return nil, err
	}

	bins := map[string]any{
		"label":  "label1",
		"length": 10,
		"title":  "title1",
		"name":   "name1",
		"empty":  true,
		"list":   []int{1, 2, 3},
		"dict":   map[string]int{"a": 1, "b": 2, "c": 3},
	}

	return &testtypes.Record{
		Key:        key1,
		Bins:       bins,
		Generation: 3,
		Expiration: 1_000,
	}, nil
}
