package util

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type SessionToken struct {
	Id    int64  `json:"id"`
	Email string `json:"email"`
	Addr  string `json:"addr"`
	Type  int8   `json:"type"`
	Name  string `json:"name"`
}

type AdminSessionToken struct {
	Id      int64  `json:"id"`
	UserId  int64  `json:"userId"`
	Email   string `json:"email"`
	Account string `json:"addr"`
	Type    int8   `json:"type"` //1.vg 2.mcn
	OrgId   int64  `json:"orgId"`
}

func Macke(session *SessionToken) (token string, err error) { //生成jwt
	claims := jwt.MapClaims{ //创建一个自己的声明
		"id":    session.Id,
		"email": session.Email,
		"addr":  session.Addr,
		"type":  session.Type,
		"name":  session.Name,
		"iss":   "vocosmos",
		"nbf":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour * 672).Unix(),
		"iat":   time.Now().Unix(),
	}

	then := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token, err = then.SignedString([]byte("gettoken"))

	return
}

func AdminMacke(session *AdminSessionToken) (token string, err error) { //生成jwt
	claims := jwt.MapClaims{ //创建一个自己的声明
		"id":      session.Id,
		"userId":  session.UserId,
		"email":   session.Email,
		"account": session.Account,
		"type":    session.Type,
		"orgId":   session.OrgId,
		"iss":     "adminVega",
		"nbf":     time.Now().Unix(),
		"exp":     time.Now().Add(time.Hour * 168).Unix(),
		"iat":     time.Now().Unix(),
	}

	then := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	fmt.Println(then) //打印&{ 0xc0000040a8 map[alg:HS256 typ:JWT] map[exp:1637212218 iat:1637212214 iss:lvjianhua name:zhansan nbf:1637212214 pwd:pwd]  false}

	token, err = then.SignedString([]byte("gettoken"))

	return
}

func secret() jwt.Keyfunc { //按照这样的规则解析
	return func(t *jwt.Token) (interface{}, error) {
		return []byte("gettoken"), nil
	}
}

// 解析token
func ParseToken(token string) (session *SessionToken, err error) {
	session = &SessionToken{}
	tokn, err := jwt.Parse(token, secret())

	if err != nil {
		err = errors.New("解析错误")
		return
	}

	claim, ok := tokn.Claims.(jwt.MapClaims)
	if !ok {
		err = errors.New("解析错误")
		return
	}
	if !tokn.Valid {
		err = errors.New("令牌错误！")
		return
	}

	session.Id = int64(claim["id"].(float64))
	session.Email = claim["email"].(string) //强行转换为string类型 这个两个值可能都为空
	session.Addr = claim["addr"].(string)   //强行转换为string类型 这个两个值可能都为空
	session.Type = int8(claim["type"].(float64))
	session.Name = claim["name"].(string)
	return
}

// 解析token
func ParseAdminToken(token string) (session *AdminSessionToken, err error) {
	session = &AdminSessionToken{}
	tokn, err := jwt.Parse(token, secret())

	if err != nil {
		err = errors.New("解析错误")
		return
	}

	claim, ok := tokn.Claims.(jwt.MapClaims)

	if !ok {
		err = errors.New("解析错误")
		return
	}

	if !tokn.Valid {
		err = errors.New("令牌错误！")
		return
	}

	session.Id = int64(claim["id"].(float64))
	session.UserId = int64(claim["userId"].(float64))
	session.Email = claim["email"].(string)     //强行转换为string类型 这个两个值可能都为空
	session.Account = claim["account"].(string) //强行转换为string类型 这个两个值可能都为空
	session.Type = int8(claim["type"].(float64))
	session.OrgId = int64(claim["orgId"].(float64))
	return
}

type AuthInfoClaims struct {
	Name               string `json:"name"`
	Avatar             string `json:"avatar"`
	MusicianId         int64  `json:"musician_id"`
	Twitter_Account    string `json:"twitt_account"`
	Twitter_Account_id string `json:"twitt_account_id"`
	jwt.StandardClaims
}

