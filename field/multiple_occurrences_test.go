package field

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/moov-io/iso8583/encoding"
	"github.com/moov-io/iso8583/padding"
	"github.com/moov-io/iso8583/prefix"
	"github.com/moov-io/iso8583/sort"
	"github.com/stretchr/testify/require"
)

var (
	multipleOccurrencesFixedLenTestSpec = &Spec{
		Length:      12,
		Description: "Test Spec",
		Pref:        prefix.ASCII.Fixed,
		Pad:         padding.None,
		Tag: &TagSpec{
			Sort: sort.StringsByInt,
		},
		Subfields: map[string]Field{
			"1": NewString(&Spec{
				Length:      2,
				Description: "String Field",
				Enc:         encoding.ASCII,
				Pref:        prefix.ASCII.Fixed,
			}),
			"2": NewString(&Spec{
				Length:      2,
				Description: "String Field",
				Enc:         encoding.ASCII,
				Pref:        prefix.ASCII.Fixed,
			}),
			"3": NewNumeric(&Spec{
				Length:      2,
				Description: "Numeric Field",
				Enc:         encoding.ASCII,
				Pref:        prefix.ASCII.Fixed,
			}),
		},
	}

	multipleOccurrencesVariableLenTestSpec = &Spec{
		Length:      36,
		Description: "Test Spec",
		Pref:        prefix.ASCII.LL,
		Tag: &TagSpec{
			Sort: sort.StringsByInt,
		},
		Subfields: map[string]Field{
			"1": NewString(&Spec{
				Length:      2,
				Description: "String Field",
				Enc:         encoding.ASCII,
				Pref:        prefix.ASCII.LL,
			}),
			"2": NewString(&Spec{
				Length:      2,
				Description: "String Field",
				Enc:         encoding.ASCII,
				Pref:        prefix.ASCII.LL,
			}),
			"3": NewNumeric(&Spec{
				Length:      2,
				Description: "Numeric Field",
				Enc:         encoding.ASCII,
				Pref:        prefix.ASCII.LL,
			}),
			"11": NewComposite(&Spec{
				Length:      6,
				Description: "Sub-Composite Field",
				Pref:        prefix.ASCII.LL,
				Tag: &TagSpec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pad:    padding.Left('0'),
					Sort:   sort.StringsByInt,
				},
				Subfields: map[string]Field{
					"1": NewString(&Spec{
						Length:      2,
						Description: "String Field",
						Enc:         encoding.ASCII,
						Pref:        prefix.ASCII.LL,
					}),
				},
			}),
		},
	}
)

type MultipleOccurrencesTestData struct {
	F1  *String
	F2  *String
	F3  *Numeric
	F11 *SubMultipleOccurrencesData
}

type SubMultipleOccurrencesData struct {
	F1 *String
}

func TestMultipleOccurrences_SetData(t *testing.T) {
	t.Run("SetData returns an error on provision of primitive data type", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)
		err := composite.SetData("primitive str")
		require.EqualError(t, err, "data is not a pointer or nil")
	})

	t.Run("SetData returns an error on provision of non slice data type", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)
		err := composite.SetData(&struct{ Field string }{})
		require.EqualError(t, err, "data is not a slice")
	})

	t.Run("SetData returns an error on provision of slice with non struct element data type", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)
		err := composite.SetData(&[]string{"10"})
		require.EqualError(t, err, "element data is not a struct")
	})
}

