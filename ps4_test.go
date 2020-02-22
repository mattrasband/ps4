package ps4

import (
	"context"
	"fmt"
	"testing"
)

func TestRun(t *testing.T) {
	candidates, err := Discover()
	if err != nil {
		fmt.Printf("Error during discovery: %s\n", err)
		return
	}

	var dev *Input
	for _, c := range candidates {
		if c.Type == Controller {
			dev = c
			break
		}
	}

	fmt.Printf("Found device: %+v\n", dev)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := Watch(ctx, dev)
	if err != nil {
		fmt.Printf("error starting watcher: %s\n", err)
		return
	}
	fmt.Printf("Watching...\n")

	for e := range events {
		if av, ok := e.(*AbsEvent); ok {
			if (av.Button >= LeftStickX && av.Button <= RightStickClick) && (av.Value >= 120 && av.Value <= 134) {
				continue
			}
		}

		fmt.Printf("%+v\n", e)
	}
}
