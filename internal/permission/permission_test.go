package permission

import (
	"context"
	"testing"
	"time"

	"github.com/opencode-ai/opencode/internal/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestNewPermissionService(t *testing.T) {
	t.Parallel()

	service := NewPermissionService()
	assert.NotNil(t, service)
	
	// 确保服务实现了所需的接口
	assert.Implements(t, (*Service)(nil), service)
}

func TestPermissionService_AutoApproveSession(t *testing.T) {
	t.Parallel()

	service := NewPermissionService().(*permissionService)
	sessionID := "test-session-123"
	
	// 添加自动批准的会话
	service.AutoApproveSession(sessionID)
	
	// 验证会话已添加到自动批准列表
	assert.Contains(t, service.autoApproveSessions, sessionID)
}

func TestPermissionService_RequestWithAuthorization(t *testing.T) {
	service := NewPermissionService().(*permissionService)
	sessionID := "test-session-123"
	
	// 设置自动批准会话
	service.AutoApproveSession(sessionID)
	
	// 创建权限请求
	request := CreatePermissionRequest{
		SessionID:   sessionID,
		ToolName:    "bash",
		Description: "Run a command",
		Action:      "execute",
		Path:        "/tmp/test",
	}
	
	// 请求应该自动批准
	result := service.Request(request)
	assert.True(t, result)
}

func TestPermissionService_RequestWithSubscription(t *testing.T) {
	service := NewPermissionService().(*permissionService)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 订阅权限请求事件
	eventCh := service.Subscribe(ctx)
	
	// 创建权限请求
	request := CreatePermissionRequest{
		SessionID:   "test-session-456",
		ToolName:    "edit",
		Description: "Edit a file",
		Action:      "write",
		Path:        "/tmp/test.txt",
	}
	
	// 在后台处理权限请求
	go func() {
		select {
		case event := <-eventCh:
			if event.Type == pubsub.CreatedEvent {
				// 批准请求
				service.Grant(event.Payload)
			}
		case <-time.After(1 * time.Second):
			t.Error("No permission request event received")
		}
	}()
	
	// 请求权限
	result := service.Request(request)
	assert.True(t, result)
}

func TestPermissionService_RequestDenied(t *testing.T) {
	service := NewPermissionService().(*permissionService)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 订阅权限请求事件
	eventCh := service.Subscribe(ctx)
	
	// 创建权限请求
	request := CreatePermissionRequest{
		SessionID:   "test-session-789",
		ToolName:    "bash",
		Description: "Run dangerous command",
		Action:      "execute",
		Path:        "/etc",
	}
	
	// 在后台拒绝权限请求
	go func() {
		select {
		case event := <-eventCh:
			if event.Type == pubsub.CreatedEvent {
				// 拒绝请求
				service.Deny(event.Payload)
			}
		case <-time.After(1 * time.Second):
			t.Error("No permission request event received")
		}
	}()
	
	// 请求权限
	result := service.Request(request)
	assert.False(t, result)
}

func TestPermissionService_GrantPersistent(t *testing.T) {
	service := NewPermissionService().(*permissionService)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 订阅权限请求事件
	eventCh := service.Subscribe(ctx)
	
	// 创建权限请求
	request := CreatePermissionRequest{
		SessionID:   "test-session-persistent",
		ToolName:    "read",
		Description: "Read a file",
		Action:      "read",
		Path:        "/tmp/data.txt",
	}
	
	// 在后台持久性批准权限请求
	go func() {
		select {
		case event := <-eventCh:
			if event.Type == pubsub.CreatedEvent {
				// 持久性批准请求
				service.GrantPersistant(event.Payload)
			}
		case <-time.After(1 * time.Second):
			t.Error("No permission request event received")
		}
	}()
	
	// 第一次请求权限
	result1 := service.Request(request)
	assert.True(t, result1)
	
	// 第二次请求相同权限应该直接批准（无需用户交互）
	result2 := service.Request(request)
	assert.True(t, result2)
	
	// 验证权限已保存到会话权限中
	found := false
	for _, p := range service.sessionPermissions {
		if p.SessionID == request.SessionID && 
		   p.ToolName == request.ToolName && 
		   p.Action == request.Action {
			found = true
			break
		}
	}
	assert.True(t, found, "Permission should be saved to session permissions")
}

func TestPermissionService_PathHandling(t *testing.T) {
	t.Parallel()

	service := NewPermissionService().(*permissionService)
	
	// 测试自动批准的会话
	service.AutoApproveSession("test-session-path")
	
	// 创建权限请求测试路径处理
	request := CreatePermissionRequest{
		SessionID:   "test-session-path",
		ToolName:    "test",
		Description: "Test path handling",
		Action:      "test",
		Path:        "/tmp/test.txt",
	}
	
	// 请求权限应该自动批准
	result := service.Request(request)
	assert.True(t, result)
}

func TestErrorPermissionDenied(t *testing.T) {
	t.Parallel()

	err := ErrorPermissionDenied
	assert.Error(t, err)
	assert.Equal(t, "permission denied", err.Error())
}