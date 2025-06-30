package permission

import (
	"sync"
	"testing"
	"time"

	"github.com/opencode-ai/opencode/internal/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePermissionRequest(t *testing.T) {
	req := CreatePermissionRequest{
		SessionID:   "session-123",
		ToolName:    "bash",
		Description: "Execute command",
		Action:      "execute",
		Params:      map[string]interface{}{"command": "ls"},
		Path:        "/tmp",
	}

	assert.Equal(t, "session-123", req.SessionID)
	assert.Equal(t, "bash", req.ToolName)
	assert.Equal(t, "Execute command", req.Description)
	assert.Equal(t, "execute", req.Action)
	assert.Equal(t, "/tmp", req.Path)
	assert.NotNil(t, req.Params)
}

func TestPermissionRequest(t *testing.T) {
	req := PermissionRequest{
		ID:          "perm-123",
		SessionID:   "session-123",
		ToolName:    "file",
		Description: "Write file",
		Action:      "write",
		Params:      map[string]string{"file": "test.txt"},
		Path:        "/home/user",
	}

	assert.Equal(t, "perm-123", req.ID)
	assert.Equal(t, "session-123", req.SessionID)
	assert.Equal(t, "file", req.ToolName)
	assert.Equal(t, "Write file", req.Description)
	assert.Equal(t, "write", req.Action)
	assert.Equal(t, "/home/user", req.Path)
	assert.NotNil(t, req.Params)
}

func TestNewPermissionService(t *testing.T) {
	service := NewPermissionService()
	assert.NotNil(t, service)
	
	// Test that it implements the Service interface
	assert.Implements(t, (*Service)(nil), service)
}

func TestPermissionService_AutoApproveSession(t *testing.T) {
	service := NewPermissionService()
	
	// Test auto-approve functionality
	service.AutoApproveSession("auto-session")
	
	// Request should be auto-approved
	approved := service.Request(CreatePermissionRequest{
		SessionID:   "auto-session",
		ToolName:    "test",
		Description: "test action",
		Action:      "test",
		Path:        "/tmp",
	})
	
	assert.True(t, approved)
}

func TestPermissionService_Grant(t *testing.T) {
	service := NewPermissionService().(*permissionService)
	
	// Create a permission request
	permissionReq := PermissionRequest{
		ID:          "test-perm",
		SessionID:   "test-session",
		ToolName:    "test",
		Description: "test",
		Action:      "test",
		Path:        "/tmp",
	}
	
	// Set up a pending request
	respCh := make(chan bool, 1)
	service.pendingRequests.Store(permissionReq.ID, respCh)
	
	// Grant the permission
	go service.Grant(permissionReq)
	
	// Should receive true
	select {
	case result := <-respCh:
		assert.True(t, result)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for grant response")
	}
}

func TestPermissionService_Deny(t *testing.T) {
	service := NewPermissionService().(*permissionService)
	
	// Create a permission request
	permissionReq := PermissionRequest{
		ID:          "test-perm",
		SessionID:   "test-session",
		ToolName:    "test",
		Description: "test",
		Action:      "test",
		Path:        "/tmp",
	}
	
	// Set up a pending request
	respCh := make(chan bool, 1)
	service.pendingRequests.Store(permissionReq.ID, respCh)
	
	// Deny the permission
	go service.Deny(permissionReq)
	
	// Should receive false
	select {
	case result := <-respCh:
		assert.False(t, result)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for deny response")
	}
}

func TestPermissionService_GrantPersistent(t *testing.T) {
	service := NewPermissionService().(*permissionService)
	
	// Create a permission request
	permissionReq := PermissionRequest{
		ID:          "test-perm",
		SessionID:   "test-session",
		ToolName:    "file",
		Description: "write file",
		Action:      "write",
		Path:        "/tmp",
	}
	
	// Set up a pending request
	respCh := make(chan bool, 1)
	service.pendingRequests.Store(permissionReq.ID, respCh)
	
	// Grant persistent permission
	go service.GrantPersistant(permissionReq)
	
	// Should receive true
	select {
	case result := <-respCh:
		assert.True(t, result)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for grant persistent response")
	}
	
	// Permission should be stored
	assert.Contains(t, service.sessionPermissions, permissionReq)
}

func TestPermissionService_Request_WithExistingPermission(t *testing.T) {
	service := NewPermissionService().(*permissionService)
	
	// Add a persistent permission
	existingPerm := PermissionRequest{
		SessionID: "test-session",
		ToolName:  "file",
		Action:    "write",
		Path:      "/tmp",
	}
	service.sessionPermissions = append(service.sessionPermissions, existingPerm)
	
	// Request the same permission
	approved := service.Request(CreatePermissionRequest{
		SessionID:   "test-session",
		ToolName:    "file",
		Description: "write file",
		Action:      "write",
		Path:        "/tmp/test.txt", // Different file but same directory
	})
	
	assert.True(t, approved)
}

