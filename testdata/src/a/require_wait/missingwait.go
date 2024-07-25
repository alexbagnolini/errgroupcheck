package require_wait

import (
	"golang.org/x/sync/errgroup"
)

func errgroupWithWait() {
	eg := errgroup.Group{}
	eg.Go(func() error {
		return nil
	})

	eg.Go(func() error {
		return nil
	})

	_ = eg.Wait()
}

func errgroupMissingWait() {
	eg := errgroup.Group{}
	eg.Go(func() error {
		return nil
	})

	eg.Go(func() error {
		return nil
	})

	// want "errgroup must have Wait called at least once"
}
