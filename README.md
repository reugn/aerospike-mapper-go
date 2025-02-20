# aerospike-mapper-go

This library provides a mechanism to map Go structs to [Aerospike](https://aerospike.com) records and conversely,
using struct field tags to define the mapping.

This module does not depend on the [Aerospike client](https://github.com/aerospike/aerospike-client-go)
or any other external package.

## Key Features

- **Struct Tag Mapping:** Define how struct fields correspond to Aerospike record bins using struct tags.
- **Encode:** Convert a Go struct instance into a `mapper.Record` for further writing to Aerospike.
- **Decode:** Populate a Go struct instance from an Aerospike record or an existing `mapper.Record`.

## Usage

The following steps illustrate the process of record mapping and transformation.

### Tags

The following tags are used to control the mapping behavior. The primary tag is `aero`.

* `aero:"<bin_name>"`: Maps the struct field to the Aerospike bin named <bin_name>. If no tag
  is present, the field is ignored.
* `aero:"meta"`: Used within standard [metadata structs](#field-mapping). It allows you to map specific
  metadata attributes (like generation or expiration) to fields within your struct.
    ```go
    type MetadataFields struct {
        Generation uint32 `aero:"meta,generation"`
    }
    ```
* `aero:"omit"`: Prevents the field from being encoded into or decoded from an Aerospike record.
  The field will be ignored.
* `aero:"omitempty"`: When encoding, the field will only be encoded if its value is not the zero
  value for its type (e.g., 0 for int, "" for string, nil for pointers/slices/maps). When
  decoding, this tag has no effect; the field will be populated if the bin exists in the record.

### Field Mapping

The library provides the following structs with tagged fields that can be embedded into your
data structures to map Aerospike record metadata and key information:

* `mapper.Metadata`: Holds the Aerospike record generation and expiration details.
* `mapper.Key`: Holds the Aerospike record key details.
* `mapper.KeyValue`: Holds the Aerospike user key value.

These standard structs can be used as follows:

```go
type Item struct {
    mapper.Metadata        // embed metadata
    mapper.Key             // embed key details
    Size            int    `aero:"item_size"` // bin
    Name            string `aero:"item_name"` // bin
}
```

### Encode

To encode a struct into a `mapper.Record`:

```go
import (
    mapper "github.com/reugn/aerospike-mapper-go"
)

encodedRecord, err := mapper.Encode(&item)
// handle the error
```

### Decode

To decode an Aerospike record or a `mapper.Record` into a struct:

```go
import (
    mapper "github.com/reugn/aerospike-mapper-go"
)

var item Item
err = mapper.Decode(aerospikeRecord, &item)
// handle the error
```

## License

Licensed under the Apache 2.0 license.