func TestPermissionService_Request_PublishesEvent(t *testing.T) {
	service := NewPermissionService()
	
	// Subscribe to events
	var receivedEvent pubsub.Event[PermissionRequest]
	var wg sync.WaitGroup
	wg.Add(1)
	
	service.Subscribe(func(event pubsub.Event[PermissionRequest]) {
		receivedEvent = event
		wg.Done()
		
		// Grant the permission to unblock the request
		service.Grant(event.Data)
	})
	
	// Make a request in a goroutine
	go func() {
		service.Request(CreatePermissionRequest{
			SessionID:   "test-session",
			ToolName:    "test",
			Description: "test",
			Action:      "test",
			Path:        "/tmp",
		})
	}()
	
	// Wait for the event
	wg.Wait()
	
	assert.Equal(t, pubsub.CreatedEvent, receivedEvent.Type)
	assert.Equal(t, "test-session", receivedEvent.Data.SessionID)
	assert.Equal(t, "test", receivedEvent.Data.ToolName)
}

func TestPermissionService_Request_PathHandling(t *testing.T) {
	service := NewPermissionService()
	
	tests := []struct {
		name     string
		inputPath string
		expectedPath string
	}{
		{
			name:      "absolute path",
			inputPath: "/home/user/file.txt",
			expectedPath: "/home/user",
		},
		{
			name:      "relative path with file",
			inputPath: "file.txt",
			expectedPath: ".", // Should be converted to working directory
		},
		{
			name:      "nested relative path",
			inputPath: "subdir/file.txt",
			expectedPath: "subdir",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedPermission PermissionRequest
			
			// Subscribe to capture the permission request
			service.Subscribe(func(event pubsub.Event[PermissionRequest]) {
				capturedPermission = event.Data
				// Immediately grant to unblock
				service.Grant(event.Data)
			})
			
			// Make the request
			service.Request(CreatePermissionRequest{
				SessionID:   "test-session",
				ToolName:    "test",
				Description: "test",
				Action:      "test",
				Path:        tt.inputPath,
			})
			
			// Check the captured path
			if tt.expectedPath == "." {
				// For relative paths, it should use working directory
				// We can't easily test the exact value since it depends on config
				assert.NotEmpty(t, capturedPermission.Path)
			} else {
				assert.Equal(t, tt.expectedPath, capturedPermission.Path)
			}
		})
	}
}

func TestPermissionService_ConcurrentRequests(t *testing.T) {
	service := NewPermissionService()
	
	const numRequests = 10
	results := make([]bool, numRequests)
	var wg sync.WaitGroup
	
	// Subscribe to auto-grant all requests
	service.Subscribe(func(event pubsub.Event[PermissionRequest]) {
		service.Grant(event.Data)
	})
	
	// Make concurrent requests
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = service.Request(CreatePermissionRequest{
				SessionID:   "test-session",
				ToolName:    "test",
				Description: "test",
				Action:      "test",
				Path:        "/tmp",
			})
		}(i)
	}
	
	wg.Wait()
	
	// All requests should be granted
	for i, result := range results {
		assert.True(t, result, "Request %d should be granted", i)
	}
}

func TestPermissionService_Request_GeneratesUniqueIDs(t *testing.T) {
	service := NewPermissionService()
	
	var ids []string
	var mu sync.Mutex
	
	// Subscribe to collect IDs
	service.Subscribe(func(event pubsub.Event[PermissionRequest]) {
		mu.Lock()
		ids = append(ids, event.Data.ID)
		mu.Unlock()
		service.Grant(event.Data)
	})
	
	const numRequests = 5
	var wg sync.WaitGroup
	
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			service.Request(CreatePermissionRequest{
				SessionID:   "test-session",
				ToolName:    "test",
				Description: "test",
				Action:      "test",
				Path:        "/tmp",
			})
		}()
	}
	
	wg.Wait()
	
	// All IDs should be unique
	idSet := make(map[string]bool)
	for _, id := range ids {
		assert.False(t, idSet[id], "ID %s should be unique", id)
		idSet[id] = true
		assert.NotEmpty(t, id, "ID should not be empty")
	}
	
	assert.Len(t, ids, numRequests, "Should have collected all IDs")
}

func TestErrorPermissionDenied(t *testing.T) {
	assert.Equal(t, "permission denied", ErrorPermissionDenied.Error())
}

// Test that the service properly implements the Service interface
func TestServiceInterface(t *testing.T) {
	service := NewPermissionService()
	
	// Test all interface methods exist and can be called
	assert.NotPanics(t, func() {
		service.AutoApproveSession("test")
		
		perm := PermissionRequest{ID: "test"}
		service.Grant(perm)
		service.Deny(perm)
		service.GrantPersistant(perm)
		
		// Subscribe to avoid blocking on Request
		service.Subscribe(func(event pubsub.Event[PermissionRequest]) {
			service.Grant(event.Data)
		})
		
		service.Request(CreatePermissionRequest{
			SessionID: "test",
			ToolName:  "test",
			Action:    "test",
			Path:      "/tmp",
		})
	})
}