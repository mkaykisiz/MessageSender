package middlewares

import (
	"github.com/mkaykisiz/sender"
)

// Middleware defines service middleware
type Middleware func(service sender.Service) sender.Service
