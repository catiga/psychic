package eli

import (
	"eli/app/model"
	"eli/database"
	ml "eli/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Fl(c *gin.Context) {

	lang := c.GetHeader("I18n-Language")

	var sysCatalogs []model.SysCatalog
	db := database.DB
	db.Find(&sysCatalogs)

	c.JSON(http.StatusOK, ml.Succ(lang, map[string]interface{}{"fl": sysCatalogs}))
}
