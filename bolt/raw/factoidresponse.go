package raw

import (
	pandora "../.."
)

var _ pandora.FactoidResponseService = &FactoidResponseService{}

type FactoidResponseService struct{}

func (s *FactoidResponseService) FactoidResponse(id uint64) (r *pandora.FactoidResponse, ok bool) {
	return
}

func (s *FactoidResponseService) Create(r *pandora.FactoidResponse) (id uint64, err error) {
	return
}

func (s *FactoidResponseService) Put(id uint64, r *pandora.FactoidResponse) (err error) {
	return
}

func (s *FactoidResponseService) Delete(id uint64) (err error) {
	return
}
