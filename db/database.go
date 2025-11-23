// CSV存储初始化

package db

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/axiaoxin-com/investool/models"
)

// InitDB 初始化存储
func InitDB() {
	// 初始化CSV存储
	InitCSV()

	// 打印日志表明使用的是CSV存储
	log.Println("Using CSV storage for comments")
}

// CSV文件路径
var csvFilePath string

// InitCSV 初始化CSV存储
func InitCSV() {
	// 使用相对路径，使程序在不同平台上都能正确运行
	workDir, _ := os.Getwd()
	csvFilePath = filepath.Join(workDir, "data", "comments.csv")

	// 创建 data 目录（如果不存在）
	dataDir := filepath.Dir(csvFilePath)
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		log.Printf("Failed to create data directory: %v", err)
	}

	// 如果CSV文件不存在，创建并写入表头
	if _, err := os.Stat(csvFilePath); os.IsNotExist(err) {
		file, err := os.Create(csvFilePath)
		if err != nil {
			log.Printf("Failed to create CSV file: %v", err)
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// 写入表头
		header := []string{"ID", "StockCode", "StockName", "Content", "CreatedAt", "UpdatedAt"}
		if err := writer.Write(header); err != nil {
			log.Printf("Failed to write CSV header: %v", err)
		}
	}
}

// GetAllComments 获取所有评论，支持分页和筛选
func GetAllComments(page, perPage int, stockCode string) ([]models.Comment, int64, error) {
	file, err := os.Open(csvFilePath)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// 读取表头
	_, err = reader.Read()
	if err != nil {
		return nil, 0, err
	}

	var allComments []models.Comment
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, err
		}

		// 解析记录
		comment, err := parseCommentFromRecord(record)
		if err != nil {
			continue // 跳过解析失败的记录
		}

		// 如果提供了股票代码筛选条件
		if stockCode != "" && comment.StockCode != stockCode {
			continue
		}

		allComments = append(allComments, comment)
	}

	// 按创建时间倒序排列
	sort.Slice(allComments, func(i, j int) bool {
		return allComments[i].CreatedAt.After(allComments[j].CreatedAt)
	})

	// 计算总数
	total := len(allComments)

	// 分页处理
	start := (page - 1) * perPage
	end := start + perPage
	if start >= total {
		return []models.Comment{}, int64(total), nil
	}
	if end > total {
		end = total
	}

	comments := allComments[start:end]
	return comments, int64(total), nil
}

// GetDistinctStockCodes 获取所有不同的股票代码
func GetDistinctStockCodes() ([]string, error) {
	file, err := os.Open(csvFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// 读取表头
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	stockCodeMap := make(map[string]bool)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(record) >= 2 && record[1] != "" {
			stockCodeMap[record[1]] = true
		}
	}

	var stockCodes []string
	for code := range stockCodeMap {
		stockCodes = append(stockCodes, code)
	}

	sort.Strings(stockCodes)
	return stockCodes, nil
}

// CreateComment 创建新评论
func CreateComment(comment *models.Comment) error {
	// 生成ID（使用时间戳）
	comment.ID = time.Now().UnixNano()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	// 以追加模式打开文件
	file, err := os.OpenFile(csvFilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 转换为记录
	record := commentToRecord(comment)

	return writer.Write(record)
}

// DeleteComment 删除评论
func DeleteComment(id int64) error {
	// 读取所有记录
	file, err := os.Open(csvFilePath)
	if err != nil {
		return err
	}

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	file.Close()

	if err != nil {
		return err
	}

	// 查找并删除记录
	found := false
	for i, record := range records {
		if len(record) > 0 {
			recordID, err := strconv.ParseInt(record[0], 10, 64)
			if err == nil && recordID == id {
				// 删除记录
				records = append(records[:i], records[i+1:]...)
				found = true
				break
			}
		}
	}

	if !found {
		return fmt.Errorf("comment with ID %d not found", id)
	}

	// 重新写入文件
	file, err = os.Create(csvFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	return writer.WriteAll(records)
}

// parseCommentFromRecord 从CSV记录解析评论对象
func parseCommentFromRecord(record []string) (models.Comment, error) {
	var comment models.Comment

	if len(record) < 6 {
		return comment, fmt.Errorf("invalid record format")
	}

	id, err := strconv.ParseInt(record[0], 10, 64)
	if err != nil {
		return comment, err
	}

	createdAt, err := time.Parse("2006-01-02 15:04:05", record[4])
	if err != nil {
		return comment, err
	}

	updatedAt, err := time.Parse("2006-01-02 15:04:05", record[5])
	if err != nil {
		return comment, err
	}

	comment = models.Comment{
		ID:        id,
		StockCode: record[1],
		StockName: record[2],
		Content:   record[3],
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return comment, nil
}

// commentToRecord 将评论对象转换为CSV记录
func commentToRecord(comment *models.Comment) []string {
	return []string{
		strconv.FormatInt(comment.ID, 10),
		comment.StockCode,
		comment.StockName,
		comment.Content,
		comment.CreatedAt.Format("2006-01-02 15:04:05"),
		comment.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
