package eli

import (
	"eli/app/model"
	"eli/database"
	ml "eli/middleware"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CoinList(c *gin.Context) {
	lang := c.GetHeader("I18n-Language")
	dateTime := c.Query("date_time")

	if len(dateTime) == 0 {
		dateTime = time.Now().Format("2006-01-02 15:04:05")
	}

	// 解析时间字符串
	_, err := time.Parse("2006-01-02 15:04:05", dateTime)
	if err != nil {
		c.JSON(http.StatusOK, ml.Fail(lang, "100011"))
		return
	}

	coinListResult := make(map[string]interface{})

	var coinList []model.CoinList
	db := database.DB
	db.Model(&model.CoinList{}).Find(&coinList)

	coinListResult["data"] = coinList

	// 返回结果
	// result := fmt.Sprintf("%s%s%s%s", yearGanZhi, monthGanZhi, dayGanZhi, hourGanZhi)
	// c.JSON(http.StatusOK, gin.H{"fourPillars": result})
	c.JSON(http.StatusOK, coinListResult)
}
