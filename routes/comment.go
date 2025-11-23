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
)

// Comment godoc
func Comment(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage := 100

	// 获取筛选参数
	stockCode := c.Query("stock_code")

	// 从CSV文件获取评论数据
	comments, total, err := db.GetAllComments(page, perPage, stockCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments: " + err.Error()})
		return
	}

	// 计算总页数
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	// 获取所有不同的股票代码用于筛选下拉框
	stockCodes, err := db.GetDistinctStockCodes()
	if err != nil {
		stockCodes = []string{} // 如果出错，返回空数组而不是失败
	}

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

	// 保存到CSV文件
	err := db.CreateComment(&comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save comment: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment added successfully", "data": comment})
}

// DeleteComment 删除留言
func DeleteComment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	err = db.DeleteComment(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}