func TestMultipleOccurrences_FieldUnmarshal(t *testing.T) {
	t.Run("Unmarshal returns an error on provision of primitive data type", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)
		err := composite.Unmarshal("primitive str")
		require.EqualError(t, err, "data is not a pointer or nil")
	})

	t.Run("Unmarshal returns an error on provision of non slice data type", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)
		err := composite.Unmarshal(&struct{ Field string }{})
		require.EqualError(t, err, "data is not a slice")
	})

	t.Run("Unmarshal returns an error on provision of slice with non struct element data type", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)
		err := composite.Unmarshal(&[]string{"10"})
		require.EqualError(t, err, "element data is not a struct")
	})

	t.Run("Unmarshal gets data for multiple occurrences field", func(t *testing.T) {
		// first, we need to populate fields of composite field
		// we will do it by packing the field
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)
		err := composite.SetData(&[]MultipleOccurrencesTestData{
			{
				F1: NewStringValue("AB"),
				F2: NewStringValue("CD"),
				F3: NewNumericValue(12),
			},
			{
				F1: NewStringValue("EF"),
				F2: NewStringValue("GH"),
				F3: NewNumericValue(14),
			},
		})
		require.NoError(t, err)

		_, err = composite.Pack()
		require.NoError(t, err)

		var data []MultipleOccurrencesTestData
		require.NoError(t, composite.Unmarshal(&data))

		require.Equal(t, "AB", data[0].F1.Value())
		require.Equal(t, "CD", data[0].F2.Value())
		require.Equal(t, 12, data[0].F3.Value())
		require.Equal(t, "EF", data[1].F1.Value())
		require.Equal(t, "GH", data[1].F2.Value())
		require.Equal(t, 14, data[1].F3.Value())
	})

	t.Run("Unmarshal gets data for multiple occurrences with sub field", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesVariableLenTestSpec)
		err := composite.SetData(&[]MultipleOccurrencesTestData{
			{
				F1: NewStringValue("7F"),
				F2: NewStringValue("2F"),
				F11: &SubMultipleOccurrencesData{
					F1: NewStringValue("4F"),
				},
			},
		})
		require.NoError(t, err)

		_, err = composite.Pack()
		require.NoError(t, err)

		var data []MultipleOccurrencesTestData
		require.NoError(t, composite.Unmarshal(&data))

		require.Equal(t, "7F", data[0].F1.Value())
		require.Equal(t, "2F", data[0].F2.Value())
		require.Equal(t, "4F", data[0].F11.F1.Value())
	})

	t.Run("Unmarshal gets data for multiple occurrences field using field tag `index`", func(t *testing.T) {
		// first, we need to populate fields of composite field
		// we will do it by packing the field
		composite := NewMultipleOccurrencesField(multipleOccurrencesVariableLenTestSpec)
		err := composite.SetData(&[]MultipleOccurrencesTestData{
			{
				F1: NewStringValue("AB"),
				F2: NewStringValue("CD"),
			},
			{
				F1: NewStringValue("EF"),
				F2: NewStringValue("GH"),
			},
		})
		require.NoError(t, err)

		_, err = composite.Pack()
		require.NoError(t, err)

		var data []struct {
			FirstCode  *String `index:"1"`
			SecondCode *String `index:"2"`
		}
		require.NoError(t, composite.Unmarshal(&data))

		require.Equal(t, "AB", data[0].FirstCode.Value())
		require.Equal(t, "CD", data[0].SecondCode.Value())
		require.Equal(t, "EF", data[1].FirstCode.Value())
		require.Equal(t, "GH", data[1].SecondCode.Value())
	})

	t.Run("Unmarshal ignores struct fields with empty index", func(t *testing.T) {
		type testObject struct {
			FirstCode  *String `index:"1"`
			SecondCode *String `index:"2"`
			ThirdCode  *String `index:""`
		}

		composite := NewMultipleOccurrencesField(multipleOccurrencesVariableLenTestSpec)
		err := composite.SetData(&[]testObject{
			{
				FirstCode:  NewStringValue("AB"),
				SecondCode: NewStringValue("CD"),
				ThirdCode:  NewStringValue("XD"),
			},
			{
				FirstCode:  NewStringValue("EF"),
				SecondCode: NewStringValue("GH"),
				ThirdCode:  NewStringValue("XD"),
			},
		})
		require.NoError(t, err)

		_, err = composite.Pack()
		require.NoError(t, err)

		var data []testObject
		require.NoError(t, composite.Unmarshal(&data))

		require.Equal(t, "AB", data[0].FirstCode.Value())
		require.Equal(t, "CD", data[0].SecondCode.Value())
		require.Nil(t, data[0].ThirdCode)
		require.Equal(t, "EF", data[1].FirstCode.Value())
		require.Equal(t, "GH", data[1].SecondCode.Value())
		require.Nil(t, data[1].ThirdCode)
	})
}

