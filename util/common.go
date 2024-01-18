package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

func FormatEthereumAddress(address string) string {
	if len(address) < 9 {
		return address // 地址太短，不进行省略
	}

	// 获取地址开头4个字符
	start := address[:4]

	// 获取地址末尾4个字符
	end := address[len(address)-4:]

	// 创建中间省略号部分，长度为4个字符的省略号
	ellipsis := strings.Repeat(".", 4)

	// 将开头、省略号和末尾拼接在一起
	formattedAddress := fmt.Sprintf("%s%s%s", start, ellipsis, end)

	return formattedAddress
}

var salt = "mFyYzfA1234*"

func ToMd5(str string) string {

	data := []byte(str) //切片
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has) //将[]byte转成16进制
	return md5str
}

func ToMd5AndSalt(str string) string {

	data := []byte(str + salt) //切片
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has) //将[]byte转成16进制
	return md5str
}

func In(target string, str_array []string) bool {

	sort.Strings(str_array)

	index := sort.SearchStrings(str_array, target)

	if index < len(str_array) && str_array[index] == target {

		return true

	}

	return false

}

// CreateMutiDir 调用os.MkdirAll递归创建文件夹
func CreateMutiDir(filePath string) error {
	_, err := os.Stat(filePath) //os.Stat获取文件信息
	if err != nil && !os.IsExist(err) {
		err = os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			fmt.Println("创建文件夹失败,error info:", err)
			return err
		}
	}
	return nil
}

func GetUUid() string {
	b := make([]byte, 16)
	io.ReadFull(rand.Reader, b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x%x%x%x%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// 屏蔽掉邮箱中间的字符
func MaskEmail(email string) string {
	atIndex := strings.Index(email, "@")
	if atIndex == -1 {
		return email
	}

	username := email[:atIndex]
	domain := email[atIndex:]

	// 屏蔽邮箱中间5个字符
	maskedUsername := string(username[0]) + strings.Repeat("*", 5) + string(username[len(username)-1])

	return maskedUsername + domain
}

// 按照指定时区转化字符串中的时间
func ConvertTimeInTimeZone(timeStr, targetTimeZone string) (time.Time, error) {
	layout := "2006-01-02 15:04:05"

	// 加载目标时区
	targetLocation, err := time.LoadLocation(targetTimeZone)
	if err != nil {
		return time.Time{}, err
	}

	// 使用 time.ParseInLocation 解析时间字符串并指定目标时区
	parsedTime, err := time.ParseInLocation(layout, timeStr, targetLocation)
	if err != nil {
		return time.Time{}, err
	}

	return parsedTime, nil
}

// 生成指定长度的随机字符串
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// 判断数组中是否包含另一个元素
func Contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
