package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"github.com/iautre/gowk"
	"github.com/iautre/gowk/auth/db"
	"github.com/jackc/pgx/v5/pgtype"
	"strconv"
	"time"
)

type AppService struct {
}

func (a *AppService) GetByKey(ctx context.Context, key string) (db.App, error) {
	return db.New(gowk.DB(ctx)).AppByKey(ctx, pgtype.Text{String: key})
}

type UserService struct {
}

func (u *UserService) GetById(ctx context.Context, id int64) (db.User, error) {
	return db.New(gowk.DB(ctx)).UserById(ctx, id)
}

func (u *UserService) GetByPhone(ctx context.Context, photo string) (db.User, error) {
	return db.New(gowk.DB(ctx)).UserByPhone(ctx, pgtype.Text{String: photo})
}

// Login 登录
func (u *UserService) Login(ctx context.Context, params *LoginParams) (db.User, error) {
	user, err := u.GetByPhone(ctx, params.Account)
	if err != nil {
		return db.User{}, err
	}
	// 校验code和account
	var otp OTP
	if !otp.CheckCode(user.Secret.String, params.Code) {
		return db.User{}, gowk.NewError("验证码错误")
	}
	return user, nil
}

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
	return (uint32(bytes[0]) << 24) + (uint32(bytes[1]) << 16) +
		(uint32(bytes[2]) << 8) + uint32(bytes[3])
}
