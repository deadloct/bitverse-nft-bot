package api

import (
	"github.com/deadloct/immutablex-cli/lib/assets"
	"github.com/deadloct/immutablex-cli/lib/collections"
	"github.com/deadloct/immutablex-cli/lib/orders"
)

type ClientsManager struct {
	AssetsClient      assets.Client
	CollectionsClient collections.Client
	OrdersClient      orders.Client
}

func NewClientsManager() *ClientsManager {
	return &ClientsManager{
		AssetsClient:      assets.NewClient(assets.NewClientConfig("")),
		CollectionsClient: collections.NewClient(collections.NewClientConfig("")),
		OrdersClient:      orders.NewClient(orders.NewClientConfig("")),
	}
}

func (cm *ClientsManager) Start() error {
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

func (cm *ClientsManager) Stop() {
	cm.AssetsClient.Stop()
	cm.CollectionsClient.Stop()
	cm.OrdersClient.Stop()
}
