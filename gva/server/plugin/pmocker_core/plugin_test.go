package pmocker_core

import (
	"context"
	"testing"

	v2 "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	pmockerplugin "github.com/pmocker-io/pmocker/pkg/pmocker/plugin"
	"github.com/gin-gonic/gin"
)

// stubPMockerPlugin 测试用插件，记录 InitPMocker 是否被调用。
// 同时实现 interfaces.Plugin（gva v2）和 pmockerplugin.PMockerPlugin 接口。
type stubPMockerPlugin struct{ called bool }

func (s *stubPMockerPlugin) Register(engine *gin.Engine) {}
func (s *stubPMockerPlugin) InitPMocker(ctx context.Context) error {
	s.called = true
	return nil
}

func TestInitAllPMockerPluginsCalled(t *testing.T) {
	stub := &stubPMockerPlugin{}
	plugins := []interface{}{stub}
	errs := pmockerplugin.InitAllPMockerPlugins(context.Background(), plugins)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if !stub.called {
		t.Errorf("InitPMocker not called on stub plugin")
	}
}

func TestNonPMockerPluginSkipped(t *testing.T) {
	// gva 原生插件不实现 PMockerPlugin，应被跳过不报错
	type nativePlugin struct{}
	plugins := []interface{}{&nativePlugin{}}
	errs := pmockerplugin.InitAllPMockerPlugins(context.Background(), plugins)
	if len(errs) != 0 {
		t.Errorf("native plugin should be skipped, got errors: %v", errs)
	}
}

// 确保 stubPMockerPlugin 实现两个接口
var _ pmockerplugin.PMockerPlugin = (*stubPMockerPlugin)(nil)
var _ v2.Plugin = (*stubPMockerPlugin)(nil)
