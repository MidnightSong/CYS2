package tests

import (
	"CYS2/core"
	"testing"
)

func TestDoOne(t *testing.T) {
	core.JpgUrls.Push("http://www.ciyo.cn/posts/32244251949057/share")
	t.Log("Testing.....................")
	core.GetJpgPage()

}
