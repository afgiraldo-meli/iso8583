package field

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/moov-io/iso8583/utils"
)

// MultipleOccurrences implements Field interface to hold ISO8583 subfields that can appear
// multiple times in the same field, calling each of these appearances an occurrence.
// This is a particular logic of iso format to store different kind of values, which have a certain
// relationship, in the same point.
// For example, Currency Exponents Field, which identifies the implicit decimal point locations associated with each
// ISO standard currency code used in a message, has one occurrence for each of those distinct currency codes.
type MultipleOccurrences struct {
	spec *Spec

	orderedSpecFieldTags []string

	// stores all fields corresponding to their occurrence according to the spec.
	subfields []map[string]Field

	// tracks which subfields were set for each of the occurrences.
	setSubfields []map[string]struct{}
}

// NewMultipleOccurrencesField creates a new instance of the *MultipleOccurrences struct,
// validates and sets its Spec before returning it.
// Refer to SetSpec() for more information on Spec validation.
func NewMultipleOccurrencesField(spec *Spec) *MultipleOccurrences {
	f := &MultipleOccurrences{}
	f.SetSpec(spec)
	f.ConstructSubfields()

	return f
}

// ConstructSubfields initializes the list of fields where the decoded data will be stored. This method should be called
// before any Unpack process is executed to avoid data inconsistency due to the preexisting information.
func (c *MultipleOccurrences) ConstructSubfields() {
	c.subfields = []map[string]Field{CreateSubfields(c.spec)}
	c.setSubfields = []map[string]struct{}{make(map[string]struct{})}
}

// Spec returns the spec that was set at the initialization of MultipleOccurrences.
func (c *MultipleOccurrences) Spec() *Spec {
	return c.spec
}

// getSubfields returns the map of subfields that were set for a given occurrence index.
func (c *MultipleOccurrences) getSubfields(occurrenceIndex int) map[string]Field {
	fields := map[string]Field{}
	for i := range c.setSubfields[occurrenceIndex] {
		fields[i] = c.subfields[occurrenceIndex][i]
	}
	return fields
}

// SetSpec validates the spec and creates new instances of Subfields defined
// in the specification.
// NOTE: MultipleOccurrences does not support padding on the base spec. Therefore, users
// should only pass None or nil values for ths type. Passing any other value
// will result in a panic.
func (c *MultipleOccurrences) SetSpec(spec *Spec) {
	if err := validateCompositeSpec(spec); err != nil {
		panic(err)
	}
	c.spec = spec
	c.orderedSpecFieldTags = orderedKeys(spec.Subfields, spec.Tag.Sort)
}

// SetData Deprecated. Use Marshal instead
func (c *MultipleOccurrences) SetData(v interface{}) error {
	return c.Marshal(v)
}

// Unmarshal traverses through the stored subfields occurrences, matches them with their field in the provided data
// parameter, and calls Unmarshal(...) to set the data in the result.
//
// A valid input is as follows:
//
//	type CompositeData struct {
//	    F1 *String
//	    F2 *String
//	    F3 *Numeric
//	    F4 *SubfieldCompositeData
//	}
//	var input []CompositeData
func (c *MultipleOccurrences) Unmarshal(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("data is not a pointer or nil")
	}

	// get the slice from the pointer
	dataSlice := rv.Elem()
	if dataSlice.Kind() != reflect.Slice {
		return errors.New("data is not a slice")
	}

	for occurrenceIndex := range c.setSubfields {
		if dataSlice.Len()-1 < occurrenceIndex {
			dataSlice.Set(reflect.Append(dataSlice, reflect.New(dataSlice.Type().Elem()).Elem()))
		}

		// get the struct from the pointer
		dataStruct := dataSlice.Index(occurrenceIndex)
		if dataStruct.Kind() != reflect.Struct {
			return errors.New("element data is not a struct")
		}

		// iterate over struct fields
		for i := 0; i < dataStruct.NumField(); i++ {
			indexOrTag, _ := getFieldIndexOrTag(dataStruct.Type().Field(i))

			// skip field without index
			if indexOrTag == "" {
				continue
			}

			messageField, ok := c.subfields[occurrenceIndex][indexOrTag]
			if !ok {
				continue
			}

			// unmarshal only subfield that has the value set
			if _, set := c.setSubfields[occurrenceIndex][indexOrTag]; !set {
				continue
			}

			dataField := dataStruct.Field(i)
			if dataField.IsNil() {
				dataField.Set(reflect.New(dataField.Type().Elem()))
			}

			if err := messageField.Unmarshal(dataField.Interface()); err != nil {
				return fmt.Errorf("failed to get data from field %s: %w", indexOrTag, err)
			}
		}
	}

	return nil
}

