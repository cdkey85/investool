// 按我的选股指标获取股票数据，对优质公司进行初步筛选（好公司不代表股价涨）

package eastmoney

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/axiaoxin-com/goutils"
	"github.com/axiaoxin-com/logging"
	"go.uber.org/zap"
)

// Filter 我的选股指标
// 注意：指标数据取决于最新一期的财报数据，可能是年报也可能是其他季报，使用时需注意应该同期对比
type Filter struct {
	// ------ 最重要的指标！！！------
	// 最低净资产收益率（%）， ROE_WEIGHT
	MinROE float64 `json:"min_roe" form:"selector_min_roe"`

	// ------ 必要参数 ------
	// 最低净利润增长率（%） ， NETPROFIT_YOY_RATIO
	MinNetprofitYoyRatio float64 `json:"min_netprofit_yoy_ratio"      form:"selector_min_netprofit_yoy_ratio"`
	// 最低营收增长率（%） ， TOI_YOY_RATIO
	MinToiYoyRatio float64 `json:"min_toi_yoy_ratio"            form:"selector_min_toi_yoy_ratio"`
	// 最低最新股息率（%）， ZXGXL
	MinZXGXL float64 `json:"min_zxgxl"                    form:"selector_min_zxgxl"`
	// 最低净利润 3 年复合增长率（%）， NETPROFIT_GROWTHRATE_3Y
	MinNetprofitGrowthrate3Y float64 `json:"min_netprofit_growthrate_3_y" form:"selector_min_netprofit_growthrate_3_y"`
	// 最低营收 3 年复合增长率（%）， INCOME_GROWTHRATE_3Y
	MinIncomeGrowthrate3Y float64 `json:"min_income_growthrate_3_y"    form:"selector_min_income_growthrate_3_y"`
	// 最低上市以来年化收益率（%） ， LISTING_YIELD_YEAR
	MinListingYieldYear float64 `json:"min_listing_yield_year"       form:"selector_min_listing_yield_year"`
	// 最低市净率， PBNEWMRQ
	MinPBNewMRQ float64 `json:"min_pb_new_mrq"               form:"selector_min_pb_new_mrq"`
	// 最高滚动市盈率， PE9
	MaxPE9 float64 `json:"max_pe9"                       form:"selector_max_pe9"`

	// ------ 可选参数 ------
	// 最大资产负债率 (%)
	MaxDebtAssetRatio float64 `json:"max_debt_asset_ratio"         form:"selector_max_debt_asset_ratio"`
	// 最低预测净利润同比增长（%）， PREDICT_NETPROFIT_RATIO
	MinPredictNetprofitRatio float64 `json:"min_predict_netprofit_ratio"     form:"selector_min_predict_netprofit_ratio"`
	// 最低预测营收同比增长（%）， PREDICT_INCOME_RATIO
	MinPredictIncomeRatio float64 `json:"min_predict_income_ratio"        form:"selector_min_predict_income_ratio"`
	// 最低总市值（亿）， TOTAL_MARKET_CAP
	MinTotalMarketCap float64 `json:"min_total_market_cap"            form:"selector_min_total_market_cap"`
	// 行业名（可选参数，不设置搜全行业）， INDUSTRY
	IndustryList []string `json:"industry_list"                   form:"selector_industry_list"`
	// 股价范围最小值（元）， NEW_PRICE
	MinPrice float64 `json:"min_price"                       form:"selector_min_price"`
	// 股价范围最大值（元）， NEW_PRICE
	MaxPrice float64 `json:"max_price"                       form:"selector_max_price"`
	// 上市时间是否超过 5 年，@LISTING_DATE="OVER5Y"
	ListingOver5Y bool `json:"listing_over_5_y"                form:"selector_listing_over_5_y"`
	// 最低上市以来年化波动率， LISTING_VOLATILITY_YEAR
	MinListingVolatilityYear float64 `json:"min_listing_volatility_year"     form:"selector_min_listing_volatility_year"`
	// 是否排除创业板 300XXX
	ExcludeCYB bool `json:"exclude_cyb"                     form:"selector_exclude_cyb"`
	// 是否排除科创板 688XXX
	ExcludeKCB bool `json:"exclude_kcb"                     form:"selector_exclude_kcb"`
	// 查询指定名称
	SpecialSecurityNameAbbrList []string `json:"special_security_name_abbr_list" form:"selector_special_security_name_abbr_list"`
	// 查询指定代码
	SpecialSecurityCodeList []string `json:"special_security_code_list"      form:"selector_special_security_code_list"`
	// 最小总资产收益率 ROA
	MinROA float64 `json:"min_roa"                         form:"selector_min_roa"`
}

