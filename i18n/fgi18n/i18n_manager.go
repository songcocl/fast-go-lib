package fgi18n

import (
	"fmt"
	"strings"
)

// Manager for i18n contents, it is concurrent safe, supporting hot reload.
type Manager struct {
	data    map[string]map[string]string // Translating map.
	pattern string                       // Pattern for regex parsing.
	options Options                      // configuration options.
}

// Options is used for i18n object configuration.
type Options struct {
	Language string // Default local language.
}

var (
	defaultLanguage = "zh_cn"
	DfManager       *Manager
)

func init() {
	DfManager = &Manager{
		data: map[string]map[string]string{
			"en":    EN,
			"zh_cn": ZH_CN,
		},
		options: Options{
			Language: defaultLanguage,
		},
	}
}

func (m *Manager) GetCurLangData() map[string]string {
	return m.data[m.options.Language]
}

func (m *Manager) GetVal(key string, values ...interface{}) string {
	curMap := m.GetCurLangData()
	if curMap == nil {
		return key
	}
	key = strings.ToLower(key)
	if val, ok := curMap[key]; ok {
		if len(values) > 0 {
			val = fmt.Sprintf(val, values...)
		}
		return val
	}
	return key
}

func FgLocalize(value string, values ...interface{}) string {
	return DfManager.GetVal(value, values...)
}
