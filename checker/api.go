package checker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const blockstreamRate = 0.15

var ErrRateLimited = errors.New("rate limited (429)")

type Result struct {
	BalanceSat int64
	TxCount    int64
}

var httpClient = &http.Client{Timeout: 8 * time.Second}

type provider struct {
	name    string
	limiter *rate.Limiter
	fetch   func(ctx context.Context, address string) (Result, error)
}

type Checker struct {
	p *provider
}

func NewBlockstreamChecker() *Checker {
	return &Checker{p: &provider{
		name:    "blockstream",
		limiter: rate.NewLimiter(rate.Limit(blockstreamRate), 1),
		fetch:   checkBlockstream,
	}}
}

func NewBlockchainInfoChecker(ratePerSec float64) *Checker {
	return &Checker{p: &provider{
		name:    "blockchain.info",
		limiter: rate.NewLimiter(rate.Limit(ratePerSec), 1),
		fetch:   checkBlockchainInfo,
	}}
}

func (c *Checker) Check(ctx context.Context, address string) (Result, error) {
	if err := c.p.limiter.Wait(ctx); err != nil {
		return Result{}, err
	}
	return retry(ctx, c.p, address)
}

func (c *Checker) Name() string {
	return c.p.name
}

func retry(ctx context.Context, p *provider, address string) (Result, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return Result{}, ctx.Err()
			case <-time.After(time.Duration(1<<uint(attempt-1)) * time.Second):
			}
		}
		r, err := p.fetch(ctx, address)
		if err == nil {
			return r, nil
		}
		if errors.Is(err, ErrRateLimited) {
			return Result{}, fmt.Errorf("%s: %w", p.name, err)
		}
		lastErr = err
	}
	return Result{}, fmt.Errorf("%s: %w", p.name, lastErr)
}

type esploraResponse struct {
	ChainStats struct {
		FundedTxoSum int64 `json:"funded_txo_sum"`
		SpentTxoSum  int64 `json:"spent_txo_sum"`
		TxCount      int64 `json:"tx_count"`
	} `json:"chain_stats"`
	MempoolStats struct {
		FundedTxoSum int64 `json:"funded_txo_sum"`
	} `json:"mempool_stats"`
}

func checkBlockstream(ctx context.Context, address string) (Result, error) {
	var data esploraResponse
	if err := get(ctx, "https://blockstream.info/api/address/"+address, &data); err != nil {
		return Result{}, err
	}
	return Result{
		BalanceSat: data.ChainStats.FundedTxoSum - data.ChainStats.SpentTxoSum + data.MempoolStats.FundedTxoSum,
		TxCount:    data.ChainStats.TxCount,
	}, nil
}

type blockchainInfoResponse struct {
	FinalBalance int64 `json:"final_balance"`
	NTx          int64 `json:"n_tx"`
}

func checkBlockchainInfo(ctx context.Context, address string) (Result, error) {
	var resp map[string]blockchainInfoResponse
	if err := get(ctx, "https://blockchain.info/balance?active="+address, &resp); err != nil {
		return Result{}, err
	}
	data := resp[address]
	return Result{
		BalanceSat: data.FinalBalance,
		TxCount:    data.NTx,
	}, nil
}

func get(ctx context.Context, url string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "pirates-gold/1.2")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 429 {
		return ErrRateLimited
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}
