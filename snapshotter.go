package logf

// Snapshotter is the interface that allows to do a custom copy of a logging
// object. If the object type implements TaskSnapshot function it will be
// called during the logging procedure in a caller's goroutine.
type Snapshotter interface {
	TakeSnapshot() interface{}
}

// snapshotField calls an appropriate function to snapshot a Field.
func snapshotField(f *Field) {
	if f.Type&FieldTypeRawMask != 0 {
		switch f.Type {
		case FieldTypeRawBytes:
			snapshotRawBytes(f)
		}
	}
	switch f.Type {
	case FieldTypeAny, FieldTypeObject, FieldTypeArray:
		if f.Any == nil {
			return
		}
		switch rv := f.Any.(type) {
		case Snapshotter:
			f.Any = rv.TakeSnapshot()
		}
	}
}

func snapshotRawBytes(f *Field) {
	cc := make([]byte, len(f.Bytes))
	copy(cc, f.Bytes)
	f.Bytes = cc
	f.Type = FieldTypeBytes
}
