// 在这个文件中注册 URL handler

package routes

import (
	"github.com/axiaoxin-com/investool/webserver"
	"github.com/gin-gonic/gin"
)

// Routes 注册 API URL 路由
func Routes(app *gin.Engine) {
	// 受保护的路由组 - 需要 HTTP Basic Auth
	protected := app.Group("", webserver.GinBasicAuth())
	{
		protected.GET("/", StockIndex)
		protected.GET("/stock", StockIndex)
		protected.POST("/selector", StockSelector)
		protected.POST("/checker", StockChecker)
		protected.GET("/comment", Comment)
		protected.POST("/comment", AddComment)
		protected.DELETE("/comment/:id", DeleteComment)
	}
	//protected.GET("/fund", FundIndex)
	//protected.GET("/fund/filter", FundFilter)
	//protected.POST("/fund/check", FundCheck)
	//protected.GET("/about", About)
	//protected.GET("/fund/similarity", FundSimilarity)
	//protected.GET("/materials", Materials)
	//protected.POST("/fund/query_by_stock", QueryFundByStock)
	//protected.GET("/fund/managers", FundManagers)
}
