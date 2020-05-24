package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	Publickey  []byte
}

//checksum is important for veryfying the transaction
//version indicates where in the blockchain the transaction resides

const (
	checksumLength = 4
	version        = byte(0x00)
)

//step3 ==> public key hash + checksum+version + base58==>address
func (w Wallet) Address() []byte {

	pubHash := PublicKeyHash(w.Publickey)
	versionedHash := append([]byte{version}, pubHash...)
	checksum := CheckSum(versionedHash)
	fullHash := append(versionedHash, checksum...)
	// fmt.Printf("fullHash %x\n", fullHash)
	address := Base58Encode(fullHash)
	// fmt.Printf("pubKey %x\n", w.Publickey)
	// fmt.Printf("pub Hash %x \n", pubHash)
	// fmt.Printf("address %x \n", address)

	return address

}
func ValidateAddress(address string) bool {
	pubkeyhash := Base58Decode([]byte(address))
	actualCheckSum := pubkeyhash[len(pubkeyhash)-checksumLength:]
	version := pubkeyhash[0]
	pubkeyhash = pubkeyhash[1 : len(pubkeyhash)-checksumLength]
	targetCheckSum := CheckSum(append([]byte{version}, pubkeyhash...))
	return bytes.Compare(actualCheckSum, targetCheckSum) == 0
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pub
}
func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}

//step 1 public key +sha256+ripemd160 => public key hash
func PublicKeyHash(pubKey []byte) []byte {
	pubhash := sha256.Sum256(pubKey)
	hasher := ripemd160.New()
	_, err := hasher.Write(pubhash[:])
	if err != nil {
		log.Panic(err)
	}
	publicRipMD := hasher.Sum(nil)
	return publicRipMD
}

//step 2 publicKeyHash=>sha256+sha256 + return first 4 bytes of the result =>checksum
func CheckSum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:checksumLength]
}
