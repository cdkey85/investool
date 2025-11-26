// 按我的选股指标获取股票数据，对优质公司进行初步筛选（好公司不代表股价涨）

package eastmoney

import (
	"context"
	"fmt"
	"reflect"
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
	Secucode              string      `json:"SECUCODE" title:"股票代码" show:"0"`
	SecurityCode          string      `json:"SECURITY_CODE" title:"股票代码" show:"1"`
	SecurityNameAbbr      string      `json:"SECURITY_NAME_ABBR" title:"股票名" show:"1"`
	Industry              string      `json:"INDUSTRY" title:"行业" show:"1"`
	RoeWeight             float64     `json:"ROE_WEIGHT" title:"ROE" show:"1"`
	NetprofitYoyRatio     float64     `json:"NETPROFIT_YOY_RATIO" title:"净利润增长率"`
	ToiYoyRatio           float64     `json:"TOI_YOY_RATIO" title:"营收增长率"`
	Zxgxl                 float64     `json:"ZXGXL" title:"最新股息率"`
	NetprofitGrowthrate3Y float64     `json:"NETPROFIT_GROWTHRATE_3Y" title:"净利润3年复合增长率"`
	IncomeGrowthrate3Y    float64     `json:"INCOME_GROWTHRATE_3Y" title:"营收3年复合增长率"`
	ListingYieldYear      float64     `json:"LISTING_YIELD_YEAR" title:"上市以来年化收益率"`
	PBNewMRQ              float64     `json:"PBNEWMRQ" title:"市净率"`
	PredictNetprofitRatio float64     `json:"PREDICT_NETPROFIT_RATIO" title:"预测净利润同比增长"`
	PredictIncomeRatio    float64     `json:"PREDICT_INCOME_RATIO" title:"预测营收同比增长"`
	TotalMarketCap        float64     `json:"TOTAL_MARKET_CAP" title:"总市值" show:"1"`
	ListingVolatilityYear float64     `json:"LISTING_VOLATILITY_YEAR" title:"上市以来年化波动率"`
	ListingDate           string      `json:"LISTING_DATE" title:"上市时间"`
	DebtAssetRatio        float64     `json:"DEBT_ASSET_RATIO" title:"资产负债率" show:"1"`
	PE                    float64     `json:"PE9" title:"滚动市盈率" show:"1"`
	DTSYL                 float64     `json:"DTSYL" title:"动态市盈率"`
	ChangeRate            interface{} `json:"change_rate" title:"涨跌幅"`
	VolumeRatio           interface{} `json:"volume_ratio" title:"量比"`
	NewPrice              interface{} `json:"NEW_PRICE" title:"最新价"`
	HighPrice             interface{} `json:"high_price" title:"最高价"`
	LowPrice              interface{} `json:"low_price" title:"最低价"`
	PreClosePrice         interface{} `json:"pre_close_price" title:"昨收价"`
	Volume                interface{} `json:"volume" title:"成交量"`
	DealAmount            interface{} `json:"deal_amount" title:"成交额"`
	Turnoverrate          float64     `json:"turnoverrate" title:"换手率"`
	Area                  string      `json:"area" title:"地区"`
	//Concept                   []string    `json:"concept" title:"概念"`	//多个概念时返回数组，单个概念时返回字符串，先不做处理
	Style                     []string    `json:"style" title:"板块"`
	IsHs300                   string      `json:"is_hs300" title:"沪300"`
	IsSz50                    string      `json:"is_sz50" title:"上证50"`
	IsZz500                   string      `json:"is_zz500" title:"中证500"`
	IsZz1000                  string      `json:"is_zz1000" title:"中证1000"`
	IsCy50                    string      `json:"is_cy50" title:"创业板50"`
	Pettmdeducted             float64     `json:"pettmdeducted" title:"市盈率TTM扣非"`
	Ps9                       float64     `json:"ps9" title:"市销率TTM"`
	Pcfjyxjl9                 float64     `json:"pcfjyxjl9" title:"市现率TTM"`
	PredictPeSyear            float64     `json:"predict_pe_syear" title:"预测市盈率今年"`
	PredictPeNyear            float64     `json:"predict_pe_nyear" title:"预测市盈率明年"`
	FreeCap                   int64       `json:"free_cap" title:"流通市值"`
	Ycpeg                     float64     `json:"ycpeg" title:"预测PEG"`
	EnterpriseValueMultiple   float64     `json:"enterprise_value_multiple" title:"企业价值倍数"`
	BasicEps                  float64     `json:"basic_eps" title:"每股收益"`
	Bvps                      float64     `json:"bvps" title:"每股净资产"`
	PerNetcashOperate         float64     `json:"per_netcash_operate" title:"每股经营现金流"`
	PerFcfe                   float64     `json:"per_fcfe" title:"每股自由现金流"`
	PerCapitalReserve         float64     `json:"per_capital_reserve" title:"每股资本公积"`
	PerUnassignProfit         float64     `json:"per_unassign_profit" title:"每股未分配利润"`
	PerSurplusReserve         float64     `json:"per_surplus_reserve" title:"每股盈余公积"`
	PerRetainedEarning        float64     `json:"per_retained_earning" title:"每股留存收益"`
	ParentNetprofit           int64       `json:"parent_netprofit" title:"归属净利润" show:"1"`
	DeductNetprofit           int64       `json:"deduct_netprofit" title:"扣非净利润" show:"1"`
	TotalOperateIncome        int64       `json:"total_operate_income" title:"营业总收入" show:"1"`
	ROA                       float64     `json:"jroa" title:"总资产净利率ROA"`
	Roic                      float64     `json:"roic" title:"投入资本回报率ROIC"`
	SaleGpr                   float64     `json:"sale_gpr" title:"毛利率" show:"1"`
	SaleNpr                   float64     `json:"sale_npr" title:"净利率" show:"1"`
	DeductNetprofitGrowthrate float64     `json:"deduct_netprofit_growthrate" title:"扣非净利润增长率"`
	BasicepsYoyRatio          float64     `json:"basiceps_yoy_ratio" title:"每股收益同比增长率"`
	TotalProfitGrowthrate     float64     `json:"total_profit_growthrate" title:"利润总额同比增长率"`
	OperateProfitGrowthrate   float64     `json:"operate_profit_growthrate" title:"营业利润同比增长率"`
	EquityRatio               float64     `json:"equity_ratio" title:"产权比率"`
	EquityMultiplier          float64     `json:"equity_multiplier" title:"权益乘数"`
	CurrentRatio              float64     `json:"current_ratio" title:"流动比率"`
	SpeedRatio                float64     `json:"speed_ratio" title:"速动比率"`
	TotalShares               int64       `json:"total_shares" title:"总股本"`
	FreeShares                int64       `json:"free_shares" title:"流通股本"`
	HolderNewest              int64       `json:"holder_newest" title:"最新股东户数"`
	HolderRatio               float64     `json:"holder_ratio" title:"股东户数增长率"`
	HoldAmount                float64     `json:"hold_amount" title:"户均持股金额"`
	AvgHoldNum                float64     `json:"avg_hold_num" title:"户均持股数量"`
	HoldnumGrowthrate3Q       float64     `json:"holdnum_growthrate_3q" title:"户均持股数季度增长率"`
	HoldnumGrowthrateHy       float64     `json:"holdnum_growthrate_hy" title:"户均持股数半年增长率"`
	HoldRatioCount            float64     `json:"hold_ratio_count" title:"十大股东持股比例合计"`
	FreeHoldRatio             float64     `json:"free_hold_ratio" title:"十大流通股东比例合计"`
	MacdGoldenFork            int16       `json:"macd_golden_fork" title:"MACD金叉日线" need:"0"`
	MacdGoldenForkz           int16       `json:"macd_golden_forkz" title:"MACD金叉周线" need:"0"`
	MacdGoldenForky           int16       `json:"macd_golden_forky" title:"MACD金叉月线" need:"0"`
	KdjGoldenFork             int16       `json:"kdj_golden_fork" title:"KDJ金叉日线" need:"0"`
	KdjGoldenForkz            int16       `json:"kdj_golden_forkz" title:"KDJ金叉周线" need:"0"`
	KdjGoldenForky            int16       `json:"kdj_golden_forky" title:"KDJ金叉月线" need:"0"`
	BreakThrough              int16       `json:"break_through" title:"放量突破" need:"0"`
	LowFundsInflow            int16       `json:"low_funds_inflow" title:"低位资金净流入" need:"0"`
	HighFundsOutflow          int16       `json:"high_funds_outflow" title:"高位资金净流出" need:"0"`
	BreakupMa5days            int16       `json:"breakup_ma_5days" title:"向上突破均线5日" need:"0"`
	BreakupMa10days           int16       `json:"breakup_ma_10days" title:"向上突破均线10日" need:"0"`
	BreakupMa20days           int16       `json:"breakup_ma_20days" title:"向上突破均线20日" need:"0"`
	BreakupMa30days           int16       `json:"breakup_ma_30days" title:"向上突破均线30日" need:"0"`
	BreakupMa60days           int16       `json:"breakup_ma_60days" title:"向上突破均线60日" need:"0"`
	LongAvgArray              int16       `json:"long_avg_array" title:"均线多头排列" need:"0"`
	ShortAvgArray             int16       `json:"short_avg_array" title:"均线空头排列" need:"0"`
	UpperLargeVolume          int16       `json:"upper_large_volume" title:"连涨放量" need:"0"`
	DownNarrowVolume          int16       `json:"down_narrow_volume" title:"下跌无量" need:"0"`
	OneDayangLine             int16       `json:"one_dayang_line" title:"一根大阳线" need:"0"`
	TwoDayangLines            int16       `json:"two_dayang_lines" title:"两根大阳线" need:"0"`
	RiseSun                   int16       `json:"rise_sun" title:"旭日东升" need:"0"`
	PowerFulgun               int16       `json:"power_fulgun" title:"强势多方炮" need:"0"`
	RestoreJustice            int16       `json:"restore_justice" title:"拨云见日" need:"0"`
	Down7days                 int16       `json:"down_7days" title:"七仙女下凡(七连阴)" need:"0"`
	Upper8days                int16       `json:"upper_8days" title:"八仙过海(八连阳)" need:"0"`
	Upper9days                int16       `json:"upper_9days" title:"九阳神功(九连阳)" need:"0"`
	Upper4days                int16       `json:"upper_4days" title:"四串阳" need:"0"`
	HeavenRule                int16       `json:"heaven_rule" title:"天量法则" need:"0"`
	UpsideVolume              int16       `json:"upside_volume" title:"放量上攻" need:"0"`
	BearishEngulfing          int16       `json:"bearish_engulfing" title:"穿头破脚" need:"0"`
	ReversingHammer           int16       `json:"reversing_hammer" title:"倒转锤头" need:"0"`
	ShootingStar              int16       `json:"shooting_star" title:"射击之星" need:"0"`
	EveningStar               int16       `json:"evening_star" title:"黄昏之星" need:"0"`
	FirstDawn                 int16       `json:"first_dawn" title:"曙光初现" need:"0"`
	Pregnant                  int16       `json:"pregnant" title:"身怀六甲" need:"0"`
	BlackCloudTops            int16       `json:"black_cloud_tops" title:"乌云盖顶" need:"0"`
	MorningStar               int16       `json:"morning_star" title:"早晨之星" need:"0"`
	NarrowFinish              int16       `json:"narrow_finish" title:"窄幅整理" need:"0"`
	LimitedLiftF6m            int16       `json:"limited_lift_f6m" title:"限售解禁未来半年" need:"0"`
	LimitedLiftF1y            int16       `json:"limited_lift_f1y" title:"限售解禁未来1年" need:"0"`
	LimitedLift6m             int16       `json:"limited_lift_6m" title:"限售解禁近半年" need:"0"`
	LimitedLift1y             int16       `json:"limited_lift_1y" title:"限售解禁近1年" need:"0"`
	DirectionalSeo1m          int16       `json:"directional_seo_1m" title:"定向增发近1个月" need:"0"`
	DirectionalSeo3m          int16       `json:"directional_seo_3m" title:"定向增发近3个月" need:"0"`
	DirectionalSeo6m          int16       `json:"directional_seo_6m" title:"定向增发近6个月" need:"0"`
	DirectionalSeo1y          int16       `json:"directional_seo_1y" title:"定向增发近1年" need:"0"`
	Recapitalize1m            int16       `json:"recapitalize_1m" title:"资产重组近1个月" need:"0"`
	Recapitalize3m            int16       `json:"recapitalize_3m" title:"资产重组近3个月" need:"0"`
	Recapitalize6m            int16       `json:"recapitalize_6m" title:"资产重组近6个月" need:"0"`
	Recapitalize1y            int16       `json:"recapitalize_1y" title:"资产重组近1年" need:"0"`
	EquityPledge1m            int16       `json:"equity_pledge_1m" title:"股权质押近1个月" need:"0"`
	EquityPledge3m            int16       `json:"equity_pledge_3m" title:"股权质押近3个月" need:"0"`
	EquityPledge6m            int16       `json:"equity_pledge_6m" title:"股权质押近6个月" need:"0"`
	EquityPledge1y            int16       `json:"equity_pledge_1y" title:"股权质押近1年" need:"0"`
	PledgeRatio               float64     `json:"pledge_ratio" title:"质押比例"`
	GoodwillScale             int64       `json:"goodwill_scale" title:"商誉规模"`
	GoodwillAssetsRatro       float64     `json:"goodwill_assets_ratro" title:"商誉占净资产比例"`
	PredictType               string      `json:"predict_type" title:"业绩预告"`
	ParDividendPretax         float64     `json:"par_dividend_pretax" title:"每股股利税前"`
	ParDividend               float64     `json:"par_dividend" title:"每股红股"`
	ParItEquity               float64     `json:"par_it_equity" title:"每股转增股本"`
	HolderChange3m            float64     `json:"holder_change_3m" title:"近3月股东增减比例"`
	ExecutiveChange3m         float64     `json:"executive_change_3m" title:"近3月高管增减比例"`
	OrgSurvey3m               int16       `json:"org_survey_3m" title:"近3月机构调研"`
	OrgRating                 string      `json:"org_rating" title:"机构评级"`
	AllcorpNum                int16       `json:"allcorp_num" title:"机构持股家数合计"`
	AllcorpFundNum            int16       `json:"allcorp_fund_num" title:"基金持股家数"`
	AllcorpQsNum              int16       `json:"allcorp_qs_num" title:"券商持股家数"`
	AllcorpQfiiNum            int16       `json:"allcorp_qfii_num" title:"QFII持股家数"`
	AllcorpBxNum              int16       `json:"allcorp_bx_num" title:"保险公司持股家数"`
	AllcorpSbNum              int16       `json:"allcorp_sb_num" title:"社保持股家数"`
	AllcorpXtNum              int16       `json:"allcorp_xt_num" title:"信托公司持股家数"`
	AllcorpRatio              float64     `json:"allcorp_ratio" title:"机构持股比例合计"`
	AllcorpFundRatio          float64     `json:"allcorp_fund_ratio" title:"基金持股比例"`
	AllcorpQsRatio            float64     `json:"allcorp_qs_ratio" title:"券商持股比例"`
	AllcorpQfiiRatio          float64     `json:"allcorp_qfii_ratio" title:"QFII持股比例"`
	AllcorpBxRatio            float64     `json:"allcorp_bx_ratio" title:"保险公司持股比例"`
	AllcorpSbRatio            float64     `json:"allcorp_sb_ratio" title:"社保持股比例"`
	AllcorpXtRatio            float64     `json:"allcorp_xt_ratio" title:"信托公司持股比例"`
	PopularityRank            int16       `json:"popularity_rank" title:"股吧人气排名"`
	RankChange                int16       `json:"rank_change" title:"人气排名变化"`
	UppDays                   int16       `json:"upp_days" title:"人气排名连涨"`
	DownDays                  int16       `json:"down_days" title:"人气排名连跌"`
	NewHigh                   int16       `json:"new_high" title:"人气排名创新高"`
	NewDown                   int16       `json:"new_down" title:"人气排名创新低"`
	NewfansRatio              float64     `json:"newfans_ratio" title:"新晋粉丝占比"`
	BigfansRatio              float64     `json:"bigfans_ratio" title:"铁杆粉丝占比"`
	ConcernRank7days          int16       `json:"concern_rank_7days" title:"7日关注排名"`
	BrowseRank                int16       `json:"browse_rank" title:"今日浏览排名"`
	Amplitude                 interface{} `json:"amplitude" title:"振幅"`
	IsIssueBreak              int16       `json:"is_issue_break" title:"破发股票"`
	IsBpsBreak                int16       `json:"is_bps_break" title:"破净股票"`
	NowNewhigh                int16       `json:"now_newhigh" title:"今日创历史新高" need:"0"`
	NowNewlow                 int16       `json:"now_newlow" title:"今日创历史新低" need:"0"`
	HighRecent3days           int16       `json:"high_recent_3days" title:"近期创历史新高近3日" need:"0"`
	HighRecent5days           int16       `json:"high_recent_5days" title:"近期创历史新高近5日" need:"0"`
	HighRecent10days          int16       `json:"high_recent_10days" title:"近期创历史新高近10日" need:"0"`
	HighRecent20days          int16       `json:"high_recent_20days" title:"近期创历史新高近20日" need:"0"`
	HighRecent30days          int16       `json:"high_recent_30days" title:"近期创历史新高近30日" need:"0"`
	LowRecent3days            int16       `json:"low_recent_3days" title:"近期创历史新低近3日" need:"0"`
	LowRecent5days            int16       `json:"low_recent_5days" title:"近期创历史新低近5日" need:"0"`
	LowRecent10days           int16       `json:"low_recent_10days" title:"近期创历史新低近10日" need:"0"`
	LowRecent20days           int16       `json:"low_recent_20days" title:"近期创历史新低近20日" need:"0"`
	LowRecent30days           int16       `json:"low_recent_30days" title:"近期创历史新低近30日" need:"0"`
	WinMarket3days            int16       `json:"win_market_3days" title:"近期跑赢大盘近3日" need:"0"`
	WinMarket5days            int16       `json:"win_market_5days" title:"近期跑赢大盘近5日" need:"0"`
	WinMarket10days           int16       `json:"win_market_10days" title:"近期跑赢大盘近10日" need:"0"`
	WinMarket20days           int16       `json:"win_market_20days" title:"近期跑赢大盘近20日" need:"0"`
	WinMarket30days           int16       `json:"win_market_30days" title:"近期跑赢大盘近30日" need:"0"`
	NetInflow                 interface{} `json:"net_inflow" title:"当日净流入额" need:"0"`
	Netinflow3days            int64       `json:"netinflow_3days" title:"3日主力净流入" need:"0"`
	Netinflow5days            int64       `json:"netinflow_5days" title:"5日主力净流入" need:"0"`
	NowinterstRatio           interface{} `json:"nowinterst_ratio" title:"当日增仓占比" need:"0"`
	NowinterstRatio3D         float64     `json:"nowinterst_ratio_3d" title:"3日增仓占比" need:"0"`
	NowinterstRatio5D         float64     `json:"nowinterst_ratio_5d" title:"5日增仓占比" need:"0"`
	Ddx                       interface{} `json:"ddx" title:"当日DDX" need:"0"`
	Ddx3D                     interface{} `json:"ddx_3d" title:"3日DDX" need:"0"`
	Ddx5D                     interface{} `json:"ddx_5d" title:"5日DDX" need:"0"`
	DdxRed10D                 interface{} `json:"ddx_red_10d" title:"10日内DDX飘红天数" need:"0"`
	Changerate3days           float64     `json:"changerate_3days" title:"3日涨跌幅" need:"0"`
	Changerate5days           float64     `json:"changerate_5days" title:"5日涨跌幅" need:"0"`
	Changerate10days          float64     `json:"changerate_10days" title:"10日涨跌幅" need:"0"`
	ChangerateTy              float64     `json:"changerate_ty" title:"今年以来涨跌幅"`
	Upnday                    int16       `json:"upnday" title:"连涨天数" need:"0"`
	Downnday                  int16       `json:"downnday" title:"连跌天数" need:"0"`
	MutualNetbuyAmt           int64       `json:"mutual_netbuy_amt" title:"沪深股通净买入金额"`
	HoldRatio                 float64     `json:"hold_ratio" title:"沪深股通持股比例"`
}

