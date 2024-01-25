package interceptor

import (
	"bytes"
	ml "eli/middleware"
	"eli/util"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

const SIGN_SALT = "202cb962ac5907sdfdsdb07152d234b70"

var userLoginException = []string{"/api/v1/user/sign_msg", "/api/v1/user/login", "/api/v1/user/handle_google_callback", "/api/v1/eli/*", "/api/v1/ws/*", "/api/v1/dis/*"}
var signException = []string{}

func FrontendSignMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("I18n-Language")
		sign := c.GetHeader("sign")

		params := extractRequestParameters(c)
		result := checkSign(params, sign, c.FullPath())

		if !result {
			c.JSON(http.StatusOK, ml.Fail(lang, "100003"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func FrontendAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取语言和令牌
		lang := c.GetHeader("I18n-Language")
		token := c.GetHeader("token")

		// 检查是否需要跳过验证
		if shouldSkipAuth(c.FullPath()) {
			c.Next()
			return
		}

		// 如果令牌为空，则返回未授权响应
		if token == "" {
			c.JSON(http.StatusUnauthorized, ml.Fail(lang, "100004"))
			c.Abort()
			return
		}

		// 验证令牌
		_, err := util.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, ml.Fail(lang, "100005"))
			c.Abort()
			return
		}

		// 继续处理请求
		c.Next()
	}
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在这里添加后台身份验证逻辑
		// 例如，检查用户是否已登录或具有后台管理权限
		// 如果验证失败，可以返回错误或重定向到登录页面
		c.Next()
	}
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端IP
		ip := c.ClientIP()

		ts := c.GetHeader("ts")
		sign := c.GetHeader("sign")

		// 获取设备信息
		userAgent := c.GetHeader("User-Agent")
		device := "PC"
		if strings.Contains(strings.ToLower(userAgent), "mobile") {
			device = "Mobile"
		}

		params := extractRequestParameters(c)

		jsonParams, err := json.Marshal(params)
		if err != nil {
			fmt.Println("JSON 编码错误:", err)
			return
		}

		// 处理请求前打印输入
		ml.Log.Infof("Method=%s, IP=%s, Device=%s, Path=%s,Header:=ts=%s&sign=%s Params=%s ",
			c.Request.Method, ip, device, c.FullPath(), ts, sign, string(jsonParams))

		c.Next()

		//responseBody := string(crw.Body)
		// 处理响应后打印输出
		//ml.Log.Infof("Response >>>>>>>> Status Code=%d Response Body=%s \n", c.Writer.Status(), responseBody)

	}
}

// 判断是否需要跳过验证
func shouldSkipAuth(path string) bool {
	for i := 0; i < len(userLoginException); i++ {
		if strings.Contains(userLoginException[i], "*") {
			if strings.HasPrefix(path, strings.ReplaceAll(userLoginException[i], "*", "")) {
				return true
			}
		} else {
			if strings.EqualFold(userLoginException[i], path) {
				return true
			}
		}
	}
	return false
}

// CustomResponseWriter 是自定义的响应写入器，它包装了原始的 gin.ResponseWriter
type CustomResponseWriter struct {
	gin.ResponseWriter
	Body []byte
}

// Write 重写 Write 方法，以便捕获响应主体
func (w *CustomResponseWriter) Write(data []byte) (int, error) {
	w.Body = data
	return w.ResponseWriter.Write(data)
}

func extractRequestParameters(c *gin.Context) map[string]string {
	params := make(map[string]string)

	// 读取请求体并处理表单参数
	body, err := io.ReadAll(c.Request.Body)
	if err == nil {
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		formParams := strings.Split(string(body), "&")
		for _, formParam := range formParams {
			parts := strings.Split(formParam, "=")
			if len(parts) > 1 {
				params[parts[0]] = parts[1]
			} else if len(parts) == 1 && parts[0] != "" {
				params[parts[0]] = ""
			}
		}
	}

	// 添加请求头中的参数
	params["ts"] = c.GetHeader("ts")

	// 处理查询参数
	queryParams := c.Request.URL.Query()
	for key, values := range queryParams {
		params[key] = strings.Join(values, "")
	}

	return params
}

func checkSign(params map[string]string, token string, fullPath string) bool {
	// 检查路径是否在签名例外列表中
	if isPathInSignException(fullPath) {
		return true
	}

	// 检查是否存在有效的令牌
	if token == "" {
		return false
	}

	// 如果令牌是特定值（后门），允许通过
	if strings.EqualFold("5Z4zWZO02eapg9igCUtwk5Z4zWZpg9igCUtwk5Z4zWZO02eapg9igCUtwk", token) {
		return true
	}

	// 计算参数的签名并检查是否匹配提供的令牌
	calculatedToken := MD5Params(params)
	return strings.EqualFold(calculatedToken, token)
}

func isPathInSignException(path string) bool {
	for _, exception := range signException {
		if strings.Contains(exception, "*") {
			if strings.HasPrefix(path, strings.ReplaceAll(exception, "*", "")) {
				return true
			}
		} else {
			if strings.EqualFold(exception, path) {
				return true
			}
		}
	}
	return false
}

// 取所有query， formPost以及header的ts 的参数，把所有参数按照正序排列，拼成字符串 + 盐值(202cb962ac59075b964b07152d234b70) 然后md5 组成字符串token 放在header
func MD5Params(params map[string]string) string {
	// 提取参数的键并按字母顺序排序
	paramKeys := make([]string, 0, len(params))
	for key := range params {
		paramKeys = append(paramKeys, key)
	}
	sort.Strings(paramKeys)

	// 构建原始字符串
	originStr := ""
	for _, key := range paramKeys {
		originStr += fmt.Sprintf("%s=%s", key, params[key])
	}

	originStr += fmt.Sprintf("key=%s", SIGN_SALT)

	originStr = url.QueryEscape(originStr)

	sign := util.ToMd5(originStr)
	return sign
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
