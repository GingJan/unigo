package unigo

import (
	"fmt"
	"math/rand"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MachineNodeDBConfig struct {
	UserName string
	Password string
	IP       string
	Port     string
	DBName   string
	Charset  string
	db       *gorm.DB
}

type MachineNode struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	HostName  string    `gorm:"column:host_name"`
	Port      string    `gorm:"column:port"`
	Type      int       `gorm:"column:type"`
	StartedAt time.Time `gorm:"column:started_at"`
}

func (this *MachineNode) TableName() string {
	return "machine_node"
}

func NewMachineNode(w *MachineNodeDBConfig) (int64, error) {
	if err := w.connectMysql(); err != nil {
		return 0, err
	}
	return w.addRecord()
}

// 连接数据
func (w *MachineNodeDBConfig) connectMysql() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local", w.UserName, w.Password, w.IP, w.Port, w.DBName, w.Charset)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	w.db = db
	return nil
}

// 添加记录，实例启动后，往数据库加一条记录，以记录本实例启动信息，并且获取一个机器（实例）id
func (w *MachineNodeDBConfig) addRecord() (int64, error) {
	hostname, environment := GetHostAndE()
	rand.Seed(time.Now().UnixNano())
	port := fmt.Sprintf("%v-%v", time.Now().UnixNano()/1e6, rand.Intn(100000))

	workerNode := MachineNode{
		HostName:  hostname,
		Port:      port,
		Type:      environment,
		StartedAt: time.Now(),
	}

	result := w.db.Create(&workerNode)
	if result.Error != nil {
		return 0, result.Error
	}

	return workerNode.ID, nil
}
