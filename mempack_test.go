package git

import (
	"bytes"
	"testing"
)

func TestMempack(t *testing.T) {
	t.Parallel()

	odb, err := NewOdb()
	checkFatal(t, err)

	repo, err := NewRepositoryWrapOdb(odb)
	checkFatal(t, err)

	mempack, err := NewMempack(odb)
	checkFatal(t, err)

	id, err := odb.Write([]byte("hello, world!"), ObjectBlob)
	checkFatal(t, err)

	expectedId, err := NewOid("30f51a3fba5274d53522d0f19748456974647b4f")
	checkFatal(t, err)
	if !expectedId.Equal(id) {
		t.Errorf("mismatched id. expected %v, got %v", expectedId.String(), id.String())
	}

	// The object should be available from the odb.
	{
		obj, err := odb.Read(expectedId)
		checkFatal(t, err)
		defer obj.Free()
	}

	data, err := mempack.Dump(repo)
	checkFatal(t, err)

	expectedData := []byte{
		0x50, 0x41, 0x43, 0x4b, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00,
		0x02, 0x9d, 0x08, 0x82, 0x3b, 0xd8, 0xa8, 0xea, 0xb5, 0x10, 0xad, 0x6a,
		0xc7, 0x5c, 0x82, 0x3c, 0xfd, 0x3e, 0xd3, 0x1e,
	}
	if !bytes.Equal(expectedData, data) {
		t.Errorf("mismatched mempack data. expected %v, got %v", expectedData, data)
	}

	mempack.Reset()

	// After the reset, the object should now be unavailable.
	{
		obj, err := odb.Read(expectedId)
		if err == nil {
			t.Errorf("object %s unexpectedly found", obj.Id().String())
			obj.Free()
		} else if !IsErrorCode(err, ErrNotFound) {
			t.Errorf("unexpected error %v", err)
		}
	}
}
