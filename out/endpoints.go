package endpoints

type AuthCredentialsResponse struct {
	SessionToken [16]byte
	SliceString  []string
	Uint32       uint32
	test         AuthCredentialsResponsse
}

func (AuthCredentialsResponse *AuthCredentialsResponse) Encode() {
}

func (AuthCredentialsResponse *AuthCredentialsResponse) Decode() {
}

type AuthCredentialsResponsse struct {
	SessionToken [16]byte
	SliceString  []string
	Uint32       uint32
}

func (AuthCredentialsResponsse *AuthCredentialsResponsse) Encode() {
}

func (AuthCredentialsResponsse *AuthCredentialsResponsse) Decode() {
}