func GenerateAuthToken(name, avatar, twitterAccount string, twitterAccountId string, musicianId int64, expirationHours int, secretKey string) (string, error) {
	// 设置JWT的过期时间
	expirationTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)

	// 创建自定义声明
	claims := &AuthInfoClaims{
		Name:               name,
		Avatar:             avatar,
		MusicianId:         musicianId,
		Twitter_Account:    twitterAccount,
		Twitter_Account_id: twitterAccountId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// 创建JWT令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名JWT并获取字符串表示
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析JWT令牌
func ParseAuthToken(tokenString, secretKey string) (*AuthInfoClaims, error) {
	// 解析JWT令牌
	parsedToken, err := jwt.ParseWithClaims(tokenString, &AuthInfoClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// 检查JWT是否有效
	if parsedToken.Valid {
		if claims, ok := parsedToken.Claims.(*AuthInfoClaims); ok {
			return claims, nil
		}
	}

	return nil, fmt.Errorf("无效的JWT令牌")
}

var emailBlacklist = []string{
	"nezid.com", "vlook.cloud", "Zslsz.com", "omeie.com", "zbock.com",
	"tyincoming.com", "qdproceedsp.com", "kohelps.com", "qperformsrx.com",
	"dminutesfb.com", "altmails.com", "altmails.com", "bbitj.com",
	"xcclectures.com", "bbitq.com", "mailnesia.com", "mikobeaute.com",
}

var ipBlacklist = []string{"106.105.64.179", "123.193.208.145", "49.216.222.10", "124.218.55.129",
	"116.241.161.254", "122.99.38.185", "219.68.214.161", "101.9.55.36",
	"211.76.67.78", "123.194.129.98", "27.242.160.203", "123.194.13.178",
	"123.252.78.168", "101.10.44.163", "36.231.174.24", "106.105.89.139",
	"27.53.145.114", "111.83.229.181", "220.137.99.13", "42.76.118.230",
	"101.10.102.168", "111.83.213.69", "101.10.58.192", "110.30.24.163",
	"114.137.228.85", "49.216.222.6"}

// 检测IP和邮箱是否在黑名单中
func IsBlacklisted(email, ip string) (emailBlacklisted, ipBlacklisted bool) {
	// 将输入的邮箱地址转换为小写
	email = strings.ToLower(email)

	// 提取邮箱后缀
	emailSuffix := strings.Split(email, "@")[1]

	for _, e := range emailBlacklist {
		if e == emailSuffix {
			emailBlacklisted = true
			break
		}
	}

	for _, i := range ipBlacklist {
		if i == ip {
			ipBlacklisted = true
			break
		}
	}

	return
}

type BindInfoClaims struct {
	Id         int64  `json:"id"`
	LineId     string `json:"line_id"`
	MusicianId int64  `json:"musician_id"`
	jwt.StandardClaims
}

func GenerateBindAuthToken(lineId string, id, musicianId int64, expirationHours int, secretKey string) (string, error) {

	// 设置JWT的过期时间
	expirationTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)

	// 创建自定义声明
	claims := &BindInfoClaims{
		Id:         id,
		LineId:     lineId,
		MusicianId: musicianId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// 创建JWT令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名JWT并获取字符串表示
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseBindAuthToken 解析JWT令牌
func ParseBindAuthToken(tokenString, secretKey string) (*BindInfoClaims, error) {
	// 解析JWT令牌
	parsedToken, err := jwt.ParseWithClaims(tokenString, &BindInfoClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// 检查JWT是否有效
	if parsedToken.Valid {
		if claims, ok := parsedToken.Claims.(*BindInfoClaims); ok {
			return claims, nil
		}
	}

	return nil, fmt.Errorf("无效的JWT令牌")
}
