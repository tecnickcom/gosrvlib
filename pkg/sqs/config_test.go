package sqs

import (
	"context"
	"testing"

	"github.com/Vonage/gosrvlib/pkg/awsopt"
	"github.com/stretchr/testify/require"
)

func Test_loadConfig(t *testing.T) {
	var (
		wt     int32 = 13
		vt     int32 = 17
		region       = "eu-central-1"
	)

	o := awsopt.Options{}
	o.WithRegion(region)
	// o.WithEndpoint("https://test.endpoint.invalid", true) // deprecated

	got, err := loadConfig(
		context.TODO(),
		WithAWSOptions(o),
		WithEndpointMutable("https://test.endpoint.invalid"),
		WithWaitTimeSeconds(wt),
		WithVisibilityTimeout(vt),
	)

	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, region, got.awsConfig.Region)
	require.Equal(t, wt, got.waitTimeSeconds)
	require.Equal(t, vt, got.visibilityTimeout)
	require.NotNil(t, got.messageEncodeFunc)
	require.NotNil(t, got.messageDecodeFunc)

	got, err = loadConfig(
		context.TODO(),
		WithMessageEncodeFunc(nil),
	)

	require.Error(t, err)
	require.Nil(t, got)

	got, err = loadConfig(
		context.TODO(),
		WithMessageDecodeFunc(nil),
	)

	require.Error(t, err)
	require.Nil(t, got)

	got, err = loadConfig(
		context.TODO(),
		WithWaitTimeSeconds(-1),
	)

	require.Error(t, err)
	require.Nil(t, got)

	got, err = loadConfig(
		context.TODO(),
		WithWaitTimeSeconds(21),
	)

	require.Error(t, err)
	require.Nil(t, got)

	got, err = loadConfig(
		context.TODO(),
		WithVisibilityTimeout(-1),
	)

	require.Error(t, err)
	require.Nil(t, got)

	got, err = loadConfig(
		context.TODO(),
		WithVisibilityTimeout(43201),
	)

	require.Error(t, err)
	require.Nil(t, got)

	// force aws config.LoadDefaultConfig to fail
	t.Setenv("AWS_ENABLE_ENDPOINT_DISCOVERY", "ERROR")

	got, err = loadConfig(context.TODO())

	require.Error(t, err)
	require.Nil(t, got)
}
