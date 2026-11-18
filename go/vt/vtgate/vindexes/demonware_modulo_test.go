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
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"vitess.io/vitess/go/sqltypes"
	"vitess.io/vitess/go/vt/key"
)

func TestDemonwareModuloRouting(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
		ids    []sqltypes.Value
		ksids  []string
	}{
		{
			name:   "default modulo without offset",
			params: map[string]string{"num_ranges": "2"},
			ids:    []sqltypes.Value{sqltypes.NewInt64(0), sqltypes.NewInt64(1), sqltypes.NewInt64(2), sqltypes.NewInt64(65536)},
			ksids:  []string{"0000000000000000", "8000000000000000", "0000000000000000", "0000000000000000"},
		},
		{
			name:   "offset",
			params: map[string]string{"num_ranges": "2", "use_offset": "true"},
			ids:    []sqltypes.Value{sqltypes.NewInt64(0), sqltypes.NewInt64(1), sqltypes.NewInt64(2), sqltypes.NewInt64(3)},
			ksids:  []string{"0000000000000000", "8000000000000000", "0000000000000001", "8000000000000001"},
		},
		{
			name:   "custom modulo",
			params: map[string]string{"num_ranges": "2", "modulo": "3"},
			ids:    []sqltypes.Value{sqltypes.NewInt64(2), sqltypes.NewInt64(3)},
			ksids:  []string{"0000000000000000", "0000000000000000"},
		},
		{
			name:   "uneven ranges",
			params: map[string]string{"num_ranges": "20"},
			ids:    []sqltypes.Value{sqltypes.NewInt64(1), sqltypes.NewInt64(10), sqltypes.NewInt64(19)},
			ksids:  []string{"0c00000000000000", "8000000000000000", "f300000000000000"},
		},
		{
			name:   "large unsigned ids",
			params: map[string]string{"num_ranges": "22", "use_offset": "true"},
			ids:    []sqltypes.Value{sqltypes.NewUint64(17673838484427429135), sqltypes.NewUint64(18032941064968782331)},
			ksids:  []string{"3a26184b30737e0c", "3a6015d598c608d1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vindex, err := CreateVindex("demonware_modulo", t.Name(), tt.params)
			require.NoError(t, err)
			modulo := vindex.(*DemonwareModulo)
			destinations, err := modulo.Map(context.Background(), nil, tt.ids)
			require.NoError(t, err)
			for i, ksid := range tt.ksids {
				require.Equal(t, "DestinationKeyspaceID("+ksid+")", destinations[i].String())
			}
		})
	}
}

func TestDemonwareModuloTwentyRanges(t *testing.T) {
	starts := []byte{
		0x00, 0x0c, 0x19, 0x26, 0x33, 0x40, 0x4c, 0x59, 0x66, 0x73,
		0x80, 0x8c, 0x99, 0xa6, 0xb3, 0xc0, 0xcc, 0xd9, 0xe6, 0xf3,
	}

	for _, useOffset := range []bool{false, true} {
		t.Run(strconv.FormatBool(useOffset), func(t *testing.T) {
			params := map[string]string{"num_ranges": "20", "use_offset": strconv.FormatBool(useOffset)}
			vindex, err := CreateVindex("demonware_modulo", t.Name(), params)
			require.NoError(t, err)

			ids := make([]sqltypes.Value, 0, 61)
			for id := int64(0); id < 60; id++ {
				ids = append(ids, sqltypes.NewInt64(id))
			}
			ids = append(ids, sqltypes.NULL)

			destinations, err := vindex.(*DemonwareModulo).Map(context.Background(), nil, ids)
			require.NoError(t, err)
			require.Len(t, destinations, len(ids))
			for id := 0; id < 60; id++ {
				offset := 0
				if useOffset {
					offset = id / len(starts)
				}
				require.Equal(t,
					fmt.Sprintf("DestinationKeyspaceID(%02x%014x)", starts[id%len(starts)], offset),
					destinations[id].String(), "id=%d", id)
			}
			require.Equal(t, key.DestinationNone{}, destinations[60])
		})
	}
}