// Marshal traverses through fields provided in the data parameter, matches them
// with their spec definition, and calls Marshal(...) on each spec field with the
// appropriate data
//
// A valid input is as follows:
//
//	type CompositeData struct {
//	    F1 *String
//	    F2 *String
//	    F3 *Numeric
//	    F4 *SubfieldCompositeData
//	}
//	input := []CompositeData{filled data}
func (c *MultipleOccurrences) Marshal(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("data is not a pointer or nil")
	}

	// get the struct from the pointer
	dataSlice := rv.Elem()
	if dataSlice.Kind() != reflect.Slice {
		return errors.New("data is not a slice")
	}

	for occurrenceIndex := 0; occurrenceIndex < dataSlice.Len(); occurrenceIndex++ {
		// get the struct from the pointer
		dataStruct := dataSlice.Index(occurrenceIndex)
		if dataStruct.Kind() != reflect.Struct {
			return errors.New("element data is not a struct")
		}

		if len(c.subfields)-1 < occurrenceIndex {
			c.addNewOccurrence()
		}

		// iterate over struct fields
		for i := 0; i < dataStruct.NumField(); i++ {
			indexOrTag, _ := getFieldIndexOrTag(dataStruct.Type().Field(i))

			// skip field without index
			if indexOrTag == "" {
				continue
			}

			messageField, ok := c.subfields[occurrenceIndex][indexOrTag]
			if !ok {
				continue
			}

			dataField := dataStruct.Field(i)
			if dataField.IsNil() {
				continue
			}

			if err := messageField.Marshal(dataField.Interface()); err != nil {
				return fmt.Errorf("failed to set data from field %s: %w", indexOrTag, err)
			}

			c.setSubfields[occurrenceIndex][indexOrTag] = struct{}{}
		}
	}

	return nil
}

// Pack deserializes data held by the receiver (via SetData) into bytes and returns an error on failure.
func (c *MultipleOccurrences) Pack() ([]byte, error) {
	packed, err := c.pack()
	if err != nil {
		return nil, err
	}

	packedLength, err := c.spec.Pref.EncodeLength(c.spec.Length, len(packed))
	if err != nil {
		return nil, fmt.Errorf("failed to encode length: %w", err)
	}

	return append(packedLength, packed...), nil
}

// Unpack takes in a byte array and serializes them into the receiver's
// subfields. An offset (unit depends on encoding and prefix values) is
// returned on success. A non-nil error is returned on failure.
func (c *MultipleOccurrences) Unpack(data []byte) (int, error) {
	dataLen, offset, err := c.spec.Pref.DecodeLength(c.spec.Length, data)
	if err != nil {
		return 0, fmt.Errorf("failed to decode length: %w", err)
	}

	isVariableLength := false
	if offset != 0 {
		isVariableLength = true
	}

	if offset+dataLen > len(data) {
		return 0, fmt.Errorf("not enough data to unpack, expected: %d, got: %d", offset+dataLen, len(data))
	}
	// data is stripped of the prefix before it is provided to unpack().
	// Therefore, it is unaware of when to stop parsing unless we bound the
	// length of the slice by the data length.
	read, err := c.unpack(data[offset:offset+dataLen], isVariableLength)
	if err != nil {
		return 0, err
	}
	if dataLen != read {
		return 0, fmt.Errorf("data length: %v does not match aggregate data read from decoded subfields: %v", dataLen, read)
	}

	return offset + read, nil
}

// SetBytes iterates over the receiver's subfields and unpacks them.
// Data passed into this method must consist of the necessary information to
// pack all subfields in full. However, unlike Unpack(), it requires the
// aggregate length of the subfields not to be encoded in the prefix.
func (c *MultipleOccurrences) SetBytes(data []byte) error {
	_, err := c.unpack(data, false)
	return err
}

// Bytes iterates over the receiver's subfields and packs them. The result
// does not incorporate the encoded aggregate length of the subfields in the
// prefix.
func (c *MultipleOccurrences) Bytes() ([]byte, error) {
	return c.pack()
}

