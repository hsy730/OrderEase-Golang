package handlers

import (
	"io"
	"net/http"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
	"sync"

	"github.com/gin-gonic/gin"
)

// 定义一个简单的基于内存的广播器
type Broadcaster struct {
	clients map[chan models.Order]bool
	mutex   sync.RWMutex
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[chan models.Order]bool),
	}
}

// Subscribe 订阅SSE事件
func (b *Broadcaster) Subscribe() chan models.Order {
	ch := make(chan models.Order)
	b.mutex.Lock()
	b.clients[ch] = true
	b.mutex.Unlock()
	return ch
}

// Unsubscribe 取消订阅
func (b *Broadcaster) Unsubscribe(ch chan models.Order) {
	b.mutex.Lock()
	delete(b.clients, ch)
	close(ch)
	b.mutex.Unlock()
}

// Broadcast 广播订单事件
func (b *Broadcaster) Broadcast(order models.Order) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	
	for clientChan := range b.clients {
		select {
		case clientChan <- order:
		default:
			// 如果通道阻塞，跳过该客户端以避免阻塞其他客户端
			log2.Warnf("SSE client channel is blocked, skipping broadcast for order ID: %d", order.ID)
		}
	}
}

// 全局广播器实例
var GlobalOrderBroadcaster = NewBroadcaster()

// SSE连接处理函数
func (h *Handler) SSEConnection(c *gin.Context) {
	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 获取店铺ID验证权限
	requestShopID, err := utils.StringToSnowflakeID(c.Query("shop_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的店铺ID"})
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 订阅订单事件
	clientChan := GlobalOrderBroadcaster.Subscribe()
	defer GlobalOrderBroadcaster.Unsubscribe(clientChan)

	// 发送初始连接确认消息
	c.Stream(func(w io.Writer) bool {
		c.SSEvent("connected", gin.H{"message": "SSE connection established"})
		return false // 只发送一次
	})

	// 监听订单事件
	for {
		select {
		case order, ok := <-clientChan:
			if !ok {
				// 通道关闭，客户端断开连接
				return
			}
			
			// 检查订单是否属于当前店铺
			if order.ShopID == validShopID {
				// 推送订单事件给客户端
				c.Stream(func(w io.Writer) bool {
					c.SSEvent("new_order", order)
					return false
				})
				
				// 确保响应被刷新到客户端
	// 注意：在Gin的Stream函数中，w参数就是ResponseWriter，但我们不需要直接操作它
	// Gin会自动处理flushing
			}
		case <-c.Request.Context().Done():
			// 客户端断开连接
			return
		}
	}
}

// NotifyNewOrder 通知新订单（供其他地方调用）
func (h *Handler) NotifyNewOrder(order models.Order) {
	GlobalOrderBroadcaster.Broadcast(order)
}