package ps4

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	ev "github.com/gvalkov/golang-evdev"
)

type Type int

//go:generate stringer -type=Type

const (
	_ Type = iota
	Controller
	MotionSensors
	Touchpad
)

type Input struct {
	Device *ev.InputDevice
	Type   Type
}

func Discover() ([]*Input, error) {
	// "Sony Interactive Entertainment" is only if wired
	controllerRegex := regexp.MustCompile("(?:Sony Interactive Entertainment )?Wireless Controller")

	candidates, err := ev.ListInputDevices("/dev/input/event*")
	if err != nil {
		return []Input{}, err
	}

	inputs := []*Input{}
	for _, candidate := range candidates {
		name := strings.TrimSpace(controllerRegex.ReplaceAllString(candidate.Name, ""))

		switch name {
		case "":
			inputs = append(inputs, &Input{
				Device: candidate,
				Type:   Controller,
			})
		case "Motion Sensors":
			inputs = append(inputs, &Input{
				Device: candidate,
				Type:   MotionSensors,
			})
		case "Touchpad":
			inputs = append(inputs, &Input{
				Device: candidate,
				Type:   Touchpad,
			})
		default:
			fmt.Printf("Skipping: %s\n", candidate)
		}
	}

	if len(inputs) == 0 {
		return inputs, errors.New("unable to find any controller inputs, is it paired and on?")
	}

	return inputs, nil
}

type Button int

//go:generate stringer -type=Button

const (
	// dpad
	DPadX Button = 16
	DPadY Button = 17

	// sticks
	LeftStickX      Button = 0
	LeftStickY      Button = 1
	LeftStickClick  Button = 317
	RightStickX     Button = 3
	RightStickY     Button = 4
	RightStickClick Button = 318

	// triggers
	L1      Button = 310
	L2Click Button = 312
	L2      Button = 2
	R1      Button = 311
	R2Click Button = 313
	R2      Button = 5

	// aux
	Share       Button = 314
	Options     Button = 315
	Playstation Button = 316

	// shapes
	Triangle Button = 307
	Circle   Button = 305
	X        Button = 304
	Square   Button = 308

	// trackpad TODO
	//TrackpadX
	//TrackpadY
	//TrackpadClick

	// motion TODO
)

type KeyState uint8

//go:generate stringer -type=KeyState

const (
	KeyUp   KeyState = iota
	KeyDown
)

// KeyEvent is a button press/release
type KeyEvent struct {
	Event  *ev.InputEvent
	Button Button
	State  KeyState
}

// AbsEvent is an absolute value report (joysticks, lower triggers, or the d-pad - surprisingly)
type AbsEvent struct {
	Event  *ev.InputEvent
	Button Button
	Value  int32
}

func Watch(ctx context.Context, input Input) (<-chan interface{}, error) {
	events := make(chan interface{}, 10)

	go func() {
		defer close(events)
		dev := input.Device

		fmt.Println("Running watcher...")
		for {
			event, err := dev.ReadOne()
			if err != nil {
				fmt.Printf("Unable to read one: %s\n", err)
				continue
			}

			if event.Type == 0 {
				continue
			}

			var e interface{}
			switch event.Type {
			case ev.EV_KEY:
				e = &KeyEvent{
					Event:  event,
					Button: Button(event.Code),
					State:  KeyState(event.Value),
				}

			case ev.EV_ABS:
				e = &AbsEvent{
					Event:  event,
					Button: Button(event.Code),
					Value:  event.Value,
				}

			case ev.EV_MSC:
				continue // only received when on USB, ignored since it adds nothing.

			default:
				fmt.Printf("Skipped: %+v\n", event)
				continue
			}

			select {
			case <-ctx.Done():
				return
			case events <-e:
			default:
			}
		}
	}()

	return events, nil
}
