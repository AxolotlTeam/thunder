package srv

import (
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

type service struct {
	opts Options
}

// newService -
func newService(opts ...Option) Service {
	options := newOptions(opts...)

	return &service{
		opts: options,
	}
}

func (s *service) Init(opts ...Option) {
	for _, o := range opts {
		o(&s.opts)
	}
	s.opts.BeforeStart()
}

func (s *service) ID() string {
	return s.opts.id
}

func (s *service) Version() string {
	return s.opts.version
}

func (s *service) Name() string {
	return s.opts.name
}

func (s *service) Options() Options {
	return s.opts
}

func (s *service) Server() *grpc.Server {
	return s.opts.Server
}

func (s *service) Run() error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	go func(o *Options) {
		<-c
		deRegister(o)
		os.Exit(1)
	}(&s.opts)

	return s.Start()
}

func (s *service) Start() error {
	// register server
	registerServer(&s.opts)

	c := make(chan error, 1)
	go func(o *Options) {
		if err := s.opts.Server.Serve(s.opts.Listener); err != nil {
			c <- err
		}
	}(&s.opts)

	s.opts.AfterStart()

	return <-c
}
