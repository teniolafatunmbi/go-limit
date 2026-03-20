package utils

import (
	"net"
)

func GetIpFromRemoteAddr(remoteAddr string) (*string, error) {
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return nil, err
	}

	return &ip, nil
}
