package user

import (
	"eli/app/model"
	"eli/constant"
	"eli/database"
	ml "eli/middleware"
	"eli/util"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

func GetSignMsg(c *gin.Context) {

	lang := c.GetHeader("I18n-Language")
	address := c.PostForm("address")

	if len(address) == 0 {
		c.JSON(http.StatusOK, ml.Fail(lang, "100001"))
		return
	}

	id, _ := util.Sf.GenerateID()

	str := fmt.Sprintf("Welcome to eli \n%d", id)

	util.CachePut(fmt.Sprintf(constant.KeyAddrSign, address), str, 1*time.Hour)

	c.JSON(http.StatusOK, ml.Succ(lang, map[string]interface{}{"msg": str}))
}

func Login(c *gin.Context) {

	lang := c.GetHeader("I18n-Language")
	// lType := c.PostForm("l_type") //现在就钱包登录
	sign := c.PostForm("sign")
	address := c.PostForm("address")
	// device := c.GetHeader("device")
	// chain_id := c.PostForm("chainId")

	if len(sign) == 0 {
		c.JSON(http.StatusOK, ml.Fail(lang, "100002"))
		return
	}

	//验证钱包
	msg, err := util.CacheGet(fmt.Sprintf(constant.KeyAddrSign, address))

	if !err {
		c.JSON(http.StatusOK, ml.Fail(lang, "100003"))
		return
	}

	msg = "\x19Ethereum Signed Message:\n" + strconv.Itoa(len(msg.(string))) + msg.(string)

	result := util.Verify(sign, msg.(string), address)

	if !result {
		c.JSON(http.StatusOK, ml.Fail(lang, "100003"))
		return
	}

	db := database.DB

	var accountUserInfo model.AccountUserInfo
	db.Where(" address = ? ", address).First(&accountUserInfo)

	if accountUserInfo == (model.AccountUserInfo{}) {

		currentTime := time.Now()

		accountUserInfo.ID, _ = util.Sf.GenerateID()
		accountUserInfo.Name = util.FormatEthereumAddress(address)
		accountUserInfo.Address = address
		accountUserInfo.CreateAt = &currentTime
		accountUserInfo.Type = 1 //谷歌注册
		accountUserInfo.IP = c.ClientIP()
		accountUserInfo.Kyc = 1

		db.Save(&accountUserInfo)
	}

	session := util.SessionToken{Id: accountUserInfo.ID, Email: "", Addr: accountUserInfo.Address, Type: int8(accountUserInfo.Type), Name: ""}

	token, _ := util.Macke(&session)

	c.JSON(http.StatusOK, ml.Succ(lang, map[string]interface{}{"token": token}))

	util.CacheDel(fmt.Sprintf(constant.KeyAddrSign, address))
}

func RegisterAccount(c *gin.Context) {

	lang := c.GetHeader("I18n-Language")
	name := c.PostForm("user_name")
	pwd := c.PostForm("pwd")

	db := database.DB

	if name == "" {
		c.JSON(http.StatusOK, ml.Fail(lang, "100016"))
		return
	}

	if pwd == "" {
		c.JSON(http.StatusOK, ml.Fail(lang, "100017"))
		return
	}

	if utf8.RuneCountInString(name) < 6 || utf8.RuneCountInString(name) > 16 {
		c.JSON(http.StatusOK, ml.Fail(lang, "100019"))
		return
	}

	currentTime := time.Now()
	var accountUserInfo model.AccountUserInfo

	db.Where(" name = ? ", name).First(&accountUserInfo)

	if accountUserInfo != (model.AccountUserInfo{}) {
		c.JSON(http.StatusOK, ml.Fail(lang, "100018"))
		return
	}

	accountUserInfo.Name = name
	accountUserInfo.Pwd = util.ToMd5AndSalt(pwd)
	accountUserInfo.CreateAt = &currentTime
	accountUserInfo.Type = 2 //账号密码
	accountUserInfo.IP = c.ClientIP()
	accountUserInfo.Kyc = 1
	db.Save(&accountUserInfo)

	c.JSON(http.StatusOK, ml.Succ(lang, nil))

}

func AccountLogin(c *gin.Context) {

	lang := c.GetHeader("I18n-Language")
	name := c.PostForm("user_name")
	pwd := c.PostForm("pwd")

	db := database.DB

	if name == "" {
		c.JSON(http.StatusOK, ml.Fail(lang, "100016"))
		return
	}

	if pwd == "" {
		c.JSON(http.StatusOK, ml.Fail(lang, "100017"))
		return
	}

	if utf8.RuneCountInString(name) < 6 || utf8.RuneCountInString(name) > 16 {
		c.JSON(http.StatusOK, ml.Fail(lang, "100019"))
		return
	}

	var accountUserInfo model.AccountUserInfo

	db.Where(" name = ? ", name).First(&accountUserInfo)

	if accountUserInfo == (model.AccountUserInfo{}) {
		c.JSON(http.StatusOK, ml.Fail(lang, "100020"))
		return
	}

	session := util.SessionToken{Id: accountUserInfo.ID, Email: "", Addr: accountUserInfo.Address, Type: int8(accountUserInfo.Type), Name: ""}

	token, _ := util.Macke(&session)

	c.JSON(http.StatusOK, ml.Succ(lang, map[string]interface{}{"token": token}))
}

func LoginUserInfo(c *gin.Context) {

	lang := c.GetHeader("I18n-Language")
	token := c.GetHeader("token")
	session, _ := util.ParseToken(token)

	var accountUserInfo model.AccountUserInfo
	database.DB.First(&accountUserInfo, session.Id)

	//先这样 看需求再调整结构
	c.JSON(http.StatusOK, ml.Succ(lang, map[string]interface{}{"user": accountUserInfo}))
}

// 登录接口
// 当前登录用户信息
// KYC接口
// 房屋列表接口设计
