package unigo

import (
	"time"

	"gorm.io/gorm"
)

type MachineNodeMgr struct {
	db *gorm.DB
}

func NewMachineNodeMgr(db *gorm.DB) *MachineNodeMgr {
	return &MachineNodeMgr{
		db: db,
	}
}

type MachineNode struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	HostName  string    `gorm:"column:host_name"`
	Port      int       `gorm:"column:port"`
	Type      int       `gorm:"column:type"`
	StartedAt time.Time `gorm:"column:started_at"`
}

func (this *MachineNode) TableName() string {
	return "machine_node"
}

// 添加记录，实例启动后，往数据库加一条记录，以记录本实例启动信息，并且获取一个机器（实例）id
func (w *MachineNodeMgr) GetNodeID(env int, hostname string, port int) (int64, error) {
	workerNode := MachineNode{
		HostName:  hostname,
		Port:      port,
		Type:      env,
		StartedAt: time.Now(),
	}

	result := w.db.Create(&workerNode)
	if result.Error != nil {
		return 0, result.Error
	}

	return workerNode.ID, nil
}
