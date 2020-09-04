// Package encodingtest provides a test suite for testing codec implementations.
package encodingtest

import (
	"bytes"
	"testing"
	"time"

	"github.com/genjidb/genji/document"
	"github.com/genjidb/genji/document/encoding"
	"github.com/stretchr/testify/require"
)

// TestCodec runs a list of tests on the given codec.
func TestCodec(t *testing.T, codecBuilder func() encoding.Codec) {
	tests := []struct {
		name string
		test func(*testing.T, func() encoding.Codec)
	}{
		{"EncodeDecode", testEncodeDecode},
		{"NewDocument", testDecodeDocument},
		{"Array/GetByIndex", testArrayGetByIndex},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.test(t, codecBuilder)
		})
	}
}

func testEncodeDecode(t *testing.T, codecBuilder func() encoding.Codec) {
	userMapDoc, err := document.NewFromMap(map[string]interface{}{
		"age":  10,
		"name": "john",
	})
	require.NoError(t, err)

	addressMapDoc, err := document.NewFromMap(map[string]string{
		"city":    "Ajaccio",
		"country": "France",
	})
	require.NoError(t, err)

	complexArray := document.NewValueBuffer().
		Append(document.NewIntegerValue(-40)).
		Append(document.NewBoolValue(true)).
		Append(document.NewTextValue("hello")).
		Append(document.NewDocumentValue(addressMapDoc)).
		Append(document.NewArrayValue(document.NewValueBuffer().Append(document.NewIntegerValue(11))))

	tests := []struct {
		name     string
		d        document.Document
		expected string
	}{
		{
			"document.FieldBuffer",
			document.NewFieldBuffer().
				Add("age", document.NewIntegerValue(10)).
				Add("name", document.NewTextValue("john")),
			`{"age": 10, "name": "john"}`,
		},
		{
			"Map",
			userMapDoc,
			`{"age": 10, "name": "john"}`,
		},
		{
			"Nested Document",
			document.NewFieldBuffer().
				Add("age", document.NewIntegerValue(10)).
				Add("name", document.NewTextValue("john")).
				Add("address", document.NewDocumentValue(addressMapDoc)).
				Add("array", document.NewArrayValue(complexArray)),
			`{"age": 10, "name": "john", "address": {"city": "Ajaccio", "country": "France"}, "array": [-40, true, "hello", {"city": "Ajaccio", "country": "France"}, [11]]}`,
		},
	}

	var buf bytes.Buffer
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()
			codec := codecBuilder()
			err := codec.NewEncoder(&buf).EncodeDocument(test.d)
			require.NoError(t, err)
			var jsonBuf bytes.Buffer
			err = document.ToJSON(&jsonBuf, codec.NewDocument(buf.Bytes()))
			require.NoError(t, err)
			require.JSONEq(t, test.expected, jsonBuf.String())
		})
	}
}

func testArrayGetByIndex(t *testing.T, codecBuilder func() encoding.Codec) {
	codec := codecBuilder()

	arr := document.NewValueBuffer().
		Append(document.NewIntegerValue(10)).
		Append(document.NewNullValue()).
		Append(document.NewTextValue("john")).
		Append(document.NewDurationValue(10 * time.Nanosecond))

	var buf bytes.Buffer

	err := codec.NewEncoder(&buf).EncodeDocument(document.NewFieldBuffer().Add("a", document.NewArrayValue(arr)))
	require.NoError(t, err)

	d := codec.NewDocument(buf.Bytes())
	v, err := d.GetByField("a")
	require.NoError(t, err)

	require.Equal(t, document.ArrayValue, v.Type)
	a := v.V.(document.Array)
	v, err = a.GetByIndex(0)
	require.NoError(t, err)

	require.Equal(t, document.NewIntegerValue(10), v)

	v, err = a.GetByIndex(1)
	require.NoError(t, err)
	require.Equal(t, document.NewNullValue(), v)

	v, err = a.GetByIndex(2)
	require.NoError(t, err)
	require.Equal(t, document.NewTextValue("john"), v)

	v, err = a.GetByIndex(1000)
	require.Equal(t, err, document.ErrValueNotFound)
}

func testDecodeDocument(t *testing.T, codecBuilder func() encoding.Codec) {
	codec := codecBuilder()

	mapDoc, err := document.NewFromMap(map[string]string{
		"city":    "Ajaccio",
		"country": "France",
	})
	require.NoError(t, err)

	doc := document.NewFieldBuffer().
		Add("age", document.NewIntegerValue(10)).
		Add("name", document.NewTextValue("john")).
		Add("address", document.NewDocumentValue(mapDoc))

	var buf bytes.Buffer

	err = codec.NewEncoder(&buf).EncodeDocument(doc)
	require.NoError(t, err)

	ec := codec.NewDocument(buf.Bytes())
	v, err := ec.GetByField("age")
	require.NoError(t, err)
	require.Equal(t, document.NewIntegerValue(10), v)
	v, err = ec.GetByField("address")
	require.NoError(t, err)
	var expected, actual bytes.Buffer
	err = document.ToJSON(&expected, document.NewFieldBuffer().Add("address", document.NewDocumentValue(mapDoc)))
	require.NoError(t, err)
	err = document.ToJSON(&actual, document.NewFieldBuffer().Add("address", v))
	require.NoError(t, err)
	require.JSONEq(t, expected.String(), actual.String())

	var i int
	err = ec.Iterate(func(f string, v document.Value) error {
		switch f {
		case "age":
			require.Equal(t, document.NewIntegerValue(10), v)
		case "address":
			var expected, actual bytes.Buffer
			err = document.ToJSON(&expected, document.NewFieldBuffer().Add("address", document.NewDocumentValue(mapDoc)))
			require.NoError(t, err)
			err = document.ToJSON(&actual, document.NewFieldBuffer().Add(f, v))
			require.NoError(t, err)
			require.JSONEq(t, expected.String(), actual.String())
		case "name":
			require.Equal(t, document.NewTextValue("john"), v)
		}
		i++
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 3, i)
}
