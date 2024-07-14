package putergetter

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ethersphere/bee/v2/pkg/soc"
	"github.com/ethersphere/bee/v2/pkg/storage"
	"github.com/ethersphere/bee/v2/pkg/swarm"

	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
)

const (
	maxIdleConnections    = 20
	maxConnectionsPerHost = 256
	requestTimeout        = 6000
)

type PutGetter struct {
	url    string
	client *http.Client
	logger logging.Logger
	batch  string
	pin    bool
}

type PutGetterWithOwner struct {
	url    string
	client *http.Client
	logger logging.Logger
	owner  string
	batch  string
	pin    bool
}

// createHTTPClient for connection re-use
func createHTTPClient() *http.Client {
	client := &http.Client{
		Timeout: time.Second * requestTimeout,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: maxIdleConnections,
			MaxConnsPerHost:     maxConnectionsPerHost,
		},
	}
	return client
}

func NewPutGetter(url, batch string, pin bool, client *http.Client, logger logging.Logger) *PutGetter {
	return &PutGetter{
		url:    url,
		client: client,
		logger: logger,
		batch:  batch,
		pin:    pin,
	}
}

func (p *PutGetter) NewPutGetterWithOwner(owner string) *PutGetterWithOwner {
	return &PutGetterWithOwner{
		url:    p.url,
		client: p.client,
		logger: p.logger,
		batch:  p.batch,
		owner:  owner,
		pin:    p.pin,
	}
}

func (p *PutGetterWithOwner) Get(ctx context.Context, address swarm.Address) (ch swarm.Chunk, err error) {
	addressHex := address.String()
	fullUrl := strings.Join([]string{p.url, "chunks", addressHex}, "/")
	req, err := http.NewRequestWithContext(ctx, "GET", fullUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating http req %w", err)
	}
	res, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed executing http req %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chunk %s not found %w", addressHex, storage.ErrNotFound)
	}
	chunkData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading chunk body %w", err)
	}
	ch = swarm.NewChunk(address, chunkData)
	return ch, nil
}

func (p *PutGetterWithOwner) Put(ctx context.Context, chs ...swarm.Chunk) (exists []bool, err error) {
	for _, ch := range chs {
		if !soc.Valid(ch) {
			return exists, errors.New("chunk not a single owner chunk")
		}

		err = p.putSOCChunk(ctx, ch)
		if err != nil {
			return exists, err
		}
	}
	return make([]bool, len(chs)), nil
}

func (p *PutGetterWithOwner) putSOCChunk(ctx context.Context, ch swarm.Chunk) error {
	chunkData := ch.Data()
	cursor := 0

	id := hex.EncodeToString(chunkData[cursor:swarm.HashSize])
	cursor += swarm.HashSize

	signature := hex.EncodeToString(chunkData[cursor : cursor+swarm.SocSignatureSize])
	cursor += swarm.SocSignatureSize

	chData := chunkData[cursor:]

	qURL, err := url.Parse(strings.Join([]string{p.url, "soc", p.owner, id}, "/"))
	if err != nil {
		return fmt.Errorf("failed parsing URL %w", err)
	}

	q := qURL.Query()
	q.Set("sig", signature)
	qURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "POST", qURL.String(), bytes.NewBuffer(chData))
	if err != nil {
		return fmt.Errorf("failed creating HTTP req %w", err)
	}
	req.Header.Set("Swarm-Postage-Batch-Id", p.batch)
	if p.pin {
		req.Header.Set("Swarm-Pin", "true")
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed executing HTTP req %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("invalid status code from response %d", resp.StatusCode)
	}

	return nil
}
