package doc

import (
	"github.com/mobiledgex/edge-cloud-infra/mc/ormapi"
)

// Success
// swagger:response success
type successResponse struct {
	// in: body
	Body ormapi.Result
}

// Status Bad Request
// swagger:response badRequest
type badReqResponse struct {
	// in:body
	Body ormapi.Result
}

// Forbidden
// swagger:response forbidden
type forbiddenResponse struct {
	// in: body
	Body ormapi.Result
}

// Not Found
// swagger:response notFound
type notFoundResponse struct {
	// in: body
	Body ormapi.Result
}

// swagger:parameters Login
type swaggerLogin struct {
	// in: body
	Body ormapi.UserLogin
}

// swagger:parameters CreateUser
type swaggerCreateUser struct {
	// in: body
	Body ormapi.CreateUser
}

// swagger:parameters DeleteUser
type swaggerDeleteUser struct {
	// in: body
	Body ormapi.User
}

// swagger:parameters ShowUser
type swaggerShowUser struct {
	// in: body
	Body ormapi.Organization
}
