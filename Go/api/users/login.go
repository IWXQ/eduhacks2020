package users

import (
	"context"
	"eduhacks2020/Go/api"
	"eduhacks2020/Go/crypto"
	"eduhacks2020/Go/models/psql"
	"eduhacks2020/Go/models/response"
	"eduhacks2020/Go/utils"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

const (
	passwordValid = "password is mismatch"
	userBanned    = "account blocked"
	unknownLogin  = "unknown login field"
	adminGetError = "an error occurred while getting the configuration from redis"
)

// LoginParam 登录使用的参数
type LoginParam struct {
	Username string `json:"user"`
	Password string `json:"password"`
	Type     int    `json:"type"`
	Salt     string `json:"salt"`
}

// LoginResponse 登录后的结果
type LoginResponse struct {
	Token string      `json:"token"`
	Data  interface{} `json:"data"`
	Time  time.Time   `json:"time"`
}

// Exec 执行登录
func (l *LoginParam) Exec(db *gorm.DB, redis *redis.Client, sessionId string) ([]byte, string, error) {

	//var result interface{}
	switch l.Type {
	case -1:
		return l.adminLogin(redis)
	case 1:
		return l.teacherLogin(db, redis, sessionId)
	case 2:
		return l.studentLogin(db, redis, sessionId)
	default:
		return nil, "未知的登录域", errors.New(unknownLogin)
	}

}

// adminLogin 管理员的登录
func (l *LoginParam) adminLogin(redis *redis.Client) ([]byte, string, error) {
	Username, err := redis.Get(context.Background(), "AdminUser").Result()
	Password, err := redis.Get(context.Background(), "AdminPassword").Result()
	if err != nil {
		log.Errorf("an error occurred while getting the admin configure from redis: %s", err.Error())
		return nil, "an error occurred while logging admin", errors.New(adminGetError)
	}
	if l.Username != Username || l.Password != Password {
		return nil, "username or password is invalid", errors.New(passwordValid)
	}
	userFlag := utils.GenUUIDv5(Username)
	claims := utils.CustomClaims{
		UID:      "Admin",
		Name:     "Admin",
		Username: l.Username,
		Phone:    "",
		Role:     -1,
		Flag:     userFlag,
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix() - 1000, // 签名生效时间
			ExpiresAt: time.Now().Unix() + 3600, // 过期时间 一小时
			Issuer:    utils.Issuer,             //签名的发行者
		},
	}
	token, err := utils.GenerateToken(claims)
	if err != nil {
		return nil, err.Error(), err
	}
	res := LoginResponse{
		Token: token,
		Data:  nil,
		Time:  time.Now(),
	}
	js, err := json.Marshal(&res)
	if err != nil {
		return nil, err.Error(), err
	}
	redisAuth := api.AuthRedis{Redis: redis}
	err = redisAuth.SetFlag(claims.UID, userFlag)
	if err != nil {
		log.Error(err.Error())
	}
	return js, "Login success !", nil
}

// managerLogin 教务的登录
func (l *LoginParam) managerLogin(db *gorm.DB, redis *redis.Client, sessionId string) ([]byte, string, error) {
	return nil, "", nil
}

