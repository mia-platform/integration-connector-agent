package client

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/v3"
)

type Client interface {
	GetByID(resourceID string) (armresources.ClientGetByIDResponse, error)
}

func New(credentials azcore.TokenCredential) Client {
	return &azureClient{
		credentials: credentials,
	}
}

type azureClient struct {
	credentials azcore.TokenCredential
}

func (a *azureClient) GetByID(resourceID string) (armresources.ClientGetByIDResponse, error) {
	genericClient, err := armresources.NewClient("", a.credentials, nil)
	if err != nil {
		return armresources.ClientGetByIDResponse{}, fmt.Errorf("failed to create Azure resources client: %w", err)
	}

	return genericClient.GetByID(context.Background(), resourceID, "2025-01-01", nil)
}
