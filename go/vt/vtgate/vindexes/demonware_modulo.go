/*
Copyright 2026 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vindexes

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/vt/key"
)

const defaultModulo uint64 = 65536

var (
	_ SingleColumn = (*DemonwareModulo)(nil)
	_ Hashing      = (*DemonwareModulo)(nil)

	errNumRangesRequired = errors.New("param num_ranges is required")
)

type DemonwareModulo struct {
	modulo    uint64
	name      string
	numRanges uint16
	ranges    map[uint16]*demonwareModuloRange
	useOffset bool
}

type demonwareModuloRange struct {
	start uint64
	width uint64
}

func init() {
	Register("demonware_modulo", NewDemonwareModulo)
}

func NewDemonwareModulo(name string, params map[string]string) (Vindex, error) {
	modulo := defaultModulo
	if raw, ok := params["modulo"]; ok {
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return nil, err
		}
		modulo = parsed
	}
	if modulo == 0 {
		return nil, errors.New("param modulo must be greater than zero")
	}

	rawRanges, ok := params["num_ranges"]
	if !ok {
		return nil, errNumRangesRequired
	}
	parsedRanges, err := strconv.ParseUint(rawRanges, 10, 16)
	if err != nil {
		return nil, err
	}
	numRanges := uint16(parsedRanges)

	var useOffset bool
	if raw, ok := params["use_offset"]; ok {
		useOffset, err = strconv.ParseBool(raw)
		if err != nil {
			return nil, err
		}
	}
	if useOffset && numRanges == 1 {
		return nil, errors.New("param use_offset requires num_ranges greater than one")
	}

	shardSpecs, err := key.GenerateShardRanges(int(numRanges), 0)
	if err != nil {
		return nil, err
	}
	ranges := make(map[uint16]*demonwareModuloRange, len(shardSpecs))
	for id, shardSpec := range shardSpecs {
		keyRanges, err := key.ParseShardingSpec(shardSpec)
		if err != nil {
			return nil, err
		}
		if len(keyRanges) != 1 {
			return nil, fmt.Errorf("unexpected number (%d) of key ranges parsed from shard spec (%s)", len(keyRanges), shardSpec)
		}
		keyRange := keyRanges[0]
		startBytes := make([]byte, 8)
		copy(startBytes, keyRange.Start)
		start := binary.BigEndian.Uint64(startBytes)
		endBytes := make([]byte, 8)
		copy(endBytes, keyRange.End)
		end := binary.BigEndian.Uint64(endBytes)
		ranges[uint16(id)] = &demonwareModuloRange{start: start, width: end - start}
	}

	return &DemonwareModulo{
		modulo:    modulo,
		name:      name,
		numRanges: numRanges,
		ranges:    ranges,
		useOffset: useOffset,
	}, nil
}

func (vind *DemonwareModulo) String() string {
	return vind.name
}

func (*DemonwareModulo) Cost() int {
	return 1
}

func (*DemonwareModulo) IsUnique() bool {
	return true
}

func (*DemonwareModulo) NeedsVCursor() bool {
	return false
}

func (vind *DemonwareModulo) Verify(_ context.Context, _ VCursor, ids []sqltypes.Value, ksids [][]byte) ([]bool, error) {
	out := make([]bool, 0, len(ids))
	for i, id := range ids {
		ksid, err := vind.Hash(id)
		if err != nil {
			return nil, err
		}
		out = append(out, bytes.Equal(ksid, ksids[i]))
	}
	return out, nil
}

func (vind *DemonwareModulo) Map(_ context.Context, _ VCursor, ids []sqltypes.Value) ([]key.ShardDestination, error) {
	out := make([]key.ShardDestination, 0, len(ids))
	for _, id := range ids {
		ksid, err := vind.Hash(id)
		if err != nil {
			out = append(out, key.DestinationNone{})
			continue
		}
		out = append(out, key.DestinationKeyspaceID(ksid))
	}
	return out, nil
}

func (vind *DemonwareModulo) Hash(id sqltypes.Value) ([]byte, error) {
	num, err := id.ToCastUint64()
	if err != nil {
		return nil, err
	}
	keyID := num % vind.modulo
	rangeID := uint16(keyID % uint64(vind.numRanges))
	keyspaceID := vind.ranges[rangeID].start
	if vind.useOffset && num >= uint64(vind.numRanges) {
		keyspaceID += (num / uint64(vind.numRanges)) % vind.ranges[rangeID].width
	}
	var keybytes [8]byte
	binary.BigEndian.PutUint64(keybytes[:], keyspaceID)
	return keybytes[:], nil
}
