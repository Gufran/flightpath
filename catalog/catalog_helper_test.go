package catalog

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
	"testing"
)

var _ ServiceFinder = &ServiceFinderMock{}

type ServicesResult struct {
	services map[string][]string
	meta     *api.QueryMeta
	err      error
}

type ServiceResult struct {
	services []*api.CatalogService
	meta     *api.QueryMeta
	err      error
}

type ServiceFinderMock struct {
	ctx           context.Context
	t             *testing.T
	servicesStack []ServicesResult
	serviceStack  map[string][]ServiceResult
	blockOnFinish bool
}

func NewServiceFinderMock(ctx context.Context, t *testing.T, service map[string][]ServiceResult, services []ServicesResult, blocking bool) *ServiceFinderMock {
	return &ServiceFinderMock{
		ctx:           ctx,
		t:             t,
		servicesStack: services,
		serviceStack:  service,
		blockOnFinish: blocking,
	}
}

func (m *ServiceFinderMock) AssertFulfilled() bool {
	if len(m.servicesStack) > 0 {
		m.t.Errorf("expecting %d more calls to Services method", len(m.serviceStack))
		return false
	}

	if len(m.serviceStack) == 0 {
		return true
	}

	fulfilled := true
	for name, s := range m.serviceStack {
		if len(s) > 0 {
			m.t.Errorf("expecting %d more calls to Service(%s)", len(s), name)
			fulfilled = false
		}
	}
	return fulfilled
}

func (m *ServiceFinderMock) Services(q *api.QueryOptions) (map[string][]string, *api.QueryMeta, error) {
	if len(m.servicesStack) == 0 {
		if m.blockOnFinish {
			<-m.ctx.Done()
			return nil, &api.QueryMeta{
				LastIndex: q.WaitIndex,
			}, nil
		}

		m.t.Errorf("unexpected call to Services. No more expectations")
		return nil, nil, fmt.Errorf("unexpected function call")
	}

	var result ServicesResult
	result, m.servicesStack = m.servicesStack[0], m.servicesStack[1:]
	return result.services, result.meta, result.err
}

func (m *ServiceFinderMock) Service(name string, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error) {
	s, ok := m.serviceStack[name]
	if !ok {
		m.t.Errorf("unexpected Service call with service name %q and tag %q", name, tag)
	}

	// TODO: tag is not taken in account, implement support for
	//   also matching the tag along with the service name

	if len(s) == 0 {
		if m.blockOnFinish {
			<-m.ctx.Done()
			return nil, &api.QueryMeta{
				LastIndex: q.WaitIndex,
			}, nil
		}

		m.t.Error("unexpected call to Service, no more expectations")
		return nil, nil, fmt.Errorf("unexpected function call")
	}

	var result ServiceResult
	result, s = s[0], s[1:]
	m.serviceStack[name] = s
	return result.services, result.meta, result.err
}

type ConnectCALeafResult struct {
	cert *api.LeafCert
	meta *api.QueryMeta
	err  error
}

func NewCertFinderMock(ctx context.Context, t *testing.T, stack map[string][]ConnectCALeafResult, blocking bool) CertFinder {
	return &MockCertFinder{
		ctx:           ctx,
		t:             t,
		stack:         stack,
		blockOnFinish: blocking,
	}
}

var _ CertFinder = &MockCertFinder{}

type MockCertFinder struct {
	t             *testing.T
	ctx           context.Context
	stack         map[string][]ConnectCALeafResult
	blockOnFinish bool
}

func (m *MockCertFinder) ConnectCALeaf(name string, q *api.QueryOptions) (*api.LeafCert, *api.QueryMeta, error) {
	s, ok := m.stack[name]
	if !ok {
		m.t.Errorf("unexpected call to ConnectCALeaf with service name %q", name)
		return nil, nil, fmt.Errorf("unexpected function call")
	}

	if len(s) == 0 {
		if m.blockOnFinish {
			<-m.ctx.Done()
			return nil, &api.QueryMeta{
				LastIndex: q.WaitIndex,
			}, nil
		}

		m.t.Error("unexpected call to ConnectCALeaf, no more expectations")
		return nil, nil, fmt.Errorf("unexpected function call")
	}

	var result ConnectCALeafResult

	result, s = s[0], s[1:]
	m.stack[name] = s

	return result.cert, result.meta, result.err
}
