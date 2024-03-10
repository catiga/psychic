package eli

import (
	"eli/app/frontend/controller/chat"
	"eli/app/model"
	"eli/database"
	ml "eli/middleware"
	"eli/util"
	"eli/util/chronos"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 提交接口 下单
func PreOrder(c *gin.Context) {

	lang := c.GetHeader("I18n-Language")
	ref := c.GetHeader("ref")
	qType := c.PostForm("q_type")
	qDetail := c.PostForm("q_detail")
	qNum := c.PostForm("q_num")

	dateTime := time.Now().Format("2006-01-02 15:04:05")

	db := database.DB

	var eliOrder model.EliOrder

	eliOrder.OrderNo, _ = util.Sf.GenerateStrID()
	eliOrder.OrderType = 1
	eliOrder.QType = qType
	eliOrder.QDetail = qDetail

	myInt64, err := strconv.ParseInt(qNum, 10, 64)

	if err != nil {
		c.JSON(http.StatusOK, ml.Fail(lang, "100021"))
		return
	}

	eliOrder.QNum = myInt64
	eliOrder.CreateAt = time.Now()
	// eliOrder.UserID =
	eliOrder.Ref = ref

	// 解析时间字符串
	t, err := time.Parse("2006-01-02 15:04:05", dateTime)
	if err != nil {
		c.JSON(http.StatusOK, ml.Fail(lang, "100011"))
		return
	}

	sizhu := chronos.New(t).Lunar().EightCharacter()

	// 计算四柱八字
	yearGanZhi := sizhu[0] + sizhu[1]
	monthGanZhi := sizhu[2] + sizhu[3]
	dayGanZhi := sizhu[4] + sizhu[5]
	hourGanZhi := sizhu[6] + sizhu[7]

	sz := map[string]interface{}{"year_pillar": yearGanZhi, "month_pillar": monthGanZhi, "day_pillar": dayGanZhi, "hour_pillar": hourGanZhi}

	jsonBytes, err := json.Marshal(sz)
	if err != nil {
		log.Fatalf("JSON encoding failed: %s", err)
	}

	eliOrder.Sizhu = string(jsonBytes)

	db.Create(&eliOrder)

	c.JSON(http.StatusOK, ml.Succ(lang, map[string]interface{}{"order_no": eliOrder.OrderNo, "year_pillar": yearGanZhi, "month_pillar": monthGanZhi,
		"day_pillar": dayGanZhi, "hour_pillar": hourGanZhi}))
}

func CalculateShenKeOrder(c *gin.Context) {
	lang := c.GetHeader("I18n-Language")
	dizhi := c.PostForm("dizhi")
	kongwang := c.PostForm("kongwang") //贵神、人元
	rumu := c.PostForm("rumu")         //神将

	typesStr := c.PostForm("type")
	randomNumStr := c.PostForm("rand")
	orderNo := c.PostForm("order_no")

	types, err := strconv.ParseInt(typesStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, ml.Fail(lang, "100006"))
		return
	}

	token := c.GetHeader("token")
	session, _ := util.ParseToken(token)

	qtypeStr := c.PostForm("qtype")
	qtype, err := strconv.ParseInt(qtypeStr, 10, 64)
	if err != nil {
		// 处理错误
		qtype = 0
	}

	db := database.DB

	var sysCat model.SysCatalog
	if qtype > 0 {
		db.Model(&model.SysCatalog{}).Where("id=?", qtype).First(&sysCat)
	}

	//新增参数 空亡   空亡概念 ，没准备好，没有，心里没底，还没发生
	//新增参数 入墓 入墓象义  发挥不出来，不能动，控制，受限制
	//新增参数时间四柱 calculateFourPillars 的角色 对应 年柱代表长辈，领导，国家；月柱代表兄弟姐妹，竞争对手，朋友，同事；日柱代表关系近的朋友，配偶，自己；时柱代表子女，晚辈，下属

	//训练数据要调整 把数字转成 支 再转成对应的 干

	//拿地支算五行
	dizhis := strings.Split(dizhi, "")

	wuxings := myform.CalElements(dizhis)
	// wangshuais := wangxiang(wuxings)
	wangshuais := myform.CalProsDec(wuxings)

	// 定义查询出来的基础象意数据
	var eliSwxys []model.EliSwxy
	var eliSwwxs []model.EliSwwx
	var eliWxws []model.EliWxws
	var eliSwfls []model.EliSwfl
	var eliDzgxs []model.EliDzgx

	// 0 人元 1 贵神 2 神将 3 地分
	//计算出对应的五行 再计算出来对应的生克 再去表里查询出对应的关系记录
	for i := 0; i < len(wuxings); i++ {
		for j := i + 1; j < len(wuxings); j++ {
			relationship, sx := myform.CalculateWuxingRelationship(wuxings[i], wuxings[j])
			fmt.Printf("%s 和 %s 的关系是：%s\n", wuxings[i], wuxings[j], relationship)
			var tmpEliSwxys []model.EliSwxy
			sxVal_i := getSxValue(i)
			sxVal_j := getSxValue(j)

			if sx == -1 {
				continue
			}

			sql := "r1 = ? and r2 = ? and relationship = ? and flag = ?"
			params := make([]interface{}, 0)

			if sx == 1 {
				params = append(params, sxVal_i)
				params = append(params, sxVal_j)
				// db.Where(sql, sxVal_i, sxVal_j, relationship, 0).Find(&tmpEliSwxys)
			} else {
				params = append(params, sxVal_j)
				params = append(params, sxVal_i)
				// db.Where(sql, sxVal_j, sxVal_i, relationship, 0).Find(&tmpEliSwxys)
			}
			params = append(params, relationship)
			params = append(params, 0)
			if sysCat.ID > 0 {
				sql = sql + " and type=?"
				params = append(params, sysCat.NameCn)
			}
			db.Where(sql, params...).Find(&tmpEliSwxys)
			eliSwxys = append(eliSwxys, tmpEliSwxys...)
		}
	}

	//根据四位五行查询断语
	swwxs := countFiveElements(wuxings)
	for key, value := range swwxs {

		var eliSwwx model.EliSwwx
		db.Where(" wuxing = ? and num = ? ", key, value).First(&eliSwwx)

		if eliSwwx != (model.EliSwwx{}) {
			eliSwwxs = append(eliSwwxs, eliSwwx)
		}
	}

	// 根据旺衰查询断语
	added := make(map[int]bool) // 假设ID字段是int类型
	fmt.Println("五行旺衰:", wuxings, wangshuais)
	for i, _ := range wangshuais {
		if wangshuais[i] == "旺" || wangshuais[i] == "相" {
			wangshuais[i] = "旺"
		} else {
			wangshuais[i] = "衰"
		}
	}
	for i := 0; i < len(wangshuais); i++ {
		var eliWxw model.EliWxws
		db.Where("wuxing = ? and type = ?", wuxings[i], wangshuais[i]).First(&eliWxw)
		if !added[int(eliWxw.ID)] { // 检查这个ID是否已经添加
			eliWxws = append(eliWxws, eliWxw)
			added[int(eliWxw.ID)] = true // 标记这个ID已经添加
		}
	}

	//查询出四位对应的角色
	sql := "type = ?"
	flParams := make([]interface{}, 0)
	flParams = append(flParams, "0")
	if sysCat.ID > 0 {
		sql = sql + " or type = ?"
		flParams = append(flParams, sysCat.NameCn)
	}
	log.Println("查找四位分类的sql:", sql)
	db.Model(&model.EliSwfl{}).Where(sql, flParams...).Find(&eliSwfls)

	//用神判定 传入干支 返回 对应的索引 -1 就是失败,小于等于3就是对应人元、贵神、神将、地分
	yongshens := []string{"人元", "贵神", "神将", "地分"}

	yongshenIndex := yongshen(dizhi)

	if yongshenIndex == -1 {
		c.JSON(http.StatusOK, ml.Fail(lang, "100013"))
		return
	}

	shen := yongshens[yongshenIndex]

	_2combos_ := myform.UniqueCombination(2, dizhis, true)
	_3combos_ := myform.UniqueCombination(3, dizhis, true)

	_totalcombos_ := append(_2combos_, _3combos_...)

	var params []interface{}
	xchzSql := "dz in ("
	for _, v := range _totalcombos_ {
		xchzSql += "?,"
		params = append(params, v)
	}

	xchzSql = xchzSql[:(len(xchzSql)-1)] + ")"
	err = db.Model(&model.EliDzgx{}).Where(xchzSql, params...).Find(&eliDzgxs).Error
	fmt.Println(err)

	fmt.Println(eliSwxys, eliDzgxs, eliSwwxs, eliWxws)

	//获取问题分类
	var anstruct []model.SysCatStruct
	var ansangle []chat.JoinQuesAngl
	db.Model(&model.SysCatStruct{}).Where("cat_id=? and flag=?", qtype, 0).Order("seq asc").Find(&anstruct)
	if len(anstruct) > 0 {
		for _, v := range anstruct {
			ansangle = append(ansangle, chat.JoinQuesAngl{
				Cntstruct: v.Cntstruct,
			})
		}
	}

	result := map[string]interface{}{"swxys": eliSwxys, "eliDzgxs": eliDzgxs, "eliSwfls": eliSwfls,
		"eliSwwxs": eliSwwxs, "eliWxws": eliWxws, "kongwang": kongwang, "shen": shen, "rumu": rumu, "struct": ansangle}

	jsonData, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error occurred during marshaling. Error: %s", err.Error())
	}

	var order model.EliOrder
	db.Where(" order_no = ? ", orderNo).First(&order)

	//保存结果 在之后询问的时候用到
	eliCalInfo := new(model.EliCalInfo)

	// eliCalInfo.ID, _ = util.Sf.GenerateID()
	eliCalInfo.Param = dizhi
	eliCalInfo.Wuxing = strings.Join(wuxings, "")
	eliCalInfo.Wangshuai = strings.Join(wangshuais, "")
	eliCalInfo.Yongyao = shen

	eliCalInfo.Result = string(jsonData)
	eliCalInfo.Type = int32(types)
	eliCalInfo.RandNum = randomNumStr

	eliCalInfo.UserID = session.Id
	eliCalInfo.CreateAt = time.Now()
	eliCalInfo.OrderID = order.ID

	db.Create(&eliCalInfo)

	// "eliSwxys": eliSwxys, "eliDzgxs": eliDzgxs, "eliSwwxs": eliSwwxs, "eliWxws": eliWxws, "eliSwfls": eliSwfls
	c.JSON(http.StatusOK, ml.Succ(lang, map[string]interface{}{"cal_id": eliCalInfo.ID}))
}

// 微信支付
func WxPay(c *gin.Context) {
	lang := c.GetHeader("I18n-Language")
}