func TestDemonwareModuloExternalShardLayouts(t *testing.T) {
	for _, numRanges := range []int{98, 196, 206, 214, 322} {
		t.Run(strconv.Itoa(numRanges), func(t *testing.T) {
			vindex, err := CreateVindex("demonware_modulo", t.Name(), map[string]string{
				"num_ranges": strconv.Itoa(numRanges),
			})
			require.NoError(t, err)

			legacySpecs := legacyModuloShardSpecs(numRanges)
			currentSpecs, err := key.GenerateShardRanges(numRanges, 0)
			require.NoError(t, err)
			require.Len(t, currentSpecs, numRanges)

			legacyMiddle, _, _ := strings.Cut(legacySpecs[numRanges/2], "-")
			currentMiddle, _, _ := strings.Cut(currentSpecs[numRanges/2], "-")
			if numRanges <= 256 {
				require.Equal(t, "7f", legacyMiddle)
				require.Equal(t, "80", currentMiddle)
			} else {
				require.Equal(t, "7fff", legacyMiddle)
				require.Equal(t, "8000", currentMiddle)
			}

			ids := make([]uint64, 0, 3*numRanges+4)
			for id := uint64(0); id < uint64(3*numRanges); id++ {
				ids = append(ids, id)
			}
			ids = append(ids, 65535, 65536, 65537, math.MaxUint64)

			for name, specs := range map[string][]string{"v18": legacySpecs, "v23": currentSpecs} {
				t.Run(name, func(t *testing.T) {
					for _, id := range ids {
						destinations, err := vindex.(*DemonwareModulo).Map(context.Background(), nil,
							[]sqltypes.Value{sqltypes.NewUint64(id)})
						require.NoError(t, err)
						require.Len(t, destinations, 1)
						ksid, ok := destinations[0].(key.DestinationKeyspaceID)
						require.True(t, ok, "id=%d", id)

						shardIndex := int((id % 65536) % uint64(numRanges))
						shardRanges, err := key.ParseShardingSpec(specs[shardIndex])
						require.NoError(t, err)
						require.Len(t, shardRanges, 1)
						require.True(t, key.KeyRangeContains(shardRanges[0], ksid),
							"layout=%s id=%d shard=%d ksid=%x", name, id, shardIndex, ksid)
					}
				})
			}
		})
	}
}

func legacyModuloShardSpecs(numRanges int) []string {
	maxBoundary, hexWidth := 256, 2
	if numRanges > 256 {
		maxBoundary, hexWidth = 65536, 4
	}

	specs := make([]string, numRanges)
	start := 0
	for shard := 0; shard < numRanges; shard++ {
		end := int(float64(shard+1) * (float64(maxBoundary) / float64(numRanges)))
		startLabel, endLabel := "", ""
		if start != 0 {
			startLabel = fmt.Sprintf("%0*x", hexWidth, start)
		}
		if end != maxBoundary {
			endLabel = fmt.Sprintf("%0*x", hexWidth, end)
		}
		specs[shard] = startLabel + "-" + endLabel
		start = end
	}
	return specs
}

func TestDemonwareModuloInvalidInput(t *testing.T) {
	_, err := CreateVindex("demonware_modulo", t.Name(), nil)
	require.Error(t, err)

	vindex, err := CreateVindex("demonware_modulo", t.Name(), map[string]string{"num_ranges": "2"})
	require.NoError(t, err)
	modulo := vindex.(*DemonwareModulo)
	destinations, err := modulo.Map(context.Background(), nil, []sqltypes.Value{sqltypes.NULL})
	require.NoError(t, err)
	require.Equal(t, key.DestinationNone{}, destinations[0])

	_, err = modulo.Verify(context.Background(), nil, []sqltypes.Value{sqltypes.NewVarBinary("bad")}, [][]byte{nil})
	require.Error(t, err)
}

func TestDemonwareModuloRejectsZeroDivisors(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
		want   string
	}{
		{
			name:   "zero modulo",
			params: map[string]string{"num_ranges": "2", "modulo": "0"},
			want:   "modulo",
		},
		{
			name:   "single range with offset",
			params: map[string]string{"num_ranges": "1", "use_offset": "true"},
			want:   "use_offset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateVindex("demonware_modulo", t.Name(), tt.params)
			require.ErrorContains(t, err, tt.want)
		})
	}
}

func TestDemonwareModuloVerify(t *testing.T) {
	vindex, err := CreateVindex("demonware_modulo", t.Name(), map[string]string{"num_ranges": "2", "use_offset": "true"})
	require.NoError(t, err)
	modulo := vindex.(*DemonwareModulo)
	matches, err := modulo.Verify(context.Background(), nil,
		[]sqltypes.Value{
			sqltypes.NewInt64(0), sqltypes.NewInt64(1), sqltypes.NewInt64(2),
			sqltypes.NewInt64(3), sqltypes.NewInt64(4),
		},
		[][]byte{
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0x80, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 1},
			{0x80, 0, 0, 0, 0, 0, 0, 1},
			{0, 0, 0, 0, 0, 0, 0, 4},
		})
	require.NoError(t, err)
	require.Equal(t, []bool{true, true, true, true, false}, matches)
}
