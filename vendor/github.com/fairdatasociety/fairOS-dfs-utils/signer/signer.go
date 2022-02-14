package signer

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethersphere/bee/pkg/swarm"
	"github.com/fairdatasociety/fairOS-dfs-utils/crypto"
	td "github.com/fairdatasociety/fairOS-dfs-utils/typedata"
)


var (
	ErrInvalidLength = errors.New("invalid signature length")
)

// EncodeForSigning encodes the hash that will be signed for the given EIP712 data
func EncodeForSigning(typedData *td.TypedData) ([]byte, error) {
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, err
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, err
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	return rawData, nil
}

// EIP712DomainType is the type description for the EIP712 Domain
var EIP712DomainType = []td.Type{
	{
		Name: "name",
		Type: "string",
	},
	{
		Name: "version",
		Type: "string",
	},
	{
		Name: "chainId",
		Type: "uint256",
	},
}


type Signer interface {
	// Sign signs data with ethereum prefix (eip191 type 0x45).
	Sign(data []byte) ([]byte, error)
	// SignTx signs an ethereum transaction.
	SignTx(transaction *types.Transaction) (*types.Transaction, error)
	SignTypedData(typedData *td.TypedData) ([]byte, error)
	// PublicKey returns the public key this signer uses.
	PublicKey() (*ecdsa.PublicKey, error)
	// EthereumAddress returns the ethereum address this signer uses.
	EthereumAddress() (common.Address, error)
}

// addEthereumPrefix adds the ethereum prefix to the data.
func addEthereumPrefix(data []byte) []byte {
	return []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data))
}

// hashWithEthereumPrefix returns the hash that should be signed for the given data.
func hashWithEthereumPrefix(data []byte) ([]byte, error) {
	return crypto.LegacyKeccak256(addEthereumPrefix(data))
}


type defaultSigner struct {
	key *ecdsa.PrivateKey
}

func NewDefaultSigner(key *ecdsa.PrivateKey) Signer {
	return &defaultSigner{
		key: key,
	}
}

// PublicKey returns the public key this signer uses.
func (d *defaultSigner) PublicKey() (*ecdsa.PublicKey, error) {
	return &d.key.PublicKey, nil
}

// Sign signs data with ethereum prefix (eip191 type 0x45).
func (d *defaultSigner) Sign(data []byte) (signature []byte, err error) {
	hash, err := hashWithEthereumPrefix(data)
	if err != nil {
		return nil, err
	}

	return d.sign(hash, true)
}

// SignTx signs an ethereum transaction.
func (d *defaultSigner) SignTx(transaction *types.Transaction) (*types.Transaction, error) {
	hash := (&types.HomesteadSigner{}).Hash(transaction).Bytes()
	// isCompressedKey is false here so we get the expected v value (27 or 28)
	signature, err := d.sign(hash, false)
	if err != nil {
		return nil, err
	}

	// v value needs to be adjusted by 27 as transaction.WithSignature expects it to be 0 or 1
	signature[64] -= 27
	return transaction.WithSignature(&types.HomesteadSigner{}, signature)
}

// EthereumAddress returns the ethereum address this signer uses.
func (d *defaultSigner) EthereumAddress() (common.Address, error) {
	publicKey, err := d.PublicKey()
	if err != nil {
		return common.Address{}, err
	}
	eth, err := crypto.NewEthereumAddress(*publicKey)
	if err != nil {
		return common.Address{}, err
	}
	var ethAddress common.Address
	copy(ethAddress[:], eth)
	return ethAddress, nil
}

func (d *defaultSigner) SignTypedData(typedData *td.TypedData) ([]byte, error) {
	rawData, err := EncodeForSigning(typedData)
	if err != nil {
		return nil, err
	}

	sighash, err := crypto.LegacyKeccak256(rawData)
	if err != nil {
		return nil, err
	}

	return d.sign(sighash, false)
}

