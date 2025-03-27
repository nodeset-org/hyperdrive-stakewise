package operator

import "math/big"

// https://github.com/stakewise/v3-operator/blob/eb02dd26d2337d39072f924386a3e97a16852a16/src/common/app_state.py#L9
type oraclesCache struct {
	checkpointBlock     *big.Int
	config              map[string]any
	validatorsThreshold int
	rewardsThreshold    int
}
