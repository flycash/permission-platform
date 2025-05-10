package domain

// BusinessConfig 业务配置领域模型
type BusinessConfig struct {
	ID        int64  // 业务ID
	OwnerID   int64  // 业务方ID
	OwnerType string // 业务方类型
	Name      string // 业务名称
	RateLimit int    // 每秒最大请求数
	Token     string // 业务方Token，内部包含uid也就是上方的ownerID
	Ctime     int64
	Utime     int64
}
