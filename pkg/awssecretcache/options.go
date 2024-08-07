package awssecretcache

import (
	"context"
	"net/url"

	"github.com/Vonage/gosrvlib/pkg/awsopt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	sep "github.com/aws/smithy-go/endpoints"
)

// SrvOptionFunc is an alias for this service option function.
type SrvOptionFunc = func(*secretsmanager.Options)

// Option is a type to allow setting custom client options.
type Option func(*cfg)

// WithAWSOptions allows to add an arbitrary AWS options.
func WithAWSOptions(opt awsopt.Options) Option {
	return func(c *cfg) {
		c.awsOpts = append(c.awsOpts, opt...)
	}
}

// WithSrvOptionFuncs allows to specify specific options.
func WithSrvOptionFuncs(opt ...SrvOptionFunc) Option {
	return func(c *cfg) {
		c.srvOptFns = append(c.srvOptFns, opt...)
	}
}

// WithSecretsManagerClient overrides the AWS secretemanager.Client with a custom one.
func WithSecretsManagerClient(smclient SecretsManagerClient) Option {
	return func(c *cfg) {
		c.smclient = smclient
	}
}

// WithEndpointMutable sets a mutable endpoint.
func WithEndpointMutable(url string) Option {
	return WithSrvOptionFuncs(
		func(o *secretsmanager.Options) {
			o.BaseEndpoint = aws.String(url)
		},
	)
}

// WithEndpointImmutable sets an immutable endpoint.
func WithEndpointImmutable(url string) Option {
	return WithSrvOptionFuncs(
		func(o *secretsmanager.Options) {
			o.EndpointResolverV2 = &endpointResolver{url: url}
		},
	)
}

type endpointResolver struct {
	url string
}

func (r *endpointResolver) ResolveEndpoint(_ context.Context, _ secretsmanager.EndpointParameters) (
	sep.Endpoint,
	error,
) {
	u, err := url.Parse(r.url)
	if err != nil {
		return sep.Endpoint{}, err //nolint:wrapcheck
	}

	return sep.Endpoint{URI: *u}, nil
}
