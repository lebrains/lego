package lego

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gorilla/mux"
	"go.opencensus.io/stats/view"
	"google.golang.org/grpc"
	"io"
)

type App interface {
	GetName() string
	SetLogErr(logErr LogErr)
}

type AppWithConfig interface {
	WithCustomConfig
}

type AppWithHttp interface {
	RegisterHTTP(router *mux.Router) error
}

type AppWithGrpc interface {
	RegisterGRPC(server *grpc.Server) error
}

type AppWithEventHandlers interface {
	RegisterEventHandlers(em EventManager) error
}

type AppWithPublishers interface {
	RegisterEventDispatcher(publisher message.Publisher) error
}

type AppWithStats interface {
	GetStats() []*view.View
}

type AppWithRegistration interface {
	Register(p Process) (io.Closer, error)
}
