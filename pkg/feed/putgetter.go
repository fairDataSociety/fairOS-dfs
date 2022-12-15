package feed

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ethersphere/bee/pkg/soc"
	"github.com/ethersphere/bee/pkg/storage"
	"github.com/ethersphere/bee/pkg/swarm"
)

// Get
func (h *Handler) Get(ctx context.Context, _ storage.ModeGet, address swarm.Address) (ch swarm.Chunk, err error) {
	chunkData, err := h.client.DownloadChunk(ctx, address.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed reading chunk body %w", err)
	}
	ch = swarm.NewChunk(address, chunkData)
	return ch, nil
}

// Put
func (h *Handler) Put(ctx context.Context, _ storage.ModePut, chs ...swarm.Chunk) (exists []bool, err error) {
	for _, ch := range chs {
		if !soc.Valid(ch) {
			return exists, errors.New("chunk not a single owner chunk")
		}

		err = h.putSOCChunk(ctx, ch)
		if err != nil {
			return exists, err
		}
	}
	return make([]bool, len(chs)), nil
}

func (h *Handler) putSOCChunk(_ context.Context, ch swarm.Chunk) error {
	chunkData := ch.Data()
	cursor := 0

	id := hex.EncodeToString(chunkData[cursor:swarm.HashSize])
	cursor += swarm.HashSize

	signature := hex.EncodeToString(chunkData[cursor : cursor+swarm.SocSignatureSize])
	cursor += swarm.SocSignatureSize

	chData := chunkData[cursor:]

	addr := h.accountInfo.GetAddress()
	_, err := h.client.UploadSOC(addr.String(), id, signature, chData)
	return err
}
