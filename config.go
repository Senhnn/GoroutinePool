package GoroutinePool

const (
	defaultScalaThreshold = 1
)

// Config 配置goroutine pool
type Config struct {
	// 阈值：只有任务队列大于阈值时，才能创建新的goroutine
	ScaleThreshold int32
}

// NewConfig 获取新配置
func NewConfig() *Config {
	c := &Config{
		ScaleThreshold: defaultScalaThreshold,
	}
	return c
}
