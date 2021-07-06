package util

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/docker/docker/api/types"
)

const proxyEndpointScheme = "https://"

var ecrPattern = regexp.MustCompile(`(^[a-zA-Z0-9][a-zA-Z0-9-_]*)\.dkr\.ecr(\-fips)?\.([a-zA-Z0-9][a-zA-Z0-9-_]*)\.amazonaws\.com(\.cn)?`)

// ECRCredentialPlugin is a wrapper to generate ECR token using the AWS Credentials
func ECRCredentialPlugin(plugin map[string]string, pr string) (authConfig types.AuthConfig, err error) {

	if strings.HasPrefix(pr, proxyEndpointScheme) {
		pr = strings.TrimPrefix(pr, proxyEndpointScheme)
	}
	matches := ecrPattern.FindStringSubmatch(pr)
	if len(matches) == 0 {
		return authConfig, fmt.Errorf("Not a valid ECR registry")
	} else if len(matches) < 3 {
		return authConfig, fmt.Errorf(pr + "is not a valid repository URI for Amazon Elastic Container Registry.")
	}

	config := &aws.Config{
		Region: aws.String(matches[3]),
	}

	var sess *session.Session
	awsAccessKeyID, accessKeyOK := plugin["aws_access_key_id"]
	awsSecretAccessKey, secretKeyOK := plugin["aws_secret_access_key"]

	// Use predefined keys and override env lookup if keys are present //
	if accessKeyOK && secretKeyOK {
		// if session token doesnt exist just pass empty string
		awsSessionToken := plugin["aws_session_token"]
		config.Credentials = credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, awsSessionToken)
		sess, err = session.NewSession(config)
	} else {
		sess, err = session.NewSessionWithOptions(session.Options{
			Config:            *config,
			SharedConfigState: session.SharedConfigEnable,
		})
	}

	if err != nil {
		return authConfig, err
	}

	ecrClient := ecr.New(sess)

	result, err := ecrClient.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return authConfig, err
	}
	if len(result.AuthorizationData) == 0 {
		return authConfig, fmt.Errorf("No authorization data returned")
	}

	authConfig, err = extractToken(*result.AuthorizationData[0].AuthorizationToken)
	return authConfig, err
}

func extractToken(token string) (authConfig types.AuthConfig, err error) {
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return authConfig, fmt.Errorf("Invalid token: %v", err)
	}

	parts := strings.SplitN(string(decodedToken), ":", 2)
	if len(parts) < 2 {
		return authConfig, fmt.Errorf("Invalid token: expected two parts, got %d", len(parts))
	}

	authConfig = types.AuthConfig{
		Username: parts[0],
		Password: parts[1],
	}

	return authConfig, nil
}
