package testtypes

import mapper "github.com/reugn/aerospike-mapper-go"

type Item1 struct {
	Title string `aero:"title"`
}

type Item2 struct {
	Name  string `aero:"name"`
	Empty bool   `aero:"empty"`
	Size  uint64 `aero:"size"`
}

type Item struct {
	mapper.Key
	mapper.KeyValue
	mapper.Metadata
	Item1
	*Item2
	Label       string         `aero:"label,omit"`
	Length      int            `aero:"length"`
	Offset      *int           `aero:"offset, omitempty"`
	Description string         `aero:"description ,omitempty"`
	IntList     []int          `aero:"list"`
	Dict        map[string]int `aero:"dict,omitempty"`
}
