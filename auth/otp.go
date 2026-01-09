package auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"strconv"
	"time"
)

// 来自&感谢 https://piaohua.github.io/post/golang/20230527-google-authenticator/
type OTP struct{}

func (o *OTP) CheckCode(secret string, code string) bool {
	// 当前值
	if o.GetCode(secret, 0) == code {
		return true
	}
	// 往前10秒
	if o.GetCode(secret, -20) == code {
		return true
	}
	// 往后10秒
	if o.GetCode(secret, 20) == code {
		return true
	}
	return false
}

// 获取Code
func (o *OTP) GetCode(secret string, offset int64) string {
	key, _ := base32.StdEncoding.DecodeString(secret)
	epochSeconds := time.Now().Unix() + offset
	return strconv.FormatInt(int64(o.OneTimePassword(key, o.ToBytes(epochSeconds/30))), 10)
}

// 获取密码
func (o *OTP) OneTimePassword(key []byte, value []byte) uint32 {
	hmacSha1 := hmac.New(sha1.New, key)
	hmacSha1.Write(value)
	hash := hmacSha1.Sum(nil)
	offset := hash[len(hash)-1] & 0x0F
	hashParts := hash[offset : offset+4]
	hashParts[0] = hashParts[0] & 0x7F
	number := o.ToUint32(hashParts)
	return number % 1000000
}
func (o *OTP) ToBytes(value int64) []byte {
	var result []byte
	mask := int64(0xFF)
	shifts := [8]uint16{56, 48, 40, 32, 24, 16, 8, 0}
	for _, shift := range shifts {
		result = append(result, byte((value>>shift)&mask))
	}
	return result
}
func (o *OTP) ToUint32(bytes []byte) uint32 {
	return (uint32(bytes[0]) << 24) + (uint32(bytes[1]) << 16) + (uint32(bytes[2]) << 8) + uint32(bytes[3])
}
