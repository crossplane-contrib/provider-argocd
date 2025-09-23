package clients

import "errors"

var (
	errProviderConfigNotSet = errors.New("providerConfigRef is not set")
	errGetProviderConfig    = errors.New("cannot get referenced ProviderConfig")
	errFailedToTrackUsage   = errors.New("cannot track ProviderConfig usage")
)
