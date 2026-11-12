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

func TestDemonwareModuloVerify(t *testing.T) {
	vindex, err := CreateVindex("demonware_modulo", t.Name(), map[string]string{"num_ranges": "2", "use_offset": "true"})
	require.NoError(t, err)
	modulo := vindex.(*DemonwareModulo)
	matches, err := modulo.Verify(context.Background(), nil,
		[]sqltypes.Value{sqltypes.NewInt64(2), sqltypes.NewInt64(3)},
		[][]byte{{0, 0, 0, 0, 0, 0, 0, 1}, {0x80, 0, 0, 0, 0, 0, 0, 4}})
	require.NoError(t, err)
	require.Equal(t, []bool{true, false}, matches)
}
