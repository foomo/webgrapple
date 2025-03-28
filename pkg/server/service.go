package server

import "github.com/foomo/webgrapple/pkg/vo"

type Service struct {
	r *registry
}

func (s *Service) Upsert(services []*vo.Service) *vo.ServiceError {
	errUpsert := s.r.upsert(services)
	if errUpsert != nil {
		return &vo.ServiceError{
			Err: errUpsert.Error(),
		}
	}
	return nil
}

func (s *Service) Remove(serviceIDs []vo.ServiceID) *vo.ServiceError {
	errRemove := s.r.remove(serviceIDs)
	if errRemove != nil {
		return &vo.ServiceError{
			Err: errRemove.Error(),
		}
	}
	return nil
}
