package endpoints

import "unsafe"

type AuthCredentialsResponse struct {
	SessionToken [16]byte
	OneByte      byte
	Uint8        uint8
	Int8         int8
	ByteSlice    []byte
	Uint32       uint32
}

func (AuthCredentialsResponse *AuthCredentialsResponse) Encode() []byte {
	var res []byte

	res = append(res, AuthCredentialsResponse.SessionToken[:]...)
	res[16] = AuthCredentialsResponse.OneByte
	res[17] = AuthCredentialsResponse.Uint8
	res[18] = byte(AuthCredentialsResponse.Int8)
	sliceByteSliceLenght := len(AuthCredentialsResponse.ByteSlice)

	a := (*[4]byte)(unsafe.Pointer(&AuthCredentialsResponse.Uint32))[:]
	res = append(res, a...)

	return res
}

func DecodeAuthCredentialsResponse(data []byte, out *AuthCredentialsResponse) {
}
