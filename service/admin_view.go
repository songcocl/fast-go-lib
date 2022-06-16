package service

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/songcocl/fast-go-lib/consts"
	"strings"
)

type sView struct {
	ModuleName    string
	ModuleBaseUrl string
}

// 视图管理服务
func View(moduleName string) *sView {
	if moduleName == "" {
		moduleName = "admin"
	}
	return &sView{
		ModuleName:    moduleName,
		ModuleBaseUrl: "/" + moduleName,
	}
}

func (s *sView) WriteDefaultLayout(ctx context.Context, tpl string, p g.Map) {
	p["contentTpl"] = tpl
	g.RequestFromCtx(ctx).Response.WriteTpl(s.ModuleName+"/layout/default.html", p)
}
func (s *sView) WriteLayoutByAction(ctx context.Context, controller, action string, p g.Map) {
	p = s.GetPageMapByAction(ctx, controller, action, p)
	g.RequestFromCtx(ctx).Response.WriteTpl(s.ModuleName+"/layout/default.html", p)
}

func (s *sView) GetPageMapByAction(ctx context.Context, controller, action string, p g.Map) g.Map {
	controllerPath := strings.Replace(controller, ".", "/", -1)
	pAdmin, isAdmin := p["admin"]
	if isAdmin {
		delete(p, "admin")
	} else {
		pAdmin = g.Map{}
	}
	p["page_config"] = g.Map{
		"modulename":     s.ModuleName,
		"moduleurl":      s.ModuleBaseUrl,
		"controllername": controller,
		"actionname":     action,
		"jsname":         "backend/" + controllerPath,
		"admin":          pAdmin,
	}
	p["contentTpl"] = s.ModuleName + "/" + controllerPath + "/" + action + ".html"
	return p
}

//GetPageMap 统一处理fast admin前端js需要的页面配置
func (s *sView) GetPageMap(ctx context.Context, param g.Map) g.Map {
	action := gconv.String(param[consts.ACTIONNAME])
	controller := gconv.String(param[consts.CONTROLLERNAME])
	return s.GetPageMapByAction(ctx, controller, action, param)
}
