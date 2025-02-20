package benchmarks

import (
	"testing"

	mapper "github.com/reugn/aerospike-mapper-go"
	"github.com/reugn/aerospike-mapper-go/internal/testtypes"
)

func BenchmarkMapper_Decode(b *testing.B) {
	record, err := newTestRecord()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var item testtypes.Item
		_ = mapper.Decode(record, &item)
	}
}

func BenchmarkMapper_DecodeBatch(b *testing.B) {
	record, err := newTestRecord()
	if err != nil {
		b.Fatal(err)
	}
	batchRecord := &testtypes.BatchRead{
		BatchRecord: testtypes.BatchRecord{
			Record: record,
		},
	}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var item testtypes.Item
		_ = mapper.Decode(batchRecord, &item)
	}
}

func BenchmarkMapper_Encode(b *testing.B) {
	record, err := newTestRecord()
	if err != nil {
		b.Fatal(err)
	}
	var item testtypes.Item
	err = mapper.Decode(record, &item)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_, _ = mapper.Encode(&item)
	}
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
