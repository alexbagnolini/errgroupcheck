package errgroupcheck

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestRequireWait(t *testing.T) {
	testdata := analysistest.TestData()
	analyzer := NewAnalyzer(&Settings{
		Mode:        RunningModeNative,
		RequireWait: true,
	})

	analysistest.Run(t, testdata, analyzer)
}
