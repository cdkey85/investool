// 获取历史市盈率测试

package eniu

import (
	"context"
	"testing"
)

// TestQueryHistoricalPE 测试获取历史市盈率
func TestQueryHistoricalPE(t *testing.T) {
	ctx := context.Background()
	eniu := NewEniu()

	// 测试股票代码：平安银行 000001.sz
	secuCode := "000001.sz"
	resp, err := eniu.QueryHistoricalPE(ctx, secuCode)
	if err != nil {
		t.Errorf("QueryHistoricalPE failed: %v", err)
		return
	}

	if len(resp.Date) == 0 || len(resp.PE) == 0 {
		t.Errorf("QueryHistoricalPE returned empty data")
		return
	}

	if len(resp.Date) != len(resp.PE) {
		t.Errorf("QueryHistoricalPE Date and PE length mismatch: %d vs %d", len(resp.Date), len(resp.PE))
		return
	}

	t.Logf("QueryHistoricalPE success, records: %d", len(resp.Date))
	t.Logf("First date: %s, PE: %f", resp.Date[0], resp.PE[0])
	t.Logf("Latest date: %s, PE: %f", resp.Date[len(resp.Date)-1], resp.PE[len(resp.PE)-1])
}

// TestGetLatestPE 测试获取最新市盈率
func TestGetLatestPE(t *testing.T) {
	resp := RespHistoricalPE{
		Date: []string{"2024-01-01", "2024-01-02", "2024-01-03"},
		PE:   []float64{10.5, 11.2, 12.8},
	}

	latestPE := resp.GetLatestPE()
	expectedPE := 12.8

	if latestPE != expectedPE {
		t.Errorf("GetLatestPE failed: expected %f, got %f", expectedPE, latestPE)
	}

	// 测试空数据
	emptyResp := RespHistoricalPE{}
	if emptyResp.GetLatestPE() != 0 {
		t.Errorf("GetLatestPE should return 0 for empty data")
	}
}

// TestGetAveragePE 测试获取平均市盈率
func TestGetAveragePE(t *testing.T) {
	resp := RespHistoricalPE{
		Date: []string{"2024-01-01", "2024-01-02", "2024-01-03", "2024-01-04", "2024-01-05"},
		PE:   []float64{10.0, 11.0, 12.0, 13.0, 14.0},
	}

	// 测试全部数据平均值
	avgAll := resp.GetAveragePE(0)
	expectedAvgAll := 12.0
	if avgAll != expectedAvgAll {
		t.Errorf("GetAveragePE(0) failed: expected %f, got %f", expectedAvgAll, avgAll)
	}

	// 测试最近3天平均值
	avg3Days := resp.GetAveragePE(3)
	expectedAvg3 := 13.0
	if avg3Days != expectedAvg3 {
		t.Errorf("GetAveragePE(3) failed: expected %f, got %f", expectedAvg3, avg3Days)
	}

	// 测试空数据
	emptyResp := RespHistoricalPE{}
	if emptyResp.GetAveragePE(0) != 0 {
		t.Errorf("GetAveragePE should return 0 for empty data")
	}
}

// TestGetMinMaxPE 测试获取最小最大市盈率
func TestGetMinMaxPE(t *testing.T) {
	resp := RespHistoricalPE{
		Date: []string{"2024-01-01", "2024-01-02", "2024-01-03", "2024-01-04", "2024-01-05"},
		PE:   []float64{10.0, 15.0, 8.0, 12.0, 14.0},
	}

	// 测试全部数据
	minPE, maxPE := resp.GetMinMaxPE(0)
	if minPE != 8.0 || maxPE != 15.0 {
		t.Errorf("GetMinMaxPE(0) failed: expected (8.0, 15.0), got (%f, %f)", minPE, maxPE)
	}

	// 测试最近3天
	min3, max3 := resp.GetMinMaxPE(3)
	if min3 != 8.0 || max3 != 14.0 {
		t.Errorf("GetMinMaxPE(3) failed: expected (8.0, 14.0), got (%f, %f)", min3, max3)
	}

	// 测试空数据
	emptyResp := RespHistoricalPE{}
	minEmpty, maxEmpty := emptyResp.GetMinMaxPE(0)
	if minEmpty != 0 || maxEmpty != 0 {
		t.Errorf("GetMinMaxPE should return (0, 0) for empty data")
	}
}
