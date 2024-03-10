package distribitor

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
)

/*@param data: any
*
* This function can take any object as input parameter and returns
* a sha1 hash of that object. If you use a struct as input parameter,
* only the public variables will be considered. */
func HashStructSha1(data interface{}) ([]byte, error) {
	var dataByteBuffer bytes.Buffer
	encoder := gob.NewEncoder(&dataByteBuffer)

	if err := encoder.Encode(data); err != nil {
		return nil, err
	}

	hashSha1 := sha1.Sum(dataByteBuffer.Bytes())

	return hashSha1[:4], nil
}

/* returns true if the byte array a i. */
func ValidateSha1Hash(a []byte, b []byte) bool {
	return bytes.Equal(a, b)
}
