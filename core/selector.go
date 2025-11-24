// selector 选股器，自动按条件筛选优质公司。（好公司，但不代表当前股价在涨）
// 筛选规则：
// 行业要分散
// 最新 ROE 高于 8%
// ROE 平均值小于 20 时，至少 3 年内逐年递增
// EPS 至少 3 年内逐年递增
// 营业总收入至少 3 年内逐年递增
// 净利润至少 3 年内逐年递增
// 估值较低或中等
// 股价低于合理价格
// 负债率低于 60%

package core

import (
	"context"

	"github.com/axiaoxin-com/investool/datacenter"
	"github.com/axiaoxin-com/investool/datacenter/eastmoney"
)

// Selector 选股器
type Selector struct {
	Filter  eastmoney.Filter
	Checker *Checker
}

// NewSelector 创建选股器
func NewSelector(ctx context.Context, filter eastmoney.Filter, checker *Checker) Selector {
	return Selector{
		Filter:  filter,
		Checker: checker,
	}
}

// AutoFilterStocks 按默认设置自动筛选股票
func (s Selector) AutoFilterStocks(ctx context.Context) (result eastmoney.StockInfoList, err error) {
	stocks, err := datacenter.EastMoney.QuerySelectedStocksWithFilter(ctx, s.Filter)
	stocks.SortByROE()

	return stocks, err
}
