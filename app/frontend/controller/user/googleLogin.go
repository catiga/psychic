package user

import (
	"context"
	"eli/app/model"
	"eli/config"
	"eli/database"
	"eli/util"
	"fmt"
	"net/http"
	"time"

	ml "eli/middleware"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
)

func HandleGoogleCallback(c *gin.Context) {
	lang := c.GetHeader("I18n-Language")

	idToken := c.PostForm("idToken")

	// 验证 Google ID 令牌
	userInfo, err := validateGoogleIDToken(idToken)
	if err != nil {
		fmt.Printf("Error verifying Google ID token: %v\n", err)
		c.JSON(http.StatusOK, ml.Fail(lang, "100007"))
		return
	}

	if !userInfo.Claims["email_verified"].(bool) {
		c.JSON(http.StatusOK, ml.Fail(lang, "100008"))
	}

	db := database.GetDb()

	var user model.AccountUserInfo
	db.Where(" email = ? ", userInfo.Claims["email"].(string)).First(&user)

	if user == (model.AccountUserInfo{}) {

		user.ID, _ = util.Sf.GenerateID()
		user.Email = userInfo.Claims["email"].(string)
		user.Name = userInfo.Claims["name"].(string)
		user.Avatar = userInfo.Claims["picture"].(string)

		currentTime := time.Now()

		user.CreateAt = &currentTime
		user.IP = c.ClientIP()
		user.Type = 2 //谷歌登录

		result := db.Save(&user)

		if result.Error != nil {
			c.JSON(http.StatusOK, ml.Fail(lang, "100009"))
			return
		}
	}

	session := util.SessionToken{Id: user.ID, Email: "", Addr: user.Address, Type: int8(user.Type)}

	token, _ := util.Macke(&session)

	c.JSON(http.StatusOK, ml.Succ(lang, map[string]interface{}{"token": token}))
}

func validateGoogleIDToken(idToken string) (*idtoken.Payload, error) {
	audience := config.Get().Google.ClientId

	ctx := context.Background()

	payload, err := idtoken.Validate(ctx, idToken, audience)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ID token: %v", err)
	}

	return payload, nil
}
