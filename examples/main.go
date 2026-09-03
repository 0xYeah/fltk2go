package main

import (
	"errors"
	"html"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/0xdevelop/fltk2go"
	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/uikit/automation"
	"github.com/0xdevelop/fltk2go/uikit/view"

	example_auth "examples/auth"
	example_comprehensive "examples/comprehensive"
	example_counter "examples/counter"
	example_input "examples/input"
	example_loginview "examples/loginview"
	example_navigationbar "examples/navigationbar"
	example_slider_progress "examples/slider_progress"
	example_splitview "examples/splitview"
	example_tableview "examples/tableview"
	example_tableview_demo "examples/tableview_demo"
	example_tabs "examples/tabs"
)

const (
	winW        = 1200
	winH        = 700
	leftSplit   = 200
	middleSplit = 500
)

type example struct {
	title     string
	html      string
	dir       string
	buildFunc func(parent *view.UIView) view.Viewable
}

var examples = []example{
	{
		title: "Counter",
		html: `<h2>Counter</h2>
<p>基础按钮点击计数器。</p>
<b>演示要点：</b>
<ul>
<li>UIButton 点击回调 OnTouchUpInside</li>
<li>UILabel 文本动态更新 SetText</li>
<li>runtime.LockOSThread() 线程绑定</li>
</ul>`,
		dir:       "./counter",
		buildFunc: example_counter.BuildView,
	},
	{
		title: "Comprehensive",
		html: `<h2>Comprehensive</h2>
<p>综合控件示例，展示多种按钮类型与 TableView 联动。</p>
<b>演示要点：</b>
<ul>
<li>SystemButton / CheckboxButton / RadioButton / ToggleButton</li>
<li>UIView.SetTooltip native hover help and automation metadata</li>
<li>TableView DataSource / Delegate 模式</li>
<li>动态添加 / 删除行，ReloadData</li>
<li>SetCustomDraw 自定义行绘制</li>
</ul>`,
		dir:       "./comprehensive",
		buildFunc: example_comprehensive.BuildView,
	},
	{
		title: "Input",
		html: `<h2>Input</h2>
<p>各类输入框示例。</p>
<b>演示要点：</b>
<ul>
<li>普通文本输入 input.New</li>
<li>整数输入 input.IntInput</li>
<li>浮点输入 input.FloatInput</li>
<li>密码框 / 多行文本框</li>
<li>Text() 读取输入内容</li>
</ul>`,
		dir:       "./input",
		buildFunc: example_input.BuildView,
	},
	{
		title: "SplitView",
		html: `<h2>SplitView</h2>
<p>水平与垂直分割视图布局示例。</p>
<b>演示要点：</b>
<ul>
<li>splitview.New 水平/垂直模式</li>
<li>SetLeftView / SetRightView 设置子面板</li>
<li>SetLeftViewFixed 固定左侧宽度</li>
<li>跨面板联动更新</li>
</ul>`,
		dir:       "./splitview",
		buildFunc: example_splitview.BuildView,
	},
	{
		title: "TableView",
		html: `<h2>TableView</h2>
<p>服务器管理列表，演示 DataSource / Delegate 模式。</p>
<b>演示要点：</b>
<ul>
<li>DataSource.NumberOfRows / CellForRow</li>
<li>Delegate.DidSelectRow / RowHeight</li>
<li>SetCustomDraw 自定义行绘制</li>
<li>动态添加 / 删除服务器记录</li>
</ul>`,
		dir:       "./tableview",
		buildFunc: example_tableview.BuildView,
	},
	{
		title: "Slider &amp; Progress",
		html: `<h2>Slider &amp; Progress</h2>
<p>滑块与进度条实时联动控制。</p>
<b>演示要点：</b>
<ul>
<li>fltk_bridge.NewSlider 水平滑块</li>
<li>fltk_bridge.NewValueSlider 带数值显示</li>
<li>fltk_bridge.NewProgress 进度条</li>
<li>WhenChanged 实时回调</li>
<li>Reset / 50% / Max 三挡按钮</li>
</ul>`,
		dir:       "./slider_progress",
		buildFunc: example_slider_progress.BuildView,
	},
	{
		title: "Tabs",
		html: `<h2>Tabs</h2>
<p>选项卡容器示例，展示 FLTK begin/end 自动归属机制。</p>
<b>演示要点：</b>
<ul>
<li>fltk_bridge.NewTabs + NewGroup begin/end</li>
<li>Choice 下拉颜色选择器</li>
<li>Spinner 整数调节</li>
<li>ValueSlider 浮点调节</li>
<li>多控件联动更新显示</li>
</ul>`,
		dir:       "./tabs",
		buildFunc: example_tabs.BuildView,
	},
	{
		title: "TableView Demo",
		html: `<h2>TableView Demo</h2>
<p>从 JSON 文件加载服务器数据的完整 TableView 演示。</p>
<b>演示要点：</b>
<ul>
<li>encoding/json 解析 servers.json</li>
<li>ServerTableDataSource / ServerTableDelegate</li>
<li>tableview.New 创建列表</li>
<li>刷新按钮重新加载数据</li>
</ul>
<p><i>可通过主入口预览，也可直接调试 tableview_demo/main.go。</i></p>`,
		dir:       "./tableview_demo",
		buildFunc: example_tableview_demo.BuildView,
	},
	{
		title: "NavigationBar",
		html: `<h2>NavigationBar</h2>
<p>iOS 风格顶部导航栏组件演示。</p>
<b>演示要点：</b>
<ul>
<li>UINavigationBar / UINavigationItem 结构</li>
<li>左右侧 BarButtonItem 动态添加</li>
<li>标题居中与底部分割线</li>
<li>动态修改背景色与导航栈模拟</li>
</ul>`,
		dir:       "./navigationbar",
		buildFunc: example_navigationbar.BuildView,
	},
	{
		title: "Auth",
		html: `<h2>Auth</h2>
<p>高仿 iOS/MacOS 的登录视图与导航栏综合演示。</p>
<b>演示要点：</b>
<ul>
<li>UIWindow 作为基础窗口承载子视图</li>
<li>UINavigationBar 提供顶部导航</li>
<li>LoginView 提供用户认证界面</li>
<li>组件间通过 AddSubview 组合</li>
</ul>`,
		dir:       "./auth",
		buildFunc: example_auth.BuildView,
	},
	{
		title: "LoginView",
		html: `<h2>LoginView</h2>
<p>登录面板组件的独立演示。</p>
<b>演示要点：</b>
<ul>
<li>LoginView 账号密码输入区域</li>
<li>OnLoginClick 回调直接获取账密</li>
<li>作为普通 UIView 子视图加入页面</li>
</ul>`,
		dir:       "./loginview",
		buildFunc: example_loginview.BuildView,
	},
}

