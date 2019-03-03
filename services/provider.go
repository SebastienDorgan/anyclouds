package services

import (
	"github.com/SebastienDorgan/anyclouds/api"
)

//ProviderService defines compute services
type ProviderService struct {
	api.Provider
	extra TemplatesExtra
}
