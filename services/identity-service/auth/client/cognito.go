package client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type CognitoClient struct {
	client *cognitoidentityprovider.Client
	appID  string
	poolID string
}

func NewCognitoClient(awsRegion, appID, poolID string) (*CognitoClient, error) {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		return nil, err
	}

	client := cognitoidentityprovider.NewFromConfig(sdkConfig)

	return &CognitoClient{
		client: client,
		appID:  appID,
		poolID: poolID,
	}, nil
}

func (c *CognitoClient) SignUp(ctx context.Context, email, password string) (string, error) {
	input := &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(c.appID),
		Username: aws.String(email),
		Password: aws.String(password),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
		},
	}

	output, err := c.client.SignUp(ctx, input)
	if err != nil {
		return "", err
	}

	return *output.UserSub, nil
}

func (c *CognitoClient) ConfirmSignUp(ctx context.Context, email, code string) error {
	input := &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String(c.appID),
		Username:         aws.String(email),
		ConfirmationCode: aws.String(code),
	}

	_, err := c.client.ConfirmSignUp(ctx, input)

	return err
}

func (c *CognitoClient) SignIn(ctx context.Context, email, password string) (*types.AuthenticationResultType, error) {
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(c.appID),
		AuthParameters: map[string]string{
			"USERNAME": email,
			"PASSWORD": password,
		},
	}

	output, err := c.client.InitiateAuth(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("cognito auth failed: %w", err)
	}

	return output.AuthenticationResult, nil
}
