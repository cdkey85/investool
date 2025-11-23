// 评论留言

package routes

import (
	"net/http"
	"strconv"

	"github.com/axiaoxin-com/investool/db"
	"github.com/axiaoxin-com/investool/models"
	"github.com/axiaoxin-com/investool/version"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// Comment godoc
func Comment(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage := 2
	offset := (page - 1) * perPage

	// 获取筛选参数
	stockCode := c.Query("stock_code")

	// 构建查询条件
	var comments []models.Comment
	query := db.DB.Order("created_at DESC").Offset(offset).Limit(perPage)

	// 如果提供了股票代码筛选条件
	if stockCode != "" {
		query = query.Where("stock_code = ?", stockCode)
	}

	// 执行查询
	result := query.Find(&comments)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	// 获取总记录数
	var total int64
	countQuery := db.DB.Model(&models.Comment{})
	if stockCode != "" {
		countQuery = countQuery.Where("stock_code = ?", stockCode)
	}
	countQuery.Count(&total)

	// 计算总页数
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	// 获取所有不同的股票代码用于筛选下拉框
	var stockCodes []string
	db.DB.Model(&models.Comment{}).Distinct("stock_code").Where("stock_code != ?", "").Pluck("stock_code", &stockCodes)

	data := gin.H{
		"Env":        viper.GetString("env"),
		"Version":    version.Version,
		"PageTitle":  "InvesTool | 感悟",
		"HostURL":    viper.GetString("server.host_url"),
		"Comments":   comments,
		"Page":       page,
		"TotalPages": totalPages,
		"Total":      total,
		"StockCode":  stockCode,
		"StockCodes": stockCodes,
	}
	c.HTML(http.StatusOK, "comment.html", data)
}

// AddComment 添加新留言
func AddComment(c *gin.Context) {
	var comment models.Comment

	// 绑定表单数据
	if err := c.ShouldBind(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存到数据库
	result := db.DB.Create(&comment)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment added successfully", "data": comment})
}

// DeleteComment 删除留言
func DeleteComment(c *gin.Context) {
	id := c.Param("id")

	result := db.DB.Delete(&models.Comment{}, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		}
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}
