package cliwrapper

import (
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
)

func (s *Client) CreateAlertReceiver(uri, token string, receiver *ormapi.AlertReceiver) (int, error) {
	args := []string{"alert", "receiver", "create"}
	return s.runObjs(uri, token, args, receiver, nil)
}

func (s *Client) DeleteAlertReceiver(uri, token string, receiver *ormapi.AlertReceiver) (int, error) {
	args := []string{"alert", "receiver", "delete"}
	return s.runObjs(uri, token, args, receiver, nil)
}

func (s *Client) ShowAlertReceiver(uri, token string) ([]ormapi.AlertReceiver, int, error) {
	args := []string{"alert", "receiver", "show"}
	receivers := []ormapi.AlertReceiver{}
	st, err := s.runObjs(uri, token, args, nil, &receivers)
	return receivers, st, err
}