// String 转为字符串的请求参数
func (f Filter) String() string {
	filter := ""
	// 特定查询
	if len(f.SpecialSecurityNameAbbrList) > 0 {
		names := []string{}
		for _, name := range f.SpecialSecurityNameAbbrList {
			names = append(names, fmt.Sprintf(`"%s"`, name))
		}
		filter += fmt.Sprintf(`(SECURITY_NAME_ABBR in (%s))`, strings.Join(names, ","))
		return filter
	}
	if len(f.SpecialSecurityCodeList) > 0 {
		codes := []string{}
		for _, code := range f.SpecialSecurityCodeList {
			codes = append(codes, fmt.Sprintf(`"%s"`, code))
		}
		filter += fmt.Sprintf(`(SECURITY_CODE in (%s))`, strings.Join(codes, ","))
		return filter
	}
	// 必要参数
	filter += fmt.Sprintf(`(ROE_WEIGHT>=%f)`, f.MinROE)
	filter += fmt.Sprintf(`(NETPROFIT_YOY_RATIO>=%f)`, f.MinNetprofitYoyRatio)
	filter += fmt.Sprintf(`(TOI_YOY_RATIO>=%f)`, f.MinToiYoyRatio)
	filter += fmt.Sprintf(`(ZXGXL>=%f)`, f.MinZXGXL)
	filter += fmt.Sprintf(`(NETPROFIT_GROWTHRATE_3Y>=%f)`, f.MinNetprofitGrowthrate3Y)
	filter += fmt.Sprintf(`(INCOME_GROWTHRATE_3Y>=%f)`, f.MinIncomeGrowthrate3Y)
	filter += fmt.Sprintf(`(LISTING_YIELD_YEAR>=%f)`, f.MinListingYieldYear)
	filter += fmt.Sprintf(`(PBNEWMRQ>=%f)`, f.MinPBNewMRQ)
	filter += fmt.Sprintf(`(PE9<=%f)`, f.MaxPE9)

	// 可选参数
	if f.MaxDebtAssetRatio != 0 {
		filter += fmt.Sprintf(`(DEBT_ASSET_RATIO<=%f)`, f.MaxDebtAssetRatio)
	}
	if f.MinPredictNetprofitRatio != 0 {
		filter += fmt.Sprintf(`(PREDICT_NETPROFIT_RATIO>=%f)`, f.MinPredictNetprofitRatio)
	}
	if f.MinPredictIncomeRatio != 0 {
		filter += fmt.Sprintf(`(PREDICT_INCOME_RATIO>=%f)`, f.MinPredictIncomeRatio)
	}
	if f.MinTotalMarketCap != 0 {
		filter += fmt.Sprintf(`(TOTAL_MARKET_CAP>=%f)`, f.MinTotalMarketCap*100000000)
	}
	if len(f.IndustryList) != 0 {
		industryIn := []string{}
		for _, i := range f.IndustryList {
			industryIn = append(industryIn, fmt.Sprintf(`"%s"`, i))
		}
		filter += fmt.Sprintf(`(INDUSTRY in (%s))`, strings.Join(industryIn, ","))
	}
	if f.MinPrice != 0 {
		filter += fmt.Sprintf(`(NEW_PRICE>=%f))`, f.MinPrice)
	}
	if f.MaxPrice != 0 {
		filter += fmt.Sprintf(`(NEW_PRICE<=%f))`, f.MaxPrice)
	}
	if f.ListingOver5Y {
		filter += `(@LISTING_DATE="OVER5Y")`
	}
	if f.MinListingVolatilityYear != 0 {
		filter += fmt.Sprintf(`(LISTING_VOLATILITY_YEAR>=%f))`, f.MinListingVolatilityYear)
	}
	if f.MinROA != 0 {
		filter += fmt.Sprintf(`(JROA>=%f)`, f.MinROA)
	}
	return filter
}

var (
	// DefaultFilter 默认指标值
	DefaultFilter = Filter{
		// 存在年报 ROE 大于 8，但季报小于 8 的情况会筛选不出来
		MinROE:            8.0,
		MinTotalMarketCap: 100.0,
		MinPBNewMRQ:       1.0,
		ExcludeCYB:        true,
		ExcludeKCB:        true,
	}
)

// StockInfo 接口返回的股票信息结构
type StockInfo struct {
	// 股票代码：带后缀
	Secucode string `json:"SECUCODE"`
	// 股票代码：无后缀
	SecurityCode string `json:"SECURITY_CODE"`
	// 股票名
	SecurityNameAbbr string `json:"SECURITY_NAME_ABBR"`
	// 行业
	Industry string `json:"INDUSTRY"`
	// 最新一期 ROE
	RoeWeight float64 `json:"ROE_WEIGHT"`
	// 净利润增长率（%）
	NetprofitYoyRatio float64 `json:"NETPROFIT_YOY_RATIO"`
	// 营收增长率（%）
	ToiYoyRatio float64 `json:"TOI_YOY_RATIO"`
	// 最新股息率 (%)
	Zxgxl float64 `json:"ZXGXL"`
	// 净利润 3 年复合增长率
	NetprofitGrowthrate3Y float64 `json:"NETPROFIT_GROWTHRATE_3Y"`
	// 营收 3 年复合增长率
	IncomeGrowthrate3Y float64 `json:"INCOME_GROWTHRATE_3Y"`
	// 上市以来年化收益率
	ListingYieldYear float64 `json:"LISTING_YIELD_YEAR"`
	// 市净率
	PBNewMRQ float64 `json:"PBNEWMRQ"`
	// 预测净利润同比增长
	PredictNetprofitRatio float64 `json:"PREDICT_NETPROFIT_RATIO"`
	// 预测营收同比增长
	PredictIncomeRatio float64 `json:"PREDICT_INCOME_RATIO"`
	// 总市值
	TotalMarketCap float64 `json:"TOTAL_MARKET_CAP"`
	// 最新价（元） 未开盘价格显示为 -
	NewPrice interface{} `json:"NEW_PRICE"`
	// 上市以来年化波动率
	ListingVolatilityYear float64 `json:"LISTING_VOLATILITY_YEAR"`
	// 上市时间
	ListingDate string `json:"LISTING_DATE"`
	// 资产负债率
	DebtAssetRatio float64 `json:"DEBT_ASSET_RATIO"`
	// ROA (%)
	ROA float64 `json:"JROA"`
	// 滚动市盈率
	PE float64 `json:"PE9"`
	// 动态市盈率
	DTSYL float64 `json:"DTSYL"`
}

