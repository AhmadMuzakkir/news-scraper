package provider

import "github.com/pkg/errors"

var ErrPageEmpty = errors.New("The page is empty")
var ErrNoContent = errors.New("Failed to get the news detail")
var ErrOldContent = errors.New("The content already exists in the database")
var ErrProviderNotFound = errors.New("Provider not found !")
