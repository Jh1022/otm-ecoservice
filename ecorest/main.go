package ecorest

import (
	"github.com/azavea/ecobenefits/ecorest/cache"
	"github.com/azavea/ecobenefits/ecorest/config"
	"github.com/azavea/ecobenefits/ecorest/endpoints"
	"net/url"
)

type restManager struct {
	ITreeCodesGET      (func() *endpoints.ITreeCodes)
	EcoGET             (func(url.Values) (*endpoints.BenefitsWrapper, error))
	EcoSummaryPOST     (func(*endpoints.SummaryPostData) (*endpoints.BenefitsWrapper, error))
	EcoScenarioPOST    (func(*endpoints.ScenarioPostData) (*endpoints.Scenario, error))
	InvalidateCacheGET (func())
}

func GetManager(cfg config.Config) *restManager {
	ecoCache, invalidateCache := cache.Init(cfg)
	invalidateCache()

	return &restManager{endpoints.ITreeCodesGET(ecoCache),
		endpoints.EcoGET(ecoCache),
		endpoints.EcoSummaryPOST(ecoCache),
		endpoints.EcoScenarioPOST(ecoCache),
		invalidateCache}
}
