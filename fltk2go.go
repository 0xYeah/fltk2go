package fltk2go

import (
	"github.com/0xYeah/fltk2go/config"
)

// FltkVersion [☑]Option
/*
	en: Get `fltk` binding version;
	zh-CN: 获取绑定的`fltk`版本;
	@return [☑]string en: version string;zh-CN: 版本字符串;
*/
func FltkVersion() string {
	return config.FLTKPreBuildVersion
}

// Version [☑]Option
/*
	en: Get `fltk_go` version;
	zh-CN: 获取`fltk_go`版本;
	@return [☑]string en: version string;zh-CN: 版本字符串;
*/
func Version() string {
	return config.ProjectVersion
}
