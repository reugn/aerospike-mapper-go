package mapper

// Metadata holds the Aerospike record generation and expiration details.
// Embed this struct into your data model to automatically unpack metadata values
// when reading records from Aerospike, or to specify metadata values when encoding.
type Metadata struct {
	// Generation shows record modification count. Represents the number of times the record has
	// been updated.
	Generation uint32 `aero:"meta,generation"`
	// Expiration is TTL (Time-To-Live). Specifies the number of seconds until the record expires.
	Expiration uint32 `aero:"meta,expiration"`
}

// Key holds the Aerospike record key details.
// Embed this struct into your data model to automatically unpack key values
// when reading records from Aerospike, or to specify key values when encoding.
type Key struct {
	// Namespace is the Aerospike namespace for the record.
	Namespace string `aero:"meta,namespace"`
	// SetName is the Aerospike set name for the record.
	SetName string `aero:"meta,set_name"`
	// Digest is the Aerospike record digest.
	Digest [20]byte `aero:"meta,digest"`
}

// KeyValue holds the Aerospike user key value.
// Embed this struct into your data model to automatically unpack the user key
// when reading records from Aerospike, or to specify the user key when encoding.
type KeyValue struct {
	// UserKey is the user-defined key for the record. It can be of any type supported by the
	// Aerospike client.
	UserKey any `aero:"meta,user_key"`
}
