package swarm_feed

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ethersphere/bee/v2/pkg/cac"
	"github.com/ethersphere/bee/v2/pkg/crypto"
	"github.com/ethersphere/bee/v2/pkg/soc"
	"github.com/ethersphere/bee/v2/pkg/swarm"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
	"golang.org/x/crypto/sha3"
)

const (
	FEED_INDEX_HEX_LENGTH = 16
	TOPIC_HEX_LENGTH      = 64
)

type Topic string

type Index interface{}
type Identifier []byte
type IndexBytes []byte

type Feed struct {
	bClient blockstore.Client
}

func NewFeed(bClient blockstore.Client) *Feed {
	return &Feed{bClient: bClient}
}

func (f *Feed) Upload(owner, topic string, signer crypto.Signer, payload swarm.Address) (swarm.Address, error) {
	topicHash := keccak256Hash([]byte(topic))
	_, _, nextIndex, _ := f.bClient.GetLatestFeedManifest(owner, utils.Encode(topicHash))
	if nextIndex == "" {
		nextIndex = strings.Repeat("0", FEED_INDEX_HEX_LENGTH)
	}
	id, err := makeFeedIdentifier(topicHash, nextIndex)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	timestamp := numberToUint64BE(time.Now().Unix())
	payloadBytes := concatBytes(timestamp, payload.Bytes())

	ch, err := cac.New(payloadBytes)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	s := soc.New(soc.ID(id), ch)
	_, err = s.Sign(signer)
	if err != nil {
		return swarm.ZeroAddress, err
	}
	_, err = f.bClient.UploadSOC(owner, utils.Encode(id), utils.Encode(s.Signature()), ch.Data())
	if err != nil {
		return swarm.ZeroAddress, err
	}
	return f.bClient.CreateFeedManifest(owner, utils.Encode(topicHash))
}

func concatBytes(byteSlices ...[]byte) []byte {
	var buffer bytes.Buffer
	for _, b := range byteSlices {
		buffer.Write(b)
	}
	return buffer.Bytes()
}

func isEpoch(epoch interface{}) bool {
	v := reflect.ValueOf(epoch)
	if v.Kind() != reflect.Struct {
		return false
	}

	if v.FieldByName("time").IsValid() && v.FieldByName("level").IsValid() {
		return true
	}

	return false
}

func keccak256Hash(data ...[]byte) Identifier {
	hash := sha3.NewLegacyKeccak256()
	for _, d := range data {
		hash.Write(d)
	}
	return hash.Sum(nil)
}

func hexToBytes(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

func hashFeedIdentifier(topic []byte, index IndexBytes) (Identifier, error) {
	return keccak256Hash(topic, index), nil
}

func numberToUint64BE(num int64) IndexBytes {
	indexBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(indexBytes, uint64(num))
	return indexBytes
}

func makeSequentialFeedIdentifier(topic []byte, index int64) (Identifier, error) {
	indexBytes := numberToUint64BE(index)
	return hashFeedIdentifier(topic, indexBytes)
}

func makeFeedIndexBytes(s string) (IndexBytes, error) {
	hex, err := makeHexString(s, FEED_INDEX_HEX_LENGTH)
	if err != nil {
		return nil, err
	}
	return hexToBytes(hex)
}

func makeHexString(s string, length int) (string, error) {
	if len(s) > length {
		return "", errors.New("string length exceeds the required length")
	}
	return fmt.Sprintf("%0*s", length, s), nil
}

func makeFeedIdentifier(topic []byte, index Index) (Identifier, error) {
	switch idx := index.(type) {
	case int:
		return makeSequentialFeedIdentifier(topic, int64(idx))
	case string:
		indexBytes, err := makeFeedIndexBytes(idx)
		if err != nil {
			return nil, err
		}
		return hashFeedIdentifier(topic, indexBytes)
	default:
		if isEpoch(index) {
			return nil, errors.New("epoch is not yet implemented")
		}
		indexBytes, ok := index.(IndexBytes)
		if !ok {
			return nil, errors.New("invalid index type")
		}
		return hashFeedIdentifier(topic, indexBytes)
	}
}