func TestMultipleOccurrences_Packing(t *testing.T) {
	t.Run("Pack returns an error on mismatch of subfield types", func(t *testing.T) {
		type TestDataIncorrectType struct {
			F1 *Numeric
		}
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)
		err := composite.SetData(&[]TestDataIncorrectType{
			{F1: NewNumericValue(1)},
		})

		require.Error(t, err)
		require.EqualError(t, err, "failed to set data from field 1: data does not match required *String type")
	})

	t.Run("Pack returns error on failure of subfield packing", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)
		err := composite.SetData(&[]MultipleOccurrencesTestData{
			{
				// This subfield will return an error on F1.Pack() as its length
				// exceeds the max length defined in the spec.
				F1: NewStringValue("ABCD"),
				F2: NewStringValue("CD"),
				F3: NewNumericValue(12),
			},
		})
		require.NoError(t, err)

		_, err = composite.Pack()
		require.EqualError(t, err, "failed to pack subfield 1: failed to encode length: field length: 4 should be fixed: 2")
	})

	t.Run("Pack returns error when encoded data length is larger than specified fixed max length", func(t *testing.T) {
		invalidSpec := &Spec{
			// Base field length < summation of lengths of subfields
			// This will throw an error when encoding the field's length.
			Length: 4,
			Pref:   prefix.ASCII.Fixed,
			Tag: &TagSpec{
				Sort: sort.StringsByInt,
			},
			Subfields: map[string]Field{
				"1": NewString(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
				"2": NewString(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
				"3": NewNumeric(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
			},
		}

		composite := NewMultipleOccurrencesField(invalidSpec)
		err := composite.Marshal(&[]MultipleOccurrencesTestData{
			{
				F1: NewStringValue("AB"),
				F2: NewStringValue("CD"),
				F3: NewNumericValue(12),
			},
		})
		require.NoError(t, err)

		_, err = composite.Pack()
		require.EqualError(t, err, "failed to encode length: field length: 6 should be fixed: 4")
	})

	t.Run("Pack returns an error when tag encoding is set to obtain TLV subfields", func(t *testing.T) {
		invalidSpec := &Spec{
			Length: 4,
			Pref:   prefix.ASCII.Fixed,
			Tag: &TagSpec{
				Length: 2,
				Enc:    encoding.ASCII,
				Sort:   sort.StringsByInt,
			},
			Subfields: map[string]Field{
				"1": NewString(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
				"2": NewString(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
				"3": NewNumeric(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
			},
		}

		composite := NewMultipleOccurrencesField(invalidSpec)

		_, err := composite.Pack()
		require.EqualError(t, err, "unsupported packing of TLV subfields")
	})

	t.Run("Pack correctly serializes data with padded tags to bytes", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)
		err := composite.SetData(&[]MultipleOccurrencesTestData{
			{
				F1: NewStringValue("AB"),
				F2: NewStringValue("CD"),
				F3: NewNumericValue(12),
			},
			{
				F1: NewStringValue("CD"),
				F2: NewStringValue("EF"),
				F3: NewNumericValue(14),
			},
		})
		require.NoError(t, err)

		packed, err := composite.Pack()
		require.NoError(t, err)

		require.Equal(t, "ABCD12CDEF14", string(packed))
	})

	t.Run("Unpack returns an error on mismatch of subfield types", func(t *testing.T) {
		type TestDataIncorrectType struct {
			F1 *Numeric
		}
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)

		_, err := composite.Unpack([]byte("ABCD12EFGH12"))
		require.NoError(t, err)

		err = composite.Unmarshal(&[]TestDataIncorrectType{})

		require.Error(t, err)
		require.EqualError(t, err, "failed to get data from field 1: data does not match required *String type")
	})

	t.Run("Unpack returns an error on failure of subfield to unpack bytes", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)

		// Last two characters must be an integer type. F3 fails to unpack.
		read, err := composite.Unpack([]byte("ABCDEF01AB50"))
		require.Equal(t, 0, read)
		require.Error(t, err)
		require.EqualError(t, err, "failed to unpack subfield 3: failed to set bytes: failed to convert into number")
		require.ErrorIs(t, err, strconv.ErrSyntax)
	})

	t.Run("Unpack returns an error when not enough data is set", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)

		// Last two characters must be an integer type. F3 fails to unpack.
		read, err := composite.Unpack([]byte("ABCD10"))
		require.Equal(t, 0, read)
		require.Error(t, err)
		require.EqualError(t, err, "not enough data to unpack, expected: 12, got: 6")
	})

	t.Run("Unpack returns an error on length of data exceeding max length", func(t *testing.T) {
		spec := &Spec{
			Length: 4,
			Pref:   prefix.ASCII.L,
			Tag: &TagSpec{
				Sort: sort.StringsByInt,
			},
			Subfields: map[string]Field{
				"1": NewString(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
				"2": NewString(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
				"3": NewNumeric(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
			},
		}

		composite := NewMultipleOccurrencesField(spec)

		// Length of denoted by prefix is too long, causing failure to decode length.
		read, err := composite.Unpack([]byte("7ABCD123"))
		require.Equal(t, 0, read)
		require.Error(t, err)
		require.EqualError(t, err, "failed to decode length: data length: 7 is larger than maximum 4")
	})

	t.Run("Unpack without error when not all subfields are set", func(t *testing.T) {
		spec := &Spec{
			Length: 4,
			Pref:   prefix.ASCII.L,
			Tag: &TagSpec{
				Sort: sort.StringsByInt,
			},
			Subfields: map[string]Field{
				"1": NewString(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
				"2": NewString(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
				"3": NewNumeric(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
			},
		}

		composite := NewMultipleOccurrencesField(spec)

		// There is data only for first subfield
		read, err := composite.Unpack([]byte("2AB"))
		require.Equal(t, 3, read)
		require.NoError(t, err)
	})

	t.Run("Unpack returns an error on offset not matching data length", func(t *testing.T) {
		invalidSpec := &Spec{
			// Base field length < summation of lengths of subfields
			// This will throw an error when encoding the field's length.
			Length: 4,
			Pref:   prefix.ASCII.Fixed,
			Tag: &TagSpec{
				Sort: sort.StringsByInt,
			},
			Subfields: map[string]Field{
				"1": NewString(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
				"2": NewString(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
				"3": NewNumeric(&Spec{
					Length: 3,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
			},
		}

		composite := NewMultipleOccurrencesField(invalidSpec)

		// Length of input too long, causing failure to decode length.
		read, err := composite.Unpack([]byte("ABCD123"))
		require.Equal(t, 0, read)
		require.Error(t, err)
		require.EqualError(t, err, "failed to unpack subfield 3: failed to decode content: not enough data to decode. expected len 3, got 0")
	})

	t.Run("Unpack returns an error when tag encoding is set to obtain TLV subfields", func(t *testing.T) {
		invalidSpec := &Spec{
			Length: 4,
			Pref:   prefix.ASCII.Fixed,
			Tag: &TagSpec{
				Length: 2,
				Enc:    encoding.ASCII,
				Sort:   sort.StringsByInt,
			},
			Subfields: map[string]Field{
				"1": NewString(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
				"2": NewString(&Spec{
					Length: 2,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
				"3": NewNumeric(&Spec{
					Length: 3,
					Enc:    encoding.ASCII,
					Pref:   prefix.ASCII.Fixed,
				}),
			},
		}

		composite := NewMultipleOccurrencesField(invalidSpec)

		// Length of input too long, causing failure to decode length.
		read, err := composite.Unpack([]byte("AB10CD123"))
		require.Equal(t, 0, read)
		require.Error(t, err)
		require.EqualError(t, err, "unsupported unpacking of TLV subfields")
	})

	t.Run("Unpack correctly deserializes bytes to the data struct", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)

		read, err := composite.Unpack([]byte("ABCD1205GH14"))
		require.Equal(t, multipleOccurrencesFixedLenTestSpec.Length, read)
		require.NoError(t, err)

		var data []MultipleOccurrencesTestData
		require.NoError(t, composite.Unmarshal(&data))

		require.Equal(t, "AB", data[0].F1.Value())
		require.Equal(t, "CD", data[0].F2.Value())
		require.Equal(t, 12, data[0].F3.Value())
		require.Nil(t, data[0].F11)
		require.Equal(t, "05", data[1].F1.Value())
		require.Equal(t, "GH", data[1].F2.Value())
		require.Equal(t, 14, data[1].F3.Value())
		require.Nil(t, data[1].F11)
	})

	t.Run("SetBytes correctly deserializes bytes to the data struct", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesFixedLenTestSpec)

		require.NoError(t, composite.SetBytes([]byte("ABCD12")))

		var data []MultipleOccurrencesTestData
		require.NoError(t, composite.Unmarshal(&data))

		require.Equal(t, "AB", data[0].F1.Value())
		require.Equal(t, "CD", data[0].F2.Value())
		require.Equal(t, 12, data[0].F3.Value())
		require.Nil(t, data[0].F11)
	})
}

func TestMultipleOccurrences_HandlesValidSpecs(t *testing.T) {
	tests := []struct {
		desc string
		spec *Spec
	}{
		{
			desc: "accepts nil Enc value",
			spec: &Spec{
				Length: 6,
				Pref:   prefix.ASCII.Fixed,
				Tag: &TagSpec{
					Sort: sort.StringsByInt,
				},
				Subfields: map[string]Field{},
			},
		},
		{
			desc: "accepts nil Pad value",
			spec: &Spec{
				Length: 6,
				Pref:   prefix.ASCII.Fixed,
				Tag: &TagSpec{
					Sort: sort.StringsByInt,
				},
				Subfields: map[string]Field{},
			},
		},
		{
			desc: "accepts None Pad value",
			spec: &Spec{
				Length: 6,
				Pref:   prefix.ASCII.Fixed,
				Pad:    padding.None,
				Tag: &TagSpec{
					Sort: sort.StringsByInt,
				},
				Subfields: map[string]Field{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("NewMultipleOccurrencesField() %v", tc.desc), func(t *testing.T) {
			f := NewMultipleOccurrencesField(tc.spec)
			require.Equal(t, tc.spec, f.Spec())
		})
		t.Run(fmt.Sprintf("MultipleOccurrencesField.SetSpec() %v", tc.desc), func(t *testing.T) {
			f := &MultipleOccurrences{}
			f.SetSpec(tc.spec)
			require.Equal(t, tc.spec, f.Spec())
		})
	}
}

func TestMultipleOccurrences_PanicsOnSpecValidationFailures(t *testing.T) {
	tests := []struct {
		desc string
		err  string
		spec *Spec
	}{
		{
			desc: "panics on nil Tag.Sort",
			err:  "Composite spec requires a Tag.Sort function to define a Tag",
			spec: &Spec{
				Length:    6,
				Pref:      prefix.ASCII.Fixed,
				Subfields: map[string]Field{},
				Tag:       &TagSpec{},
			},
		},
		{
			desc: "panics on non-None / non-nil Pad value being defined in spec",
			err:  "Composite spec only supports nil or None spec padding values",
			spec: &Spec{
				Length:    6,
				Pref:      prefix.ASCII.Fixed,
				Pad:       padding.Left('0'),
				Subfields: map[string]Field{},
				Tag: &TagSpec{
					Sort: sort.StringsByInt,
				},
			},
		},
		{
			desc: "panics on non-nil Enc value being defined in spec",
			err:  "Composite spec only supports a nil Enc value",
			spec: &Spec{
				Length:    6,
				Enc:       encoding.ASCII,
				Pref:      prefix.ASCII.Fixed,
				Subfields: map[string]Field{},
				Tag: &TagSpec{
					Sort: sort.StringsByInt,
				},
			},
		},
		{
			desc: "panics on nil Enc value being defined in spec if Tag.Length > 0",
			err:  "Composite spec requires a Tag.Enc to be defined if Tag.Length > 0",
			spec: &Spec{
				Length:    6,
				Pref:      prefix.ASCII.Fixed,
				Subfields: map[string]Field{},
				Tag: &TagSpec{
					Length: 2,
					Pad:    padding.Left('0'),
					Sort:   sort.StringsByInt,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("NewMultipleOccurrencesField() %v", tc.desc), func(t *testing.T) {
			require.PanicsWithError(t, tc.err, func() {
				NewMultipleOccurrencesField(tc.spec)
			})
		})
		t.Run(fmt.Sprintf("MultipleOccurrencesField.SetSpec() %v", tc.desc), func(t *testing.T) {
			require.PanicsWithError(t, tc.err, func() {
				(&MultipleOccurrences{}).SetSpec(tc.spec)
			})
		})
	}
}

func TestMultipleOccurrences_JSONConversion(t *testing.T) {
	json := `[{"1":"AB","2":"CD","3":12,"11":{"1":"YZ"}}]`

	t.Run("MarshalJSON typed", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesVariableLenTestSpec)
		require.NoError(t, composite.SetData(&[]MultipleOccurrencesTestData{
			{
				F1: NewStringValue("AB"),
				F2: NewStringValue("CD"),
				F3: NewNumericValue(12),
				F11: &SubMultipleOccurrencesData{
					F1: NewStringValue("YZ"),
				},
			},
		}))

		actual, err := composite.MarshalJSON()
		require.NoError(t, err)

		require.JSONEq(t, json, string(actual))
	})

	t.Run("MarshalJSON untyped", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesVariableLenTestSpec)
		require.NoError(t, composite.SetBytes([]byte("02AB02CD0212060102YZ")))

		actual, err := composite.MarshalJSON()
		require.NoError(t, err)

		require.JSONEq(t, json, string(actual))
	})

	t.Run("MarshalJSON multiple objects untyped", func(t *testing.T) {
		expectedResult := `[{"1":"AB","2":"CD","3":12,"11":{"1":"YZ"}},{"1":"AB","2":"CD","3":12,"11":{"1":"YZ"}}]`
		composite := NewMultipleOccurrencesField(multipleOccurrencesVariableLenTestSpec)
		require.NoError(t, composite.SetBytes([]byte("02AB02CD0212060102YZ02AB02CD0212060102YZ")))

		actual, err := composite.MarshalJSON()
		require.NoError(t, err)

		require.JSONEq(t, expectedResult, string(actual))
	})

	t.Run("UnmarshalJSON error when invalid body is set", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesVariableLenTestSpec)

		err := composite.UnmarshalJSON([]byte(`{"1":"AB","2":"CD","3":12,"11":{"1":"YZ"}}`))

		require.EqualError(t, err, "failed to JSON unmarshal bytes to map list")
	})

	t.Run("UnmarshalJSON error when invalid subfield is set", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesVariableLenTestSpec)

		err := composite.UnmarshalJSON([]byte(`[{"1":"AB","2":"CD","3":12,"11":"YZ"}]`))

		require.EqualError(t, err, "failed to unmarshal subfield 11")
	})

	t.Run("UnmarshalJSON typed", func(t *testing.T) {
		multipleJson := `[{"1":"AB","2":"CD","3":12,"11":{"1":"YZ"}},{"1":"EF","2":"GH","3":15}]`
		composite := NewMultipleOccurrencesField(multipleOccurrencesVariableLenTestSpec)
		require.NoError(t, composite.UnmarshalJSON([]byte(multipleJson)))

		var data []MultipleOccurrencesTestData
		require.NoError(t, composite.Unmarshal(&data))

		require.Equal(t, "AB", data[0].F1.Value())
		require.Equal(t, "CD", data[0].F2.Value())
		require.Equal(t, 12, data[0].F3.Value())
		require.Equal(t, "YZ", data[0].F11.F1.Value())
		require.Equal(t, "EF", data[1].F1.Value())
		require.Equal(t, "GH", data[1].F2.Value())
		require.Equal(t, 15, data[1].F3.Value())
		require.Nil(t, data[1].F11)
	})

	t.Run("UnmarshalJSON untyped", func(t *testing.T) {
		composite := NewMultipleOccurrencesField(multipleOccurrencesVariableLenTestSpec)

		require.NoError(t, composite.UnmarshalJSON([]byte(json)))

		s, err := composite.String()
		require.NoError(t, err)
		require.Equal(t, "02AB02CD0212060102YZ", s)
	})
}
