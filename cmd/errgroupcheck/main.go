package main

import (
	"github.com/alexbagnolini/errgroupcheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(errgroupcheck.NewAnalyzer(nil))
}
