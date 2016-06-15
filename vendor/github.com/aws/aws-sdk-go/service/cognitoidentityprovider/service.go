// THIS FILE IS AUTOMATICALLY GENERATED. DO NOT EDIT.

package cognitoidentityprovider

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/private/protocol/jsonrpc"
	"github.com/aws/aws-sdk-go/private/signer/v4"
)

// You can create a user pool in Amazon Cognito Identity to manage directories
// and users. You can authenticate a user to obtain tokens related to user identity
// and access policies.
//
// This API reference provides information about user pools in Amazon Cognito
// Identity, which is a new capability that is available as a beta.
//The service client's operations are safe to be used concurrently.
// It is not safe to mutate any of the client's properties though.
type CognitoIdentityProvider struct {
	*client.Client
}

// Used for custom client initialization logic
var initClient func(*client.Client)

// Used for custom request initialization logic
var initRequest func(*request.Request)

// A ServiceName is the name of the service the client will make API calls to.
const ServiceName = "cognito-idp"

// New creates a new instance of the CognitoIdentityProvider client with a session.
// If additional configuration is needed for the client instance use the optional
// aws.Config parameter to add your extra config.
//
// Example:
//     // Create a CognitoIdentityProvider client from just a session.
//     svc := cognitoidentityprovider.New(mySession)
//
//     // Create a CognitoIdentityProvider client with additional configuration
//     svc := cognitoidentityprovider.New(mySession, aws.NewConfig().WithRegion("us-west-2"))
func New(p client.ConfigProvider, cfgs ...*aws.Config) *CognitoIdentityProvider {
	c := p.ClientConfig(ServiceName, cfgs...)
	return newClient(*c.Config, c.Handlers, c.Endpoint, c.SigningRegion)
}

// newClient creates, initializes and returns a new service client instance.
func newClient(cfg aws.Config, handlers request.Handlers, endpoint, signingRegion string) *CognitoIdentityProvider {
	svc := &CognitoIdentityProvider{
		Client: client.New(
			cfg,
			metadata.ClientInfo{
				ServiceName:   ServiceName,
				SigningRegion: signingRegion,
				Endpoint:      endpoint,
				APIVersion:    "2016-04-18",
				JSONVersion:   "1.1",
				TargetPrefix:  "AWSCognitoIdentityProviderService",
			},
			handlers,
		),
	}

	// Handlers
	svc.Handlers.Sign.PushBack(v4.Sign)
	svc.Handlers.Build.PushBackNamed(jsonrpc.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(jsonrpc.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(jsonrpc.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(jsonrpc.UnmarshalErrorHandler)

	// Run custom client initialization if present
	if initClient != nil {
		initClient(svc.Client)
	}

	return svc
}

// newRequest creates a new request for a CognitoIdentityProvider operation and runs any
// custom request initialization.
func (c *CognitoIdentityProvider) newRequest(op *request.Operation, params, data interface{}) *request.Request {
	req := c.NewRequest(op, params, data)

	// Run custom request initialization if present
	if initRequest != nil {
		initRequest(req)
	}

	return req
}