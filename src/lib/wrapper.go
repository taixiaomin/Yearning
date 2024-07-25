package lib

import (
	"Yearning-go/src/model"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/Jeffail/gabs/v2"
	"golang.org/x/crypto/pbkdf2"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ArrayRemove(source []byte, flag string) ([]byte, error) {
	p, err := gabs.ParseJSON(source)
	if err != nil {
		return nil, err
	}
	for i, c := range p.Children() {
		if c.Data().(string) == flag {
			_ = p.ArrayRemove(i)
		}
	}
	return p.EncodeJSON(), nil
}

func MultiArrayRemove(source []byte, sep []string, flag string) ([]byte, error) {
	p, err := gabs.ParseJSON(source)
	if err != nil {
		return nil, err
	}
	var wait sync.WaitGroup
	wait.Add(len(sep))
	for _, dl := range sep {
		go func(wait *sync.WaitGroup, dl string) {
			for i, c := range p.S(dl).Children() {
				if c.Data().(string) == flag {
					_ = p.ArrayRemove(i, dl)
				}
			}
			wait.Done()
		}(&wait, dl)
	}
	wait.Wait()
	return p.EncodeJSON(), nil
}

func GetRandom() []byte {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	destr := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 12; i++ {
		result = append(result, destr[r.Intn(len(destr))])
	}
	return result
}

func DjangoEncrypt(password string, sl string) string {
	pwd := []byte(password)
	salt := []byte(sl)
	iterations := 120000
	digest := sha256.New
	dk := pbkdf2.Key(pwd, salt, iterations, 32, digest)
	str := base64.StdEncoding.EncodeToString(dk)
	return "pbkdf2_sha256" + "$" + strconv.FormatInt(int64(iterations), 10) + "$" + string(salt) + "$" + str
}

func DjangoCheckPassword(account *model.CoreAccount, password string) bool {
	sl := strings.Split(account.Password, "$")[2]
	checkPasswordToken := DjangoEncrypt(password, sl)
	if account.Password == checkPasswordToken {
		return true
	} else {
		return false
	}
}

func hmacSha256(stringToSign string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
