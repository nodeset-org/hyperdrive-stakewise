package swcommon

import (
	"fmt"
	"reflect"

	"github.com/nodeset-org/hyperdrive-daemon/module-utils/services"
	swconfig "github.com/nodeset-org/hyperdrive-stakewise/shared/config"
)

type StakeWiseServiceProvider struct {
	*services.ServiceProvider
	swCfg              *swconfig.StakeWiseConfig
	wallet             *Wallet
	resources          *swconfig.StakewiseResources
	depositDataManager *DepositDataManager
}

// Create a new service provider with Stakewise daemon-specific features
func NewStakeWiseServiceProvider(sp *services.ServiceProvider) (*StakeWiseServiceProvider, error) {
	// Create the resources
	swCfg, ok := sp.GetModuleConfig().(*swconfig.StakeWiseConfig)
	if !ok {
		return nil, fmt.Errorf("stakewise config is not the correct type, it's a %s", reflect.TypeOf(swCfg))
	}
	res := swCfg.GetStakeWiseResources()

	return NewStakeWiseServiceProviderFromCustomServices(sp, swCfg, res)
}

// Create a new service provider with Stakewise daemon-specific features, using custom services instead of loading them from the module service provider.
func NewStakeWiseServiceProviderFromCustomServices(sp *services.ServiceProvider, cfg *swconfig.StakeWiseConfig, resources *swconfig.StakewiseResources) (*StakeWiseServiceProvider, error) {
	// Create the wallet
	wallet, err := NewWallet(sp)
	if err != nil {
		return nil, fmt.Errorf("error initializing wallet: %w", err)
	}

	// Make the provider
	stakewiseSp := &StakeWiseServiceProvider{
		ServiceProvider: sp,
		swCfg:           cfg,
		wallet:          wallet,
		resources:       resources,
	}

	// Create the deposit data manager
	ddMgr, err := NewDepositDataManager(stakewiseSp)
	if err != nil {
		return nil, fmt.Errorf("error initializing deposit data manager: %w", err)
	}
	stakewiseSp.depositDataManager = ddMgr
	return stakewiseSp, nil
}

func (s *StakeWiseServiceProvider) GetModuleConfig() *swconfig.StakeWiseConfig {
	return s.swCfg
}

func (s *StakeWiseServiceProvider) GetWallet() *Wallet {
	return s.wallet
}

func (s *StakeWiseServiceProvider) GetResources() *swconfig.StakewiseResources {
	return s.resources
}

func (s *StakeWiseServiceProvider) GetDepositDataManager() *DepositDataManager {
	return s.depositDataManager
}
