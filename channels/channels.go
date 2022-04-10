package channels

import "sync"

func MergeChannelsString(channels ...chan string) chan string {
	merged := make(chan string)

	var activeChannels sync.WaitGroup
	for _, ch := range channels {
		activeChannels.Add(1)
		go func(c chan string) {
			for val := range c {
				merged <- val
			}
			activeChannels.Done()
		}(ch)
	}

	// Close the `merged` channel after all other channels have closed.
	go func() {
		activeChannels.Wait()
		close(merged)
	}()

	return merged
}
