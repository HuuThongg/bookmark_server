package util

import (
	"github.com/go-rod/rod"
)

//	func RodGetUrlScreenshot(page *rod.Page) {
//		page.MustScreenshot("a.jpeg")
//	}
func RodGetUrlScreenshot(page *rod.Page) []byte {
	screenShotBytes := page.MustScreenshot("a.jpeg")
	return screenShotBytes
}