// ColumnInfo 表示列信息
type ColumnInfo struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	DefaultShow bool   `json:"default_show"`
}

// 根据StockInfo结构体字段获取表头信息
func (s StockInfoList) GetColumnInfos() []ColumnInfo {
	columnInfos := []ColumnInfo{}

	stockInfoType := reflect.TypeOf(StockInfo{})
	for i := 0; i < stockInfoType.NumField(); i++ {
		field := stockInfoType.Field(i)
		columnInfos = append(columnInfos, ColumnInfo{
			Key:         field.Tag.Get("json"),
			Title:       field.Tag.Get("title"),
			DefaultShow: field.Tag.Get("show") == "1",
		})
	}
	return columnInfos
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
		"filter": filter.String(),
		"p":      "1",      // page
		"ps":     "100000", // page size
	}

	// 自动获取StockInfo结构体内的字段，作为请求参数
	stockInfoType := reflect.TypeOf(StockInfo{})
	sty := ""
	for i := 0; i < stockInfoType.NumField(); i++ {
		field := stockInfoType.Field(i)
		need := field.Tag.Get("need")
		if need != "0" {
			sty += field.Tag.Get("json") + ","
		}
	}
	reqData["sty"] = sty

	logging.Info(ctx, "EastMoney QuerySelectedStocksWithFilter "+apiurl+" begin", zap.Any("reqData", reqData))
	beginTime := time.Now()
	req, err := goutils.NewHTTPMultipartReq(ctx, apiurl, reqData)
	if err != nil {
		return nil, err
	}
	resp := RespSelectStocks{}
	err = goutils.HTTPPOST(ctx, e.HTTPClient, req, &resp)
	latency := time.Since(beginTime).Milliseconds()
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
		if filter.ExcludeCYB && strings.HasPrefix(i.SecurityCode, "300") {
			logging.Debugf(ctx, "EastMoney SelectStocksWithFilter ExcludeCYB %s %s", i.SecurityNameAbbr, i.SecurityCode)
			continue
		}
		// 排除科创板
		if filter.ExcludeKCB && strings.HasPrefix(i.SecurityCode, "688") {
			logging.Debugf(ctx, "EastMoney SelectStocksWithFilter ExcludeKCB %s %s", i.SecurityNameAbbr, i.SecurityCode)
			continue
		}
		result = append(result, i)
	}
	return result, nil
}