// String iterates over the receiver's subfields, packs them and converts the
// result to a string. The result does not incorporate the encoded aggregate
// length of the subfields in the prefix.
func (c *MultipleOccurrences) String() (string, error) {
	b, err := c.Bytes()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// MarshalJSON implements the encoding/json.Marshaler interface and return the json array with the stored subfields for
// the multiple occurrences.
//
// A json result for a field with multiple occurrences looks like:
//
// [{"subfield1":"valueOccurrence1","subfield2":"valueOccurrence1"},
// {"subfield1":"valueOccurrence2","subfield2":"valueOccurrence2"}]
func (c *MultipleOccurrences) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.Write([]byte{'['})
	for occurrenceIndex := range c.setSubfields {
		subfieldsData := OrderedMap(c.getSubfields(occurrenceIndex))

		subfieldsJsonData, err := json.Marshal(subfieldsData)
		if err != nil {
			return nil, utils.NewSafeError(err, "failed to JSON marshal map to bytes")
		}

		buf.Write(subfieldsJsonData)

		if occurrenceIndex != len(c.setSubfields)-1 {
			buf.Write([]byte{','})
		}
	}
	buf.Write([]byte{']'})
	return buf.Bytes(), nil
}

// UnmarshalJSON implements the encoding/json.Unmarshaler interface.
// An error is thrown if the JSON consists of a subfield that has not
// been defined in the spec.
func (c *MultipleOccurrences) UnmarshalJSON(b []byte) error {
	var data []map[string]json.RawMessage
	err := json.Unmarshal(b, &data)
	if err != nil {
		return utils.NewSafeError(err, "failed to JSON unmarshal bytes to map list")
	}

	c.ConstructSubfields()

	for occurrenceIndex, occurrenceMap := range data {
		if len(c.subfields)-1 < occurrenceIndex {
			c.addNewOccurrence()
		}

		for tag, rawMsg := range occurrenceMap {
			if _, ok := c.spec.Subfields[tag]; !ok {
				return fmt.Errorf("failed to unmarshal subfield %v: received subfield not defined in spec", tag)
			}

			subfield, ok := c.subfields[occurrenceIndex][tag]
			if !ok {
				continue
			}

			if err := json.Unmarshal(rawMsg, subfield); err != nil {
				return utils.NewSafeErrorf(err, "failed to unmarshal subfield %v", tag)
			}

			c.setSubfields[occurrenceIndex][tag] = struct{}{}
		}
	}

	return nil
}

func (c *MultipleOccurrences) pack() ([]byte, error) {
	if c.spec.Tag != nil && c.spec.Tag.Enc != nil {
		return nil, fmt.Errorf("unsupported packing of TLV subfields")
	}
	var packed []byte
	for occurrenceIndex := range c.setSubfields {
		for _, tag := range c.orderedSpecFieldTags {
			f, ok := c.subfields[occurrenceIndex][tag]
			if !ok {
				return nil, fmt.Errorf("no subfield for tag %s", tag)
			}

			if _, set := c.setSubfields[occurrenceIndex][tag]; !set {
				continue
			}

			packedBytes, err := f.Pack()
			if err != nil {
				return nil, fmt.Errorf("failed to pack subfield %v: %w", tag, err)
			}
			packed = append(packed, packedBytes...)
		}
	}

	return packed, nil
}

func (c *MultipleOccurrences) unpack(data []byte, isVariableLength bool) (int, error) {
	if c.spec.Tag.Enc != nil {
		return 0, fmt.Errorf("unsupported unpacking of TLV subfields")
	}
	return c.unpackSubfields(data, isVariableLength)
}

func (c *MultipleOccurrences) unpackSubfields(data []byte, isVariableLength bool) (int, error) {
	c.ConstructSubfields()
	offset := 0
	occurrenceIndex := 0

	for offset < len(data) {
		for _, tag := range c.orderedSpecFieldTags {
			f, ok := c.subfields[occurrenceIndex][tag]
			if !ok {
				continue
			}

			read, err := f.Unpack(data[offset:])
			if err != nil {
				return 0, fmt.Errorf("failed to unpack subfield %v: %w", tag, err)
			}

			c.setSubfields[occurrenceIndex][tag] = struct{}{}

			offset += read

			if isVariableLength && offset >= len(data) {
				return offset, nil
			}
		}

		if offset >= len(data) {
			break
		}

		c.addNewOccurrence()
		occurrenceIndex++
	}

	return offset, nil
}

// appends a new item in the subfields list representing a new occurrence.
func (c *MultipleOccurrences) addNewOccurrence() {
	c.subfields = append(c.subfields, CreateSubfields(c.spec))
	c.setSubfields = append(c.setSubfields, make(map[string]struct{}))
}
