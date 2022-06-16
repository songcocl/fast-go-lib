package utils

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/songcocl/fast-go-lib/i18n/fgi18n"
)

var (
	btnAttr = g.Map{
		"refresh": g.Map{"href": "javascript:;", "class": "btn btn-primary btn-refresh", "icon": "fa fa-refresh", "text": "", "title": fgi18n.DfManager.GetVal("Refresh")},
		"add":     g.Map{"href": "javascript:;", "class": "btn btn-success btn-add", "icon": "fa fa-plus", "text": fgi18n.DfManager.GetVal("Add"), "title": fgi18n.DfManager.GetVal("Add")},
		"edit":    g.Map{"href": "javascript:;", "class": "btn btn-success btn-edit btn-disabled disabled", "icon": "fa fa-pencil", "text": fgi18n.DfManager.GetVal("Edit"), "title": fgi18n.DfManager.GetVal("Edit")},
		"delete":  g.Map{"href": "javascript:;", "class": "btn btn-danger btn-del btn-disabled disabled", "icon": "fa fa-trash", "text": fgi18n.DfManager.GetVal("Delete"), "title": fgi18n.DfManager.GetVal("Delete")},
		"import":  g.Map{"href": "javascript:;", "class": "btn btn-info btn-import", "icon": "fa fa-upload", "text": fgi18n.DfManager.GetVal("Import"), "title": fgi18n.DfManager.GetVal("Import")},
	}
)

func BuildToolbar(values ...interface{}) string {
	strHtml := ""
	if values == nil {
		return strHtml
	}
	for _, value := range values {
		btn := gconv.String(value)
		btnMap := gconv.Map(btnAttr[btn])
		if btnMap == nil {
			continue
		}
		href := gconv.String(btnMap["href"])
		class := gconv.String(btnMap["class"])
		title := gconv.String(btnMap["title"])
		icon := gconv.String(btnMap["icon"])
		text := gconv.String(btnMap["text"])
		if gstr.Equal(btn, "import") {
			//TODO 处理导入导出
		} else {
			strHtml += "<a href=\"" + href + "\" class=\"" + class + "\" title=\"" + title + "\"><i class=\"" + icon + "\"></i> " + text + "</a> "
		}
	}
	return strHtml
}

func AuthCheck(path string) int {
	return 1
}