// StockInfoList 股票列表
type StockInfoList []StockInfo

// SortByROE 股票列表按 ROE 排序
func (s StockInfoList) SortByROE() {
	sort.Slice(s, func(i, j int) bool {
		return s[i].RoeWeight > s[j].RoeWeight
	})
}

// RespSelectStocks 接口返回 json 结构
type RespSelectStocks struct {
	Result struct {
		Nextpage    bool          `json:"nextpage"`
		Currentpage int           `json:"currentpage"`
		Data        StockInfoList `json:"data"`
		Config      []struct {
			IndicatorName string `json:"INDICATOR_NAME"`
			Datatype      string `json:"DATATYPE"`
		} `json:"config"`
	} `json:"result"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// QuerySelectedStocks 按选股指标默认值筛选股票
func (e EastMoney) QuerySelectedStocks(ctx context.Context) (StockInfoList, error) {
	return e.QuerySelectedStocksWithFilter(ctx, DefaultFilter)
}

// QuerySelectedStocksWithFilter 自定义选股指标值筛选股票
func (e EastMoney) QuerySelectedStocksWithFilter(ctx context.Context, filter Filter) (StockInfoList, error) {
	apiurl := "http://datacenter.eastmoney.com/stock/selection/api/data/get/"
	reqData := map[string]string{
		"source": "SELECT_SECURITIES",
		"client": "APP",
		"type":   "RPTA_APP_STOCKSELECT",
		"sty":    "SECUCODE,SECURITY_CODE,SECURITY_NAME_ABBR,INDUSTRY,ROE_WEIGHT,NETPROFIT_YOY_RATIO,TOI_YOY_RATIO,ZXGXL,NETPROFIT_GROWTHRATE_3Y,INCOME_GROWTHRATE_3Y,LISTING_YIELD_YEAR,PBNEWMRQ,PREDICT_NETPROFIT_RATIO,PREDICT_INCOME_RATIO,TOTAL_MARKET_CAP,NEW_PRICE,LISTING_VOLATILITY_YEAR,LISTING_DATE,DEBT_ASSET_RATIO,JROA,PE9,DTSYL",
		"filter": filter.String(),
		"p":      "1",      // page
		"ps":     "100000", // page size
	}
	logging.Info(ctx, "EastMoney QuerySelectedStocksWithFilter "+apiurl+" begin", zap.Any("reqData", reqData))
	beginTime := time.Now()
	req, err := goutils.NewHTTPMultipartReq(ctx, apiurl, reqData)
	if err != nil {
		return nil, err
	}
	resp := RespSelectStocks{}
	err = goutils.HTTPPOST(ctx, e.HTTPClient, req, &resp)
	latency := time.Now().Sub(beginTime).Milliseconds()
	logging.Info(
		ctx,
		"EastMoney SelectStocksWithFilter "+apiurl+" end",
		zap.Int64("latency(ms)", latency),
		// zap.Any("resp", resp),
	)
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("%s %#v", filter.String(), resp)
	}
	result := StockInfoList{}
	for _, i := range resp.Result.Data {
		// 排除创业板
		if filter.ExcludeCYB && strings.HasPrefix(i.Secucode, "300") {
			logging.Debugf(ctx, "EastMoney SelectStocksWithFilter ExcludeCYB %s %s", i.SecurityNameAbbr, i.Secucode)
			continue
		}
		// 排除科创板
		if filter.ExcludeKCB && strings.HasPrefix(i.Secucode, "688") {
			logging.Debugf(ctx, "EastMoney SelectStocksWithFilter ExcludeKCB %s %s", i.SecurityNameAbbr, i.Secucode)
			continue
		}
		result = append(result, i)
	}
	return result, nil
}

// CNStockSelection 综合选股结构体，对应 TABLE_CN_STOCK_SELECTION 表
type CNStockSelection struct {
	Date                      string  `json:"date" map:"MAX_TRADE_DATE"`                                     // 日期
	Code                      string  `json:"code" map:"SECURITY_CODE"`                                      // 代码
	Name                      string  `json:"name" map:"SECURITY_NAME_ABBR"`                                 // 名称
	NewPrice                  float64 `json:"new_price" map:"NEW_PRICE"`                                     // 最新价
	ChangeRate                float64 `json:"change_rate" map:"CHANGE_RATE"`                                 // 涨跌幅
	VolumeRatio               float64 `json:"volume_ratio" map:"VOLUME_RATIO"`                               // 量比
	HighPrice                 float64 `json:"high_price" map:"HIGH_PRICE"`                                   // 最高价
	LowPrice                  float64 `json:"low_price" map:"LOW_PRICE"`                                     // 最低价
	PreClosePrice             float64 `json:"pre_close_price" map:"PRE_CLOSE_PRICE"`                         // 昨收价
	Volume                    int64   `json:"volume" map:"VOLUME"`                                           // 成交量
	DealAmount                int64   `json:"deal_amount" map:"DEAL_AMOUNT"`                                 // 成交额
	Turnoverrate              float64 `json:"turnoverrate" map:"TURNOVERRATE"`                               // 换手率
	ListingDate               string  `json:"listing_date" map:"LISTING_DATE"`                               // 上市时间
	Industry                  string  `json:"industry" map:"INDUSTRY"`                                       // 行业
	Area                      string  `json:"area" map:"AREA"`                                               // 地区
	Concept                   string  `json:"concept" map:"CONCEPT"`                                         // 概念
	Style                     string  `json:"style" map:"STYLE"`                                             // 板块
	IsHs300                   string  `json:"is_hs300" map:"IS_HS300"`                                       // 沪300
	IsSz50                    string  `json:"is_sz50" map:"IS_SZ50"`                                         // 上证50
	IsZz500                   string  `json:"is_zz500" map:"IS_ZZ500"`                                       // 中证500
	IsZz1000                  string  `json:"is_zz1000" map:"IS_ZZ1000"`                                     // 中证1000
	IsCy50                    string  `json:"is_cy50" map:"IS_CY50"`                                         // 创业板50
	Pe9                       float64 `json:"pe9" map:"PE9"`                                                 // 市盈率TTM
	Pbnewmrq                  float64 `json:"pbnewmrq" map:"PBNEWMRQ"`                                       // 市净率MRQ
	Pettmdeducted             float64 `json:"pettmdeducted" map:"PETTMDEDUCTED"`                             // 市盈率TTM扣非
	Ps9                       float64 `json:"ps9" map:"PS9"`                                                 // 市销率TTM
	Pcfjyxjl9                 float64 `json:"pcfjyxjl9" map:"PCFJYXJL9"`                                     // 市现率TTM
	PredictPeSyear            float64 `json:"predict_pe_syear" map:"PREDICT_PE_SYEAR"`                       // 预测市盈率今年
	PredictPeNyear            float64 `json:"predict_pe_nyear" map:"PREDICT_PE_NYEAR"`                       // 预测市盈率明年
	TotalMarketCap            int64   `json:"total_market_cap" map:"TOTAL_MARKET_CAP"`                       // 总市值
	FreeCap                   int64   `json:"free_cap" map:"FREE_CAP"`                                       // 流通市值
	Dtsyl                     float64 `json:"dtsyl" map:"DTSYL"`                                             // 动态市盈率
	Ycpeg                     float64 `json:"ycpeg" map:"YCPEG"`                                             // 预测PEG
	EnterpriseValueMultiple   float64 `json:"enterprise_value_multiple" map:"ENTERPRISE_VALUE_MULTIPLE"`     // 企业价值倍数
	BasicEps                  float64 `json:"basic_eps" map:"BASIC_EPS"`                                     // 每股收益
	Bvps                      float64 `json:"bvps" map:"BVPS"`                                               // 每股净资产
	PerNetcashOperate         float64 `json:"per_netcash_operate" map:"PER_NETCASH_OPERATE"`                 // 每股经营现金流
	PerFcfe                   float64 `json:"per_fcfe" map:"PER_FCFE"`                                       // 每股自由现金流
	PerCapitalReserve         float64 `json:"per_capital_reserve" map:"PER_CAPITAL_RESERVE"`                 // 每股资本公积
	PerUnassignProfit         float64 `json:"per_unassign_profit" map:"PER_UNASSIGN_PROFIT"`                 // 每股未分配利润
	PerSurplusReserve         float64 `json:"per_surplus_reserve" map:"PER_SURPLUS_RESERVE"`                 // 每股盈余公积
	PerRetainedEarning        float64 `json:"per_retained_earning" map:"PER_RETAINED_EARNING"`               // 每股留存收益
	ParentNetprofit           int64   `json:"parent_netprofit" map:"PARENT_NETPROFIT"`                       // 归属净利润
	DeductNetprofit           int64   `json:"deduct_netprofit" map:"DEDUCT_NETPROFIT"`                       // 扣非净利润
	TotalOperateIncome        int64   `json:"total_operate_income" map:"TOTAL_OPERATE_INCOME"`               // 营业总收入
	RoeWeight                 float64 `json:"roe_weight" map:"ROE_WEIGHT"`                                   // 净资产收益率ROE
	Jroa                      float64 `json:"jroa" map:"JROA"`                                               // 总资产净利率ROA
	Roic                      float64 `json:"roic" map:"ROIC"`                                               // 投入资本回报率ROIC
	Zxgxl                     float64 `json:"zxgxl" map:"ZXGXL"`                                             // 最新股息率
	SaleGpr                   float64 `json:"sale_gpr" map:"SALE_GPR"`                                       // 毛利率
	SaleNpr                   float64 `json:"sale_npr" map:"SALE_NPR"`                                       // 净利率
	NetprofitYoyRatio         float64 `json:"netprofit_yoy_ratio" map:"NETPROFIT_YOY_RATIO"`                 // 净利润增长率
	DeductNetprofitGrowthrate float64 `json:"deduct_netprofit_growthrate" map:"DEDUCT_NETPROFIT_GROWTHRATE"` // 扣非净利润增长率
	ToiYoyRatio               float64 `json:"toi_yoy_ratio" map:"TOI_YOY_RATIO"`                             // 营收增长率
	NetprofitGrowthrate3Y     float64 `json:"netprofit_growthrate_3y" map:"NETPROFIT_GROWTHRATE_3Y"`         // 净利润3年复合增长率
	IncomeGrowthrate3Y        float64 `json:"income_growthrate_3y" map:"INCOME_GROWTHRATE_3Y"`               // 营收3年复合增长率
	PredictNetprofitRatio     float64 `json:"predict_netprofit_ratio" map:"PREDICT_NETPROFIT_RATIO"`         // 预测净利润同比增长
	PredictIncomeRatio        float64 `json:"predict_income_ratio" map:"PREDICT_INCOME_RATIO"`               // 预测营收同比增长
	BasicepsYoyRatio          float64 `json:"basiceps_yoy_ratio" map:"BASICEPS_YOY_RATIO"`                   // 每股收益同比增长率
	TotalProfitGrowthrate     float64 `json:"total_profit_growthrate" map:"TOTAL_PROFIT_GROWTHRATE"`         // 利润总额同比增长率
	OperateProfitGrowthrate   float64 `json:"operate_profit_growthrate" map:"OPERATE_PROFIT_GROWTHRATE"`     // 营业利润同比增长率
	DebtAssetRatio            float64 `json:"debt_asset_ratio" map:"DEBT_ASSET_RATIO"`                       // 资产负债率
	EquityRatio               float64 `json:"equity_ratio" map:"EQUITY_RATIO"`                               // 产权比率
	EquityMultiplier          float64 `json:"equity_multiplier" map:"EQUITY_MULTIPLIER"`                     // 权益乘数
	CurrentRatio              float64 `json:"current_ratio" map:"CURRENT_RATIO"`                             // 流动比率
	SpeedRatio                float64 `json:"speed_ratio" map:"SPEED_RATIO"`                                 // 速动比率
	TotalShares               int64   `json:"total_shares" map:"TOTAL_SHARES"`                               // 总股本
	FreeShares                int64   `json:"free_shares" map:"FREE_SHARES"`                                 // 流通股本
	HolderNewest              int64   `json:"holder_newest" map:"HOLDER_NEWEST"`                             // 最新股东户数
	HolderRatio               float64 `json:"holder_ratio" map:"HOLDER_RATIO"`                               // 股东户数增长率
	HoldAmount                float64 `json:"hold_amount" map:"HOLD_AMOUNT"`                                 // 户均持股金额
	AvgHoldNum                float64 `json:"avg_hold_num" map:"AVG_HOLD_NUM"`                               // 户均持股数量
	HoldnumGrowthrate3Q       float64 `json:"holdnum_growthrate_3q" map:"HOLDNUM_GROWTHRATE_3Q"`             // 户均持股数季度增长率
	HoldnumGrowthrateHy       float64 `json:"holdnum_growthrate_hy" map:"HOLDNUM_GROWTHRATE_HY"`             // 户均持股数半年增长率
	HoldRatioCount            float64 `json:"hold_ratio_count" map:"HOLD_RATIO_COUNT"`                       // 十大股东持股比例合计
	FreeHoldRatio             float64 `json:"free_hold_ratio" map:"FREE_HOLD_RATIO"`                         // 十大流通股东比例合计
	MacdGoldenFork            bool    `json:"macd_golden_fork" map:"MACD_GOLDEN_FORK"`                       // MACD金叉日线
	MacdGoldenForkz           bool    `json:"macd_golden_forkz" map:"MACD_GOLDEN_FORKZ"`                     // MACD金叉周线
	MacdGoldenForky           bool    `json:"macd_golden_forky" map:"MACD_GOLDEN_FORKY"`                     // MACD金叉月线
	KdjGoldenFork             bool    `json:"kdj_golden_fork" map:"KDJ_GOLDEN_FORK"`                         // KDJ金叉日线
	KdjGoldenForkz            bool    `json:"kdj_golden_forkz" map:"KDJ_GOLDEN_FORKZ"`                       // KDJ金叉周线
	KdjGoldenForky            bool    `json:"kdj_golden_forky" map:"KDJ_GOLDEN_FORKY"`                       // KDJ金叉月线
	BreakThrough              bool    `json:"break_through" map:"BREAK_THROUGH"`                             // 放量突破
	LowFundsInflow            bool    `json:"low_funds_inflow" map:"LOW_FUNDS_INFLOW"`                       // 低位资金净流入
	HighFundsOutflow          bool    `json:"high_funds_outflow" map:"HIGH_FUNDS_OUTFLOW"`                   // 高位资金净流出
	BreakupMa5days            bool    `json:"breakup_ma_5days" map:"BREAKUP_MA_5DAYS"`                       // 向上突破均线5日
	BreakupMa10days           bool    `json:"breakup_ma_10days" map:"BREAKUP_MA_10DAYS"`                     // 向上突破均线10日
	BreakupMa20days           bool    `json:"breakup_ma_20days" map:"BREAKUP_MA_20DAYS"`                     // 向上突破均线20日
	BreakupMa30days           bool    `json:"breakup_ma_30days" map:"BREAKUP_MA_30DAYS"`                     // 向上突破均线30日
	BreakupMa60days           bool    `json:"breakup_ma_60days" map:"BREAKUP_MA_60DAYS"`                     // 向上突破均线60日
	LongAvgArray              bool    `json:"long_avg_array" map:"LONG_AVG_ARRAY"`                           // 均线多头排列
	ShortAvgArray             bool    `json:"short_avg_array" map:"SHORT_AVG_ARRAY"`                         // 均线空头排列
	UpperLargeVolume          bool    `json:"upper_large_volume" map:"UPPER_LARGE_VOLUME"`                   // 连涨放量
	DownNarrowVolume          bool    `json:"down_narrow_volume" map:"DOWN_NARROW_VOLUME"`                   // 下跌无量
	OneDayangLine             bool    `json:"one_dayang_line" map:"ONE_DAYANG_LINE"`                         // 一根大阳线
	TwoDayangLines            bool    `json:"two_dayang_lines" map:"TWO_DAYANG_LINES"`                       // 两根大阳线
	RiseSun                   bool    `json:"rise_sun" map:"RISE_SUN"`                                       // 旭日东升
	PowerFulgun               bool    `json:"power_fulgun" map:"POWER_FULGUN"`                               // 强势多方炮
	RestoreJustice            bool    `json:"restore_justice" map:"RESTORE_JUSTICE"`                         // 拨云见日
	Down7days                 bool    `json:"down_7days" map:"DOWN_7DAYS"`                                   // 七仙女下凡(七连阴)
	Upper8days                bool    `json:"upper_8days" map:"UPPER_8DAYS"`                                 // 八仙过海(八连阳)
	Upper9days                bool    `json:"upper_9days" map:"UPPER_9DAYS"`                                 // 九阳神功(九连阳)
	Upper4days                bool    `json:"upper_4days" map:"UPPER_4DAYS"`                                 // 四串阳
	HeavenRule                bool    `json:"heaven_rule" map:"HEAVEN_RULE"`                                 // 天量法则
	UpsideVolume              bool    `json:"upside_volume" map:"UPSIDE_VOLUME"`                             // 放量上攻
	BearishEngulfing          bool    `json:"bearish_engulfing" map:"BEARISH_ENGULFING"`                     // 穿头破脚
	ReversingHammer           bool    `json:"reversing_hammer" map:"REVERSING_HAMMER"`                       // 倒转锤头
	ShootingStar              bool    `json:"shooting_star" map:"SHOOTING_STAR"`                             // 射击之星
	EveningStar               bool    `json:"evening_star" map:"EVENING_STAR"`                               // 黄昏之星
	FirstDawn                 bool    `json:"first_dawn" map:"FIRST_DAWN"`                                   // 曙光初现
	Pregnant                  bool    `json:"pregnant" map:"PREGNANT"`                                       // 身怀六甲
	BlackCloudTops            bool    `json:"black_cloud_tops" map:"BLACK_CLOUD_TOPS"`                       // 乌云盖顶
	MorningStar               bool    `json:"morning_star" map:"MORNING_STAR"`                               // 早晨之星
	NarrowFinish              bool    `json:"narrow_finish" map:"NARROW_FINISH"`                             // 窄幅整理
	LimitedLiftF6m            bool    `json:"limited_lift_f6m" map:"LIMITED_LIFT_F6M"`                       // 限售解禁未来半年
	LimitedLiftF1y            bool    `json:"limited_lift_f1y" map:"LIMITED_LIFT_F1Y"`                       // 限售解禁未来1年
	LimitedLift6m             bool    `json:"limited_lift_6m" map:"LIMITED_LIFT_6M"`                         // 限售解禁近半年
	LimitedLift1y             bool    `json:"limited_lift_1y" map:"LIMITED_LIFT_1Y"`                         // 限售解禁近1年
	DirectionalSeo1m          bool    `json:"directional_seo_1m" map:"DIRECTIONAL_SEO_1M"`                   // 定向增发近1个月
	DirectionalSeo3m          bool    `json:"directional_seo_3m" map:"DIRECTIONAL_SEO_3M"`                   // 定向增发近3个月
	DirectionalSeo6m          bool    `json:"directional_seo_6m" map:"DIRECTIONAL_SEO_6M"`                   // 定向增发近6个月
	DirectionalSeo1y          bool    `json:"directional_seo_1y" map:"DIRECTIONAL_SEO_1Y"`                   // 定向增发近1年
	Recapitalize1m            bool    `json:"recapitalize_1m" map:"RECAPITALIZE_1M"`                         // 资产重组近1个月
	Recapitalize3m            bool    `json:"recapitalize_3m" map:"RECAPITALIZE_3M"`                         // 资产重组近3个月
	Recapitalize6m            bool    `json:"recapitalize_6m" map:"RECAPITALIZE_6M"`                         // 资产重组近6个月
	Recapitalize1y            bool    `json:"recapitalize_1y" map:"RECAPITALIZE_1Y"`                         // 资产重组近1年
	EquityPledge1m            bool    `json:"equity_pledge_1m" map:"EQUITY_PLEDGE_1M"`                       // 股权质押近1个月
	EquityPledge3m            bool    `json:"equity_pledge_3m" map:"EQUITY_PLEDGE_3M"`                       // 股权质押近3个月
	EquityPledge6m            bool    `json:"equity_pledge_6m" map:"EQUITY_PLEDGE_6M"`                       // 股权质押近6个月
	EquityPledge1y            bool    `json:"equity_pledge_1y" map:"EQUITY_PLEDGE_1Y"`                       // 股权质押近1年
	PledgeRatio               float64 `json:"pledge_ratio" map:"PLEDGE_RATIO"`                               // 质押比例
	GoodwillScale             int64   `json:"goodwill_scale" map:"GOODWILL_SCALE"`                           // 商誉规模
	GoodwillAssetsRatro       float64 `json:"goodwill_assets_ratro" map:"GOODWILL_ASSETS_RATRO"`             // 商誉占净资产比例
	PredictType               string  `json:"predict_type" map:"PREDICT_TYPE"`                               // 业绩预告
	ParDividendPretax         float64 `json:"par_dividend_pretax" map:"PAR_DIVIDEND_PRETAX"`                 // 每股股利税前
	ParDividend               float64 `json:"par_dividend" map:"PAR_DIVIDEND"`                               // 每股红股
	ParItEquity               float64 `json:"par_it_equity" map:"PAR_IT_EQUITY"`                             // 每股转增股本
	HolderChange3m            float64 `json:"holder_change_3m" map:"HOLDER_CHANGE_3M"`                       // 近3月股东增减比例
	ExecutiveChange3m         float64 `json:"executive_change_3m" map:"EXECUTIVE_CHANGE_3M"`                 // 近3月高管增减比例
	OrgSurvey3m               int16   `json:"org_survey_3m" map:"ORG_SURVEY_3M"`                             // 近3月机构调研
	OrgRating                 string  `json:"org_rating" map:"ORG_RATING"`                                   // 机构评级
	AllcorpNum                int16   `json:"allcorp_num" map:"ALLCORP_NUM"`                                 // 机构持股家数合计
	AllcorpFundNum            int16   `json:"allcorp_fund_num" map:"ALLCORP_FUND_NUM"`                       // 基金持股家数
	AllcorpQsNum              int16   `json:"allcorp_qs_num" map:"ALLCORP_QS_NUM"`                           // 券商持股家数
	AllcorpQfiiNum            int16   `json:"allcorp_qfii_num" map:"ALLCORP_QFII_NUM"`                       // QFII持股家数
	AllcorpBxNum              int16   `json:"allcorp_bx_num" map:"ALLCORP_BX_NUM"`                           // 保险公司持股家数
	AllcorpSbNum              int16   `json:"allcorp_sb_num" map:"ALLCORP_SB_NUM"`                           // 社保持股家数
	AllcorpXtNum              int16   `json:"allcorp_xt_num" map:"ALLCORP_XT_NUM"`                           // 信托公司持股家数
	AllcorpRatio              float64 `json:"allcorp_ratio" map:"ALLCORP_RATIO"`                             // 机构持股比例合计
	AllcorpFundRatio          float64 `json:"allcorp_fund_ratio" map:"ALLCORP_FUND_RATIO"`                   // 基金持股比例
	AllcorpQsRatio            float64 `json:"allcorp_qs_ratio" map:"ALLCORP_QS_RATIO"`                       // 券商持股比例
	AllcorpQfiiRatio          float64 `json:"allcorp_qfii_ratio" map:"ALLCORP_QFII_RATIO"`                   // QFII持股比例
	AllcorpBxRatio            float64 `json:"allcorp_bx_ratio" map:"ALLCORP_BX_RATIO"`                       // 保险公司持股比例
	AllcorpSbRatio            float64 `json:"allcorp_sb_ratio" map:"ALLCORP_SB_RATIO"`                       // 社保持股比例
	AllcorpXtRatio            float64 `json:"allcorp_xt_ratio" map:"ALLCORP_XT_RATIO"`                       // 信托公司持股比例
	PopularityRank            int16   `json:"popularity_rank" map:"POPULARITY_RANK"`                         // 股吧人气排名
	RankChange                int16   `json:"rank_change" map:"RANK_CHANGE"`                                 // 人气排名变化
	UppDays                   int16   `json:"upp_days" map:"UPP_DAYS"`                                       // 人气排名连涨
	DownDays                  int16   `json:"down_days" map:"DOWN_DAYS"`                                     // 人气排名连跌
	NewHigh                   int16   `json:"new_high" map:"NEW_HIGH"`                                       // 人气排名创新高
	NewDown                   int16   `json:"new_down" map:"NEW_DOWN"`                                       // 人气排名创新低
	NewfansRatio              float64 `json:"newfans_ratio" map:"NEWFANS_RATIO"`                             // 新晋粉丝占比
	BigfansRatio              float64 `json:"bigfans_ratio" map:"BIGFANS_RATIO"`                             // 铁杆粉丝占比
	ConcernRank7days          int16   `json:"concern_rank_7days" map:"CONCERN_RANK_7DAYS"`                   // 7日关注排名
	BrowseRank                int16   `json:"browse_rank" map:"BROWSE_RANK"`                                 // 今日浏览排名
	Amplitude                 float64 `json:"amplitude" map:"AMPLITUDE"`                                     // 振幅
	IsIssueBreak              bool    `json:"is_issue_break" map:"IS_ISSUE_BREAK"`                           // 破发股票
	IsBpsBreak                bool    `json:"is_bps_break" map:"IS_BPS_BREAK"`                               // 破净股票
	NowNewhigh                bool    `json:"now_newhigh" map:"NOW_NEWHIGH"`                                 // 今日创历史新高
	NowNewlow                 bool    `json:"now_newlow" map:"NOW_NEWLOW"`                                   // 今日创历史新低
	HighRecent3days           bool    `json:"high_recent_3days" map:"HIGH_RECENT_3DAYS"`                     // 近期创历史新高近3日
	HighRecent5days           bool    `json:"high_recent_5days" map:"HIGH_RECENT_5DAYS"`                     // 近期创历史新高近5日
	HighRecent10days          bool    `json:"high_recent_10days" map:"HIGH_RECENT_10DAYS"`                   // 近期创历史新高近10日
	HighRecent20days          bool    `json:"high_recent_20days" map:"HIGH_RECENT_20DAYS"`                   // 近期创历史新高近20日
	HighRecent30days          bool    `json:"high_recent_30days" map:"HIGH_RECENT_30DAYS"`                   // 近期创历史新高近30日
	LowRecent3days            bool    `json:"low_recent_3days" map:"LOW_RECENT_3DAYS"`                       // 近期创历史新低近3日
	LowRecent5days            bool    `json:"low_recent_5days" map:"LOW_RECENT_5DAYS"`                       // 近期创历史新低近5日
	LowRecent10days           bool    `json:"low_recent_10days" map:"LOW_RECENT_10DAYS"`                     // 近期创历史新低近10日
	LowRecent20days           bool    `json:"low_recent_20days" map:"LOW_RECENT_20DAYS"`                     // 近期创历史新低近20日
	LowRecent30days           bool    `json:"low_recent_30days" map:"LOW_RECENT_30DAYS"`                     // 近期创历史新低近30日
	WinMarket3days            bool    `json:"win_market_3days" map:"WIN_MARKET_3DAYS"`                       // 近期跑赢大盘近3日
	WinMarket5days            bool    `json:"win_market_5days" map:"WIN_MARKET_5DAYS"`                       // 近期跑赢大盘近5日
	WinMarket10days           bool    `json:"win_market_10days" map:"WIN_MARKET_10DAYS"`                     // 近期跑赢大盘近10日
	WinMarket20days           bool    `json:"win_market_20days" map:"WIN_MARKET_20DAYS"`                     // 近期跑赢大盘近20日
	WinMarket30days           bool    `json:"win_market_30days" map:"WIN_MARKET_30DAYS"`                     // 近期跑赢大盘近30日
	NetInflow                 float64 `json:"net_inflow" map:"NET_INFLOW"`                                   // 当日净流入额
	Netinflow3days            int64   `json:"netinflow_3days" map:"NETINFLOW_3DAYS"`                         // 3日主力净流入
	Netinflow5days            int64   `json:"netinflow_5days" map:"NETINFLOW_5DAYS"`                         // 5日主力净流入
	NowinterstRatio           int64   `json:"nowinterst_ratio" map:"NOWINTERST_RATIO"`                       // 当日增仓占比
	NowinterstRatio3D         float64 `json:"nowinterst_ratio_3d" map:"NOWINTERST_RATIO_3D"`                 // 3日增仓占比
	NowinterstRatio5D         float64 `json:"nowinterst_ratio_5d" map:"NOWINTERST_RATIO_5D"`                 // 5日增仓占比
	Ddx                       float64 `json:"ddx" map:"DDX"`                                                 // 当日DDX
	Ddx3D                     float64 `json:"ddx_3d" map:"DDX_3D"`                                           // 3日DDX
	Ddx5D                     float64 `json:"ddx_5d" map:"DDX_5D"`                                           // 5日DDX
	DdxRed10D                 int16   `json:"ddx_red_10d" map:"DDX_RED_10D"`                                 // 10日内DDX飘红天数
	Changerate3days           float64 `json:"changerate_3days" map:"CHANGERATE_3DAYS"`                       // 3日涨跌幅
	Changerate5days           float64 `json:"changerate_5days" map:"CHANGERATE_5DAYS"`                       // 5日涨跌幅
	Changerate10days          float64 `json:"changerate_10days" map:"CHANGERATE_10DAYS"`                     // 10日涨跌幅
	ChangerateTy              float64 `json:"changerate_ty" map:"CHANGERATE_TY"`                             // 今年以来涨跌幅
	Upnday                    int16   `json:"upnday" map:"UPNDAY"`                                           // 连涨天数
	Downnday                  int16   `json:"downnday" map:"DOWNNDAY"`                                       // 连跌天数
	ListingYieldYear          float64 `json:"listing_yield_year" map:"LISTING_YIELD_YEAR"`                   // 上市以来年化收益率
	ListingVolatilityYear     float64 `json:"listing_volatility_year" map:"LISTING_VOLATILITY_YEAR"`         // 上市以来年化波动率
	MutualNetbuyAmt           int64   `json:"mutual_netbuy_amt" map:"MUTUAL_NETBUY_AMT"`                     // 沪深股通净买入金额
	HoldRatio                 float64 `json:"hold_ratio" map:"HOLD_RATIO"`                                   // 沪深股通持股比例
	Secucode                  string  `json:"secucode" map:"SECUCODE"`                                       // 全代码
}
