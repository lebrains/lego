package events

import (
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"go.uber.org/dig"
	watermilllog "logur.dev/integration/watermill"

	"github.com/vseinstrumentiru/lego/v2/multilog"
)

type ChannelArgs struct {
	dig.In
	Config *gochannel.Config `optional:"true"`
	Logger multilog.Logger
}

func ProvideChannel(in ChannelArgs) (*gochannel.GoChannel, error) {
	cfg := gochannel.Config{}

	if in.Config != nil {
		cfg = *in.Config
	}

	ch := gochannel.NewGoChannel(
		cfg,
		watermilllog.New(in.Logger.WithFields(map[string]interface{}{"component": "events.pubsub", "provider": "gochannel"})),
	)

	return ch, nil
}