// teacherLogin 教师的登录
func (l *LoginParam) teacherLogin(db *gorm.DB, redis *redis.Client, sessionId string) ([]byte, string, error) {
	cipher := crypto.ChaCha20Poly1305{}
	cipher.Init()
	result := psql.Teacher{}
	row := db.Where("username = ?", l.Username).Or("phone = ?", l.Username).Or("email = ?", l.Username).First(&result)
	if row.Error != nil {
		if row.Error == gorm.ErrRecordNotFound {
			return nil, "username or password is invalid", row.Error
		}
		return nil, "unknown error: " + row.Error.Error(), row.Error
	}
	restorePwd, _ := cipher.DecryptedFromHex(result.Password)
	if string(restorePwd) != l.Password {
		return nil, "username or password is invalid", errors.New(passwordValid)
	}

	//generate token
	userFlag := utils.GenUUIDv5(result.UserID.String())
	claims := utils.CustomClaims{
		UID:      result.UserID.String(),
		Name:     result.Nickname,
		Username: result.Username,
		Phone:    result.Phone,
		Role:     1,
		Flag:     userFlag,
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix() - 1000, // 签名生效时间
			ExpiresAt: time.Now().Unix() + 3600, // 过期时间 一小时
			Issuer:    utils.Issuer,             //签名的发行者
		},
	}
	token, err := utils.GenerateToken(claims)
	if err != nil {
		return nil, err.Error(), err
	}
	teacherInfo := response.TeacherInfo{}
	db.First(&teacherInfo, "username = ?", result.Username)
	res := LoginResponse{
		Token: token,
		Data:  teacherInfo,
		Time:  time.Now(),
	}
	js, err := json.Marshal(&res)
	if err != nil {
		return nil, err.Error(), err
	}
	redisAuth := api.AuthRedis{Redis: redis}
	err = redisAuth.SetFlag(claims.UID, userFlag)
	if err != nil {
		log.Error(err.Error())
	}
	return js, "Login success !", nil
}

// studentLogin 学生的登录
func (l *LoginParam) studentLogin(db *gorm.DB, redis *redis.Client, sessionId string) ([]byte, string, error) {
	cipher := crypto.ChaCha20Poly1305{}
	cipher.Init()
	result := psql.Student{}
	row := db.Where("username = ?", l.Username).Or("phone = ?", l.Username).Or("email = ?", l.Username).First(&result)
	if row.Error != nil {
		if row.Error == gorm.ErrRecordNotFound {
			return nil, "username or password is invalid", row.Error
		}
		return nil, "unknown error: " + row.Error.Error(), row.Error
	}
	restorePwd, _ := cipher.DecryptedFromHex(result.Password)
	if string(restorePwd) != l.Password {
		return nil, "username or password is invalid", errors.New(passwordValid)
	}
	if result.Banned {
		return nil, "this account is banned", errors.New(userBanned)
	}
	//generate token
	userFlag := utils.GenUUIDv5(result.UserID.String())
	claims := utils.CustomClaims{
		UID:      result.UserID.String(),
		Name:     result.Nickname,
		Username: result.Username,
		Phone:    result.Phone,
		Role:     2,
		Flag:     userFlag,
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix() - 1000, // 签名生效时间
			ExpiresAt: time.Now().Unix() + 3600, // 过期时间 一小时
			Issuer:    utils.Issuer,             //签名的发行者
		},
	}
	token, err := utils.GenerateToken(claims)
	if err != nil {
		return nil, err.Error(), err
	}
	studentInfo := response.StudentInfo{}
	db.Model(&psql.Student{}).Select("student.users.*,college.classes.class_name,college.classes.class_id,college.majors.major_name,college.majors.major_id,college.colleges.college_name,college.colleges.college_id").
		Joins("left join college.classes on student.users.class_id = college.classes.id left join college.majors on college.classes.major_id = college.majors.id LEFT JOIN college.colleges on college.majors.college_id = college.colleges.id").
		Where("student.users.username = ?", result.Username).Scan(&studentInfo)
	res := LoginResponse{
		Token: token,
		Data:  studentInfo,
		Time:  time.Now(),
	}
	js, err := json.Marshal(&res)
	if err != nil {
		return nil, err.Error(), err
	}
	//session := database.SessionManager{Values: make(map[interface{}]interface{})}
	//session.Values["login"] = true
	//cipherText, err := session.EncryptedData(database.SessionName)
	//if err != nil {
	//	log.Errorf("Encrypted an error has occurred %s", err.Error())
	//}
	//session.SaveData(id, cipherText)
	redisAuth := api.AuthRedis{Redis: redis}
	err = redisAuth.SetFlag(claims.UID, userFlag)
	if err != nil {
		log.Error(err.Error())
	}
	return js, "Login success !", nil
}
