package testtypes

// BinMap is used to define a map of bin names to values.
type BinMap map[string]interface{}

// Record is the container struct for database records.
// Records are equivalent to rows.
type Record struct {
	// Key is the record's key.
	// Might be empty, or may only consist of digest value.
	Key *Key

	// Bins is the map of requested name/value bins.
	Bins BinMap

	// Generation shows record modification count.
	Generation uint32

	// Expiration is TTL (Time-To-Live).
	// Number of seconds until record expires.
	Expiration uint32
}

// BatchRecord encapsulates the Batch key and record result.
type BatchRecord struct {
	// Key.
	Key *Key

	// Record result after batch command has completed. Will be nil if record was not found
	// or an error occurred. See ResultCode.
	Record *Record

	// ResultCode for this returned record.
	ResultCode int

	// Err encapsulates the possible error chain for this key.
	Err error

	// InDoubt signifies the possibility that the write command may have completed even though an error
	// occurred for this record. This may be the case when a client error occurs (like timeout)
	// after the command was sent to the server.
	InDoubt bool

	// Does this command contain a write operation. For internal use only.
	HasWrite bool
}

// BatchRead specifies the Key and bin names used in batch read commands
// where variable bins are needed for each key.
type BatchRead struct {
	BatchRecord

	// BinNames specifies the Bins to retrieve for this key.
	// BinNames are mutually exclusive with Ops.
	BinNames []string

	// ReadAllBins defines what data should be read from the record.
	// If true, ignore binNames and read all bins.
	// If false and binNames are set, read specified binNames.
	// If false and binNames are not set, read record header (generation, expiration) only.
	ReadAllBins bool // = false
}