// sign the provided hash and convert it to the ethereum (r,s,v) format.
func (d *defaultSigner) sign(sighash []byte, isCompressedKey bool) ([]byte, error) {
	signature, err := btcec.SignCompact(btcec.S256(), (*btcec.PrivateKey)(d.key), sighash, false)
	if err != nil {
		return nil, err
	}

	// Convert to Ethereum signature format with 'recovery id' v at the end.
	v := signature[0]
	copy(signature, signature[1:])
	signature[64] = v
	return signature, nil
}


// NewChunk is a convenience function to create a single-owner chunk ready to be sent
// on the network.
func NewChunk(id Id, ch swarm.Chunk, signer Signer) (swarm.Chunk, error) {
	s := NewSoc(id, ch)
	err := s.AddSigner(signer)
	if err != nil {
		return nil, err
	}
	return s.ToChunk()
}

// CreateAddress creates a new soc address from the soc id and the ethereum address of the signer
func CreateAddress(id Id, owner *Owner) (swarm.Address, error) {
	h := swarm.NewHasher()
	_, err := h.Write(id)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	_, err = h.Write(owner.address)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	sum := h.Sum(nil)
	return swarm.NewAddress(sum), nil
}

// Address returns the soc Chunk address.
func (s *Soc) Address() (swarm.Address, error) {
	return CreateAddress(s.id, s.owner)
}

func (s *Soc) toBytes() []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(s.id)
	buf.Write(s.signature)
	buf.Write(s.Chunk.Data())
	return buf.Bytes()
}

// ToChunk generates a signed chunk payload ready for submission to the swarm network.
//
// The method will fail if no signer has been added.
func (s *Soc) ToChunk() (swarm.Chunk, error) {
	var err error
	if s.signer == nil {
		return nil, errors.New("signer missing")
	}

	// generate the data to sign
	toSignBytes, err := toSignDigest(s.id, s.Chunk.Address().Bytes())
	if err != nil {
		return nil, err
	}

	// sign the chunk
	signature, err := s.signer.Sign(toSignBytes)
	if err != nil {
		return nil, err
	}
	s.signature = signature

	// create chunk
	socAddress, err := s.Address()
	if err != nil {
		return nil, err
	}
	return swarm.NewChunk(socAddress, s.toBytes()), nil
}

// NewOwner creates a new Owner.
func NewOwner(address []byte) (*Owner, error) {
	if len(address) != 20 {
		return nil, fmt.Errorf("invalid address %x", address)
	}
	return &Owner{
		address: address,
	}, nil
}

// AddSigner currently sets a single signer for the soc.
//
// This method will overwrite any value set with WithOwnerAddress with
// the address derived from the given signer.
func (s *Soc) AddSigner(signer Signer) error {
	publicKey, err := signer.PublicKey()
	if err != nil {
		return err
	}
	ownerAddressBytes, err := crypto.NewEthereumAddress(*publicKey)
	if err != nil {
		return err
	}
	ownerAddress, err := NewOwner(ownerAddressBytes)
	if err != nil {
		return err
	}
	s.signer = signer
	s.owner = ownerAddress
	return nil
}

// Id is a soc identifier
type Id []byte

func NewSoc(id Id, ch swarm.Chunk) *Soc {
	return &Soc{
		id:    id,
		Chunk: ch,
	}
}

// Owner is a wrapper that enforces valid length address of soc owner.
type Owner struct {
	address []byte
}

// Soc wraps a single soc.
type Soc struct {
	id        Id
	signature []byte
	signer    Signer
	owner     *Owner
	Chunk     swarm.Chunk
}


// toSignDigest creates a digest suitable for signing to represent the soc.
func toSignDigest(id, sum []byte) ([]byte, error) {
	h := swarm.NewHasher()
	_, err := h.Write(id)
	if err != nil {
		return nil, err
	}
	_, err = h.Write(sum)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}