func main() {
	runtime.LockOSThread()
	fltk_bridge.EnableTooltips()
	fltk_bridge.SetTooltipDelay(0.4)

	win := fltk_bridge.NewWindow(winW, winH, "FLTK2Go — Examples Launcher")

	// Tile covers the whole window and provides the draggable split handle.
	tile := fltk_bridge.NewTile(0, 0, winW, winH)

	// ── Left panel: example list ──────────────────────────────────────────
	leftGrp := fltk_bridge.NewGroup(0, 0, leftSplit, winH)

	listHdr := fltk_bridge.NewBox(fltk_bridge.UP_BOX, 0, 0, leftSplit, 32, "Examples")
	listHdr.SetLabelFont(fltk_bridge.HELVETICA_BOLD)
	listHdr.SetLabelSize(13)

	browser := fltk_bridge.NewHoldBrowser(0, 32, leftSplit, winH-32)
	for _, e := range examples {
		browser.Add(e.title)
	}
	browser.End() // HoldBrowser is a Group subclass; restore leftGrp as current.

	leftGrp.Resizable(browser)
	leftGrp.End()

	// ── Middle panel: description ──────────────────────────────────────────────
	midW := middleSplit - leftSplit
	midGrp := fltk_bridge.NewGroup(leftSplit, 0, midW, winH)

	titleBar := fltk_bridge.NewBox(fltk_bridge.UP_BOX, leftSplit, 0, midW, 44,
		"  Description")
	titleBar.SetLabelFont(fltk_bridge.HELVETICA_BOLD)
	titleBar.SetLabelSize(15)
	titleBar.SetAlign(fltk_bridge.ALIGN_LEFT | fltk_bridge.ALIGN_INSIDE)

	helpView := fltk_bridge.NewHelpView(leftSplit, 44, midW, winH-44-58)
	helpView.SetValue(
		`<p><font color="#999999">← 从左侧列表选择一个示例，查看说明并在右侧预览。</font></p>`)
	helpView.TextSize(13)

	runBtn := fltk_bridge.NewButton(leftSplit+10, winH-48, midW-20, 38,
		"Run Selected Example in New Process")
	runBtn.SetColor(fltk_bridge.Color(0x42A5F500))
	runBtn.SetLabelColor(fltk_bridge.Color(0xFFFFFF00))
	runBtn.SetLabelFont(fltk_bridge.HELVETICA_BOLD)
	runBtn.SetLabelSize(13)

	midGrp.Resizable(helpView)
	midGrp.End()

	// ── Right panel: preview ──────────────────────────────────────────────
	rightW := winW - middleSplit
	rightGrp := fltk_bridge.NewGroup(middleSplit, 0, rightW, winH)

	previewTitle := fltk_bridge.NewBox(fltk_bridge.UP_BOX, middleSplit, 0, rightW, 44,
		"  Preview")
	previewTitle.SetLabelFont(fltk_bridge.HELVETICA_BOLD)
	previewTitle.SetLabelSize(15)
	previewTitle.SetAlign(fltk_bridge.ALIGN_LEFT | fltk_bridge.ALIGN_INSIDE)

	previewArea := fltk_bridge.NewGroup(middleSplit, 44, rightW, winH-44)
	previewArea.End()

	rightGrp.Resizable(previewArea)
	rightGrp.End()

	tile.End()

	// Create a UIView for previewArea so BuildView can use it
	previewView := &view.UIView{}
	previewView.BindRaw(previewArea)
	previewView.BindHost(previewArea)
	previewView.SetAutomationID("examples.preview").
		SetAutomationName("Examples preview area").
		SetAutomationRole("region")

	browserAutomation := &view.UIView{}
	browserAutomation.BindRaw(browser)
	browserAutomation.SetAutomationID("examples.launcher.list").
		SetAutomationName("Examples list").
		SetAutomationRole("listbox").
		SetAutomationValueHandler(func() (string, bool) {
			idx := browser.Value() - 1
			if idx >= 0 && idx < len(examples) {
				return examples[idx].title, true
			}
			return "", true
		})

	runBtnAutomation := &view.UIView{}
	runBtnAutomation.BindRaw(runBtn)
	runBtnAutomation.SetAutomationID("examples.launcher.run_selected").
		SetAutomationName("Run selected example in new process").
		SetAutomationRole("button")

	// ── Callbacks ─────────────────────────────────────────────────────────
	selectExample := func() {
		idx := browser.Value() - 1 // FLTK browser is 1-based
		if idx < 0 || idx >= len(examples) {
			return
		}
		view.AutomationUnregisterPrefix("counter.")
		view.AutomationUnregisterPrefix("input.")
		view.AutomationUnregisterPrefix("slider.")
		previewView.ClearAutomationChildren()
		e := examples[idx]
		titleBar.SetLabel("  " + e.title)
		titleBar.Redraw()
		helpView.SetValue(e.html)

		// Clear the preview area
		previewArea.Clear()

		// Build the new view
		if e.buildFunc != nil {
			previewArea.Begin()
			e.buildFunc(previewView)
			previewArea.End()
		}
		previewArea.Redraw()
	}
	browserAutomation.SetAutomationTextHandlers(func(title string) error {
		for idx, e := range examples {
			if e.title == title || html.UnescapeString(e.title) == title {
				browser.SetValue(idx + 1)
				selectExample()
				return nil
			}
		}
		return errors.New("example not found")
	}, func() (string, bool) {
		idx := browser.Value() - 1
		if idx >= 0 && idx < len(examples) {
			return examples[idx].title, true
		}
		return "", true
	})
	browser.SetCallback(selectExample)

	runSelected := func() {
		idx := browser.Value() - 1
		if idx < 0 || idx >= len(examples) {
			return
		}
		wd, _ := os.Getwd()
		cmd := exec.Command("go", "run", examples[idx].dir+"/main.go")
		cmd.Dir = wd
		_ = cmd.Start()
	}
	runBtn.SetCallback(runSelected)
	runBtnAutomation.OnAutomationClick(func() error {
		runSelected()
		return nil
	})

	browser.SetValue(1)
	selectExample()

	if automation.Enabled() {
		addr := os.Getenv("FLTK2GO_AUTOMATION_ADDR")
		if addr == "" {
			addr = "127.0.0.1:8765"
		}
		// FLTK requires Fl::lock() before background goroutines can safely wake
		// the UI event loop through Fl::awake().
		fltk_bridge.Lock()
		srv, err := automation.StartDebugServer(automation.Config{Addr: addr})
		if err != nil {
			log.Printf("automation debug server disabled: %v", err)
		} else {
			defer srv.Close()
			log.Printf("automation debug server listening on http://%s", srv.Addr())
		}
	}

	win.Show()
	fltk2go.Run()
}
