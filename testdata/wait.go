package require_wait

import (
	"context"

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
	eg := errgroup.Group{} // want "errgroup 'eg' does not have Wait called"

	eg.Go(func() error {
		return nil
	})

	eg.Go(func() error {
		return nil
	})
}

func errgroupContextWithWait() {
	eg, _ := errgroup.WithContext(context.Background())

	eg.Go(func() error {
		return nil
	})

	eg.Go(func() error {
		return nil
	})

	_ = eg.Wait()
}

func errgroupContextMissingWait() {
	eg, _ := errgroup.WithContext(context.Background()) // want "errgroup 'eg' does not have Wait called"

	eg.Go(func() error {
		return nil
	})

	eg.Go(func() error {
		return nil
	})
}

func errgroupMultipleScopesWithWait() {
	eg := errgroup.Group{}

	eg.Go(func() error {
		return nil
	})

	eg.Go(func() error {
		eg2 := errgroup.Group{}

		eg2.Go(func() error {
			return nil
		})

		return eg2.Wait()
	})

	_ = eg.Wait()
}

func errgroupMultipleScopesMissingWait() {
	eg := errgroup.Group{}

	eg.Go(func() error {
		return nil
	})

	eg.Go(func() error {
		eg2 := errgroup.Group{} // want "errgroup 'eg2' does not have Wait called"

		eg2.Go(func() error {
			return nil
		})

		return nil
	})

	_ = eg.Wait()
}
