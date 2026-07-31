package uikit

import "testing"

func TestFacadeNilReceiversAreSafe(t *testing.T) {
	var win *UIWindow
	win.Show()
	if win.RootView() != nil || win.Raw() != nil {
		t.Fatal("nil UIWindow should return nil root/raw")
	}

	var label *UILabel
	label.SetText("ignored")
	label.SetFontSize(12)
	label.SetTextColor(0)
	label.SetBackgroundColor(0)
	if label.View() != nil || label.Raw() != nil {
		t.Fatal("nil UILabel should return nil view/raw")
	}

	var button *UIButton
	button.SetTitle("ignored")
	button.SetBackgroundColor(0)
	button.SetTitleColor(0)
	button.OnTouchUpInside(nil)
	if button.View() != nil || button.Raw() != nil {
		t.Fatal("nil UIButton should return nil view/raw")
	}

	var input *Input
	input.SetText("ignored")
	input.SetPlaceholder("ignored")
	input.SetFontSize(12)
	input.SetTextColor(0)
	input.SetBackgroundColor(0)
	input.SetEnabled(false)
	input.OnChange(nil)
	if input.Text() != "" || input.Placeholder() != "" || input.IsEnabled() || input.View() != nil || input.Raw() != nil {
		t.Fatal("nil Input should return zero values")
	}

	var slider *UISlider
	slider.SetMinimumValue(0)
	slider.SetMaximumValue(1)
	slider.SetStep(0.1)
	slider.SetValue(0.5)
	slider.OnValueChanged(nil)
	slider.SetVertical(true)
	if slider.Value() != 0 || slider.View() != nil || slider.Raw() != nil {
		t.Fatal("nil UISlider should return zero values")
	}

	var progress *UIProgressView
	progress.SetMinimumValue(0)
	progress.SetMaximumValue(1)
	progress.SetProgress(0.5)
	progress.SetTrackColor(0)
	progress.SetProgressTintColor(0)
	if progress.Progress() != 0 || progress.View() != nil || progress.Raw() != nil {
		t.Fatal("nil UIProgressView should return zero values")
	}

	var sw *UISwitch
	sw.SetOn(true)
	sw.OnValueChanged(nil)
	if sw.IsOn() || sw.View() != nil || sw.Raw() != nil {
		t.Fatal("nil UISwitch should return zero values")
	}

	var scroll *UIScrollView
	scroll.AddSubview(nil)
	scroll.ScrollTo(1, 2)
	scroll.SetScrollType(0)
	if x, y := scroll.ContentOffset(); x != 0 || y != 0 {
		t.Fatalf("nil UIScrollView offset = (%d, %d), want (0, 0)", x, y)
	}
	if scroll.View() != nil || scroll.Raw() != nil {
		t.Fatal("nil UIScrollView should return nil view/raw")
	}

	var split *UISplitView
	split.SetLeftView(nil)
	split.SetRightView(nil)
	split.SetLeftViewFixed(10)
	split.SetRightViewFixed(10)
	split.SetPosition(0.5)
	split.SetPositionPixels(100)
	split.SetPositionPolicy(SplitPreserveRatio)
	split.SetMinimumSizes(10, 10)
	split.SetDividerSize(6)
	split.SetDividerColors(0, 0, 0)
	split.SetResizable(false)
	split.OnPositionChanged(nil)
	if split.Position() != 0 || split.PositionPixels() != 0 || split.Geometry() != (SplitGeometry{}) || split.Resizable() {
		t.Fatal("nil UISplitView should return zero state")
	}
	if split.SetAutomationID("ignored") != nil || split.LeadingPane() != nil || split.TrailingPane() != nil || split.DividerView() != nil {
		t.Fatal("nil UISplitView should return nil subviews")
	}
	if split.View() != nil || split.Raw() != nil {
		t.Fatal("nil UISplitView should return nil view/raw")
	}

	var stack *UIStackView
	stack.AddArrangedSubview(nil)
	stack.SetAxis(AxisHorizontal)
	stack.SetSpacing(8)
	stack.SetMargin(8)
	stack.SetFixedSize(nil, 10)
	stack.Layout()
	stack.End()
	if stack.View() != nil || stack.Raw() != nil {
		t.Fatal("nil UIStackView should return nil view/raw")
	}

	var text *UITextView
	text.SetText("ignored")
	text.Append("ignored")
	text.SetWrapAtBounds()
	text.SetFontSize(12)
	text.SetTextColor(0)
	text.OnTextChanged(nil)
	if text.Text() != "" || text.View() != nil || text.Raw() != nil || text.TextBuffer() != nil {
		t.Fatal("nil UITextView should return zero values")
	}
}
