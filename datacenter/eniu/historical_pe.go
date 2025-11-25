// 获取历史市盈率

package eniu

import (
	"context"
	"fmt"
	"time"

	"github.com/axiaoxin-com/goutils"
	"github.com/axiaoxin-com/logging"
	"go.uber.org/zap"
)

// RespHistoricalPE 历史市盈率接口返回结构
type RespHistoricalPE struct {
	Date []string  `json:"date"`
	PE   []float64 `json:"pe_ttm"`
}

// QueryHistoricalPE 获取历史市盈率
// secuCode: 股票代码，格式为 "600000.sh" 或 "000001.sz"
// 返回数据最新数据在最后，有一天的延迟
func (e Eniu) QueryHistoricalPE(ctx context.Context, secuCode string) (RespHistoricalPE, error) {
	apiurl := fmt.Sprintf("https://eniu.com/chart/pea/%s/t/all", e.GetPathCode(ctx, secuCode))
	logging.Debug(ctx, "Eniu QueryHistoricalPE "+apiurl+" begin")
	beginTime := time.Now()
	resp := RespHistoricalPE{}
	err := goutils.HTTPGET(ctx, e.HTTPClient, apiurl, nil, &resp)
	latency := time.Since(beginTime).Milliseconds()
	logging.Debug(ctx, "Eniu QueryHistoricalPE "+apiurl+" end", zap.Int64("latency(ms)", latency), zap.Any("resp", resp))
	return resp, err
}

// GetLatestPE 获取最新的市盈率
func (p RespHistoricalPE) GetLatestPE() float64 {
	if len(p.PE) == 0 {
		return 0
	}
	return p.PE[len(p.PE)-1]
}

// GetAveragePE 获取指定时间段内的平均市盈率
// days: 天数，如果为 0 或负数则返回全部数据的平均值
func (p RespHistoricalPE) GetAveragePE(days int) float64 {
	if len(p.PE) == 0 {
		return 0
	}

	if days <= 0 {
		// 计算全部数据的平均值
		sum := 0.0
		for _, pe := range p.PE {
			sum += pe
		}
		return sum / float64(len(p.PE))
	}

	// 取最后N天的数据
	start := len(p.PE) - days
	if start < 0 {
		start = 0
	}

	sum := 0.0
	count := 0
	for i := start; i < len(p.PE); i++ {
		if p.PE[i] > 0 {
			sum += p.PE[i]
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// GetMinMaxPE 获取指定时间段内的最小和最大市盈率
// days: 天数，如果为 0 或负数则返回全部数据的最小和最大值
func (p RespHistoricalPE) GetMinMaxPE(days int) (float64, float64) {
	if len(p.PE) == 0 {
		return 0, 0
	}

	start := 0
	if days > 0 {
		start = len(p.PE) - days
		if start < 0 {
			start = 0
		}
	}

	minPE := p.PE[start]
	maxPE := p.PE[start]

	for i := start; i < len(p.PE); i++ {
		if p.PE[i] > 0 {
			if p.PE[i] < minPE {
				minPE = p.PE[i]
			}
			if p.PE[i] > maxPE {
				maxPE = p.PE[i]
			}
		}
	}

	return minPE, maxPE
}
