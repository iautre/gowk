package gowk

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/go-basic/uuid"
)

func UUID() string {
	return strings.Replace(uuid.New(), "-", "", -1)
}

func UUID64() string {
	return UUID() + UUID()
}

func NewAuid() uint {
	return uint(time.Now().Unix())
}

func UUID36() string {
	rand.Seed(time.Now().UnixNano())
	ra := rand.Intn(900) + 100
	timesss := time.Now().UnixNano() / 1e6 // - 1622994898
	hai := fmt.Sprintf("%s%s", strconv.Itoa(ra), strconv.FormatInt(timesss, 10))
	aaa, _ := strconv.ParseInt(hai, 10, 64)
	t2, _ := time.Parse("2006-01-02 15:04:05", "2021-06-07 23:00:00")
	bb := t2.UnixNano() / 1e6

	return NumToBHex(aaa-bb, 36)
}

var num2char = "0123456789abcdefghijklmnopqrstuvwxyz"

func NumToBHex(num int64, n int64) string {
	num_str := ""
	for num != 0 {
		yu := num % n
		num_str = string(num2char[yu]) + num_str
		num = num / n
	}
	return strings.ToLower(num_str)
}
