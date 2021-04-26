package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"sync"
)


//user新表
type AuthUserCP struct {
	gorm.Model
	AuthUser `gorm:"embedded"`
}
//user旧表
type AuthUser struct {
	Email string
	Username string
	Password string
	IsEmailActivated bool
	RegisterTimestamp int64
	LastLoginTimestamp int64
	RegisterIp string
	LastLoginIp string
	Role string
	IsIncognito bool //用来表示用户的状态，false表示正常状态，true表示在incognito状态
}
//数据库信息
type Dbconfig struct {
	Username string
	Password string
	Host string
	Port int
	DbName string
}
//远程服务器配置与复制到数据库配置
type Conf struct {
	DbOrigin Dbconfig
	DbCP Dbconfig
}

var (
	dbOrigin *gorm.DB //原数据库
	dbCP *gorm.DB //复制数据库
	conf *Conf //数据库信息
	once sync.Once //单例模式

	counter int64 //数据库行数
)
//读取配置文件
func Config() *Conf {
	once.Do(func(){
		filePath:="UserDBMigrate/db.toml"
		if _,err:=toml.DecodeFile(filePath,&conf);err!=nil{
			panic(err)
		}
	})
	return conf
}

//数据库配置转为dsn
func DBConfToDsn(conf Dbconfig) string{
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", conf.Username, conf.Password, conf.Host, conf.Port, conf.DbName)
}

func init()  {
	var err error
	//连接本地数据库
	dbCP,err=gorm.Open("mysql",DBConfToDsn(Config().DbCP))
	if err!=nil{
		panic("连接数据库失败"+err.Error())
	}
	dbCP.AutoMigrate(&AuthUserCP{})
	//连接远程数据库
	dbOrigin,err=gorm.Open("mysql",DBConfToDsn(Config().DbOrigin))
	if err!=nil{
		panic("连接远程数据库失败"+err.Error())
	}
}


func main() {
	for {
		bufUser :=AuthUserCP{} //读取数据库row缓存
		dbOrigin.Limit(1).Offset(counter).Find(&bufUser.AuthUser)

		fmt.Println("=======================")
		fmt.Println(bufUser)
		if bufUser.Email==""{
			break
		}
		//sqlCreateAndCreate()
		dbCP.Debug().Create(&bufUser)
		fmt.Println(bufUser.ID)
		fmt.Println("==========================")
		counter++
	}
}

