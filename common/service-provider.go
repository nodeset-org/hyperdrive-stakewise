package swcommon

import (
	"context"
	"fmt"
	"reflect"

	"github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
	"github.com/rocket-pool/node-manager-core/wallet"
)

// Provides the StakeWise module config and resources
type IStakeWiseConfigProvider interface {
	// Gets the StakeWise config
	GetConfig() *swconfig.StakeWiseConfig

	// Gets the StakeWise resources
	GetResources() *swconfig.MergedResources
}

// Provides the StakeWise wallet
type IStakeWiseWalletProvider interface {
	// Gets the wallet
	GetWallet() *Wallet
}

// Provides the deposit data manager
type IDepositDataManagerProvider interface {
	// Gets the deposit data manager
	GetDepositDataManager() *DepositDataManager
}

// Provides requirements for the StakeWise daemon
type IStakeWiseRequirementsProvider interface {
	RequireStakewiseWalletReady(ctx context.Context, status wallet.WalletStatus) error
	WaitForStakewiseWallet(ctx context.Context) error
}

type IStakeWiseServiceProvider interface {
	IStakeWiseConfigProvider
	IStakeWiseWalletProvider
	IDepositDataManagerProvider
	IStakeWiseRequirementsProvider

	services.IModuleServiceProvider
}

type stakeWiseServiceProvider struct {
	services.IModuleServiceProvider
	swCfg              *swconfig.StakeWiseConfig
	wallet             *Wallet
	resources          *swconfig.MergedResources
	depositDataManager *DepositDataManager
}

// Create a new service provider with Stakewise daemon-specific features
func NewStakeWiseServiceProvider(sp services.IModuleServiceProvider, settingsList []*swconfig.StakeWiseSettings) (IStakeWiseServiceProvider, error) {
	// Create the resources
	swCfg, ok := sp.GetModuleConfig().(*swconfig.StakeWiseConfig)
	if !ok {
		return nil, fmt.Errorf("stakewise config is not the correct type, it's a %s", reflect.TypeOf(swCfg))
	}
	hdCfg := sp.GetHyperdriveConfig()
	hdRes := sp.GetHyperdriveResources()

	// Get the resources from the selected network
	var selectedResources *swconfig.MergedResources
	for _, network := range settingsList {
		if network.Key == hdCfg.Network.Value {
			selectedResources = &swconfig.MergedResources{
				MergedResources:    hdRes,
				StakeWiseResources: network.StakeWiseResources,
			}
			break
		}
	}
	if selectedResources == nil {
		return nil, fmt.Errorf("no stakewise resources found for selected network [%s]", hdCfg.Network.Value)
	}

	return NewStakeWiseServiceProviderFromCustomServices(sp, swCfg, selectedResources)
}

// Create a new service provider with Stakewise daemon-specific features, using custom services instead of loading them from the module service provider.
func NewStakeWiseServiceProviderFromCustomServices(sp services.IModuleServiceProvider, cfg *swconfig.StakeWiseConfig, resources *swconfig.MergedResources) (IStakeWiseServiceProvider, error) {
	// Create the wallet
	wallet, err := NewWallet(sp)
	if err != nil {
		return nil, fmt.Errorf("error initializing wallet: %w", err)
	}

	// Make the provider
	stakewiseSp := &stakeWiseServiceProvider{
		IModuleServiceProvider: sp,
		swCfg:                  cfg,
		wallet:                 wallet,
		resources:              resources,
	}

	// Create the deposit data manager
	ddMgr, err := NewDepositDataManager(stakewiseSp)
	if err != nil {
		return nil, fmt.Errorf("error initializing deposit data manager: %w", err)
	}
	stakewiseSp.depositDataManager = ddMgr
	return stakewiseSp, nil
}

func (s *stakeWiseServiceProvider) GetConfig() *swconfig.StakeWiseConfig {
	return s.swCfg
}

func (s *stakeWiseServiceProvider) GetResources() *swconfig.MergedResources {
	return s.resources
}

func (s *stakeWiseServiceProvider) GetWallet() *Wallet {
	return s.wallet
}

func (s *stakeWiseServiceProvider) GetDepositDataManager() *DepositDataManager {
	return s.depositDataManager
}
