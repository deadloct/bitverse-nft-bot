package main

import (
	"github.com/deadloct/immutablex-cli/lib/assets"
	"github.com/deadloct/immutablex-cli/lib/collections"
	"github.com/deadloct/immutablex-cli/lib/orders"
)

type ClientManager struct {
	AssetsClient      assets.Client
	CollectionsClient collections.Client
	OrdersClient      orders.Client
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		AssetsClient:      assets.NewClient(assets.NewClientConfig("")),
		CollectionsClient: collections.NewClient(collections.NewClientConfig("")),
		OrdersClient:      orders.NewClient(orders.NewClientConfig("")),
	}
}

func (cm *ClientManager) Start() error {
	if err := cm.AssetsClient.Start(); err != nil {
		return err
	}

	if err := cm.CollectionsClient.Start(); err != nil {
		return err
	}

	if err := cm.OrdersClient.Start(); err != nil {
		return err
	}

	return nil
}

func (cm *ClientManager) Stop() {
	cm.AssetsClient.Stop()
	cm.CollectionsClient.Stop()
	cm.OrdersClient.Stop()
}
