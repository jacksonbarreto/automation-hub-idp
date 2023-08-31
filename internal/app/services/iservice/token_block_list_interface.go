package iservice

import "time"

type TokenBlockListService interface {
	AddToBlockList(jwtUUID string, expirationTime time.Duration) error
	IsInBlockList(jwtUUID string) (bool, error)
}
