package bot

import (
	"errors"
	"fmt"
	"net/http"

	"k8s.io/client-go/dynamic"
)

type Server interface {
	Serve(port int) error
}

func NewServer(dynamicClient dynamic.Interface) Server {
	return &server{
		dynamicClient: dynamicClient,
	}
}

type server struct {
	dynamicClient dynamic.Interface
}

func (s *server) handler(adapter Adapter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, err := adapter.Parse(r)
		if err != nil {
			adapter.SendResponse(fmt.Sprintf("cannot parse request: %v", err), w)
			return
		}
		if res, err := s.execute(cmd); err != nil {
			adapter.SendResponse(fmt.Sprintf("cannot execute command: %v", err), w)
		} else {
			adapter.SendResponse(res, w)
		}
	}
}

func (s *server) execute(cmd Command) (string, error) {
	switch {
	case cmd.ListSubscriptions != nil:
		return s.listSubscriptions(cmd.Subscriber)
	default:
		return "", errors.New("unknown command")
	}
}

func (s *server) listSubscriptions(subscriber string) (string, error) {
	return fmt.Sprintf("here are your subscriptions mr. %s", subscriber), nil
}

func (s *server) Serve(port int) error {
	http.HandleFunc("/slack", s.handler(&slack{}))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
