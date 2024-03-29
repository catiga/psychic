package router

import (
	"eli/app/frontend/controller/dis"
	"eli/app/frontend/controller/eli"
	"eli/app/frontend/controller/user"
	"eli/interceptor"

	"eli/app/frontend/controller/chat"

	"github.com/gin-gonic/gin"
)

func Load(r *gin.RouterGroup) {
	userGroup := r.Group("/user", interceptor.LoggerMiddleware(), interceptor.FrontendSignMiddleware(), interceptor.FrontendAuthMiddleware())

	userGroup.POST("/sign_msg", user.GetSignMsg)
	userGroup.POST("/login", user.Login)
	userGroup.GET("/login_user_info", user.LoginUserInfo)
	userGroup.POST("/handle_google_callback", user.HandleGoogleCallback)
	userGroup.POST("/register_account", user.RegisterAccount)
	userGroup.POST("/account_login", user.AccountLogin)

	eliGroup := r.Group("/eli", interceptor.LoggerMiddleware(), interceptor.FrontendSignMiddleware(), interceptor.FrontendAuthMiddleware())
	eliGroup.GET("/calculateFourPillars", eli.CalculateFourPillars)
	eliGroup.POST("/calculateShenKe", eli.CalculateShenKe)
	eliGroup.POST("/fl", eli.Fl)
	eliGroup.POST("/web3/coinList", eli.CoinList)
	eliGroup.POST("/sys/catalog", eli.Catalog)
	eliGroup.POST("/sys/catques", eli.PrepareQuestion)

	disGroup := r.Group("/dis", interceptor.LoggerMiddleware(), interceptor.FrontendSignMiddleware(), interceptor.FrontendAuthMiddleware())
	disGroup.POST("/check_white", dis.CheckWhite)
	disGroup.POST("/get_maker_proof", dis.GetMakerProof)

	wsGroup := r.Group("/ws")

	wsGroup.GET("chat", chat.Chat)
}
