package service_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dcm-project/catalog-manager/internal/service"
	"github.com/dcm-project/catalog-manager/internal/store"
	"github.com/dcm-project/catalog-manager/internal/store/model"
)

// Mock ServiceTypeStore for testing
type mockServiceTypeStore struct {
	listFunc   func(ctx context.Context, opts *store.ServiceTypeListOptions) (*store.ServiceTypeListResult, error)
	createFunc func(ctx context.Context, serviceType model.ServiceType) (*model.ServiceType, error)
	getFunc    func(ctx context.Context, id string) (*model.ServiceType, error)
}

func (m *mockServiceTypeStore) List(ctx context.Context, opts *store.ServiceTypeListOptions) (*store.ServiceTypeListResult, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, opts)
	}
	return &store.ServiceTypeListResult{
		ServiceTypes: []model.ServiceType{},
	}, nil
}

func (m *mockServiceTypeStore) Create(ctx context.Context, serviceType model.ServiceType) (*model.ServiceType, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, serviceType)
	}
	return &serviceType, nil
}

func (m *mockServiceTypeStore) Get(ctx context.Context, id string) (*model.ServiceType, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return &model.ServiceType{}, nil
}

// Mock Store
type mockStore struct {
	serviceTypeStore store.ServiceTypeStore
}

func (m *mockStore) ServiceType() store.ServiceTypeStore {
	return m.serviceTypeStore
}

func (m *mockStore) CatalogItem() store.CatalogItemStore {
	return nil
}

func (m *mockStore) CatalogItemInstance() store.CatalogItemInstanceStore {
	return nil
}

func (m *mockStore) Close() error {
	return nil
}

var _ = Describe("ServiceType Service", func() {
	var (
		ctx         context.Context
		mockSTStore *mockServiceTypeStore
		mockStr     store.Store
		svc         service.Service
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockSTStore = &mockServiceTypeStore{}
		mockStr = &mockStore{serviceTypeStore: mockSTStore}
		svc = service.NewService(mockStr)
	})

	Describe("Create", func() {
		Context("with valid allowed service types", func() {
			It("should create a service type with 'vm'", func() {
				mockSTStore.createFunc = func(ctx context.Context, st model.ServiceType) (*model.ServiceType, error) {
					Expect(st.ServiceType).To(Equal("vm"))
					Expect(st.ID).ToNot(BeEmpty())
					Expect(st.Path).To(Equal("service-types/" + st.ID))
					return &st, nil
				}

				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "vm",
					Spec:        map[string]any{"vcpu": map[string]any{"count": 2}},
				}

				result, err := svc.ServiceType().Create(ctx, req)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.ServiceType).To(Equal("vm"))
			})

			It("should create a service type with 'container'", func() {
				mockSTStore.createFunc = func(ctx context.Context, st model.ServiceType) (*model.ServiceType, error) {
					Expect(st.ServiceType).To(Equal("container"))
					return &st, nil
				}

				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "container",
					Spec:        map[string]any{"image": "nginx"},
				}

				result, err := svc.ServiceType().Create(ctx, req)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.ServiceType).To(Equal("container"))
			})

			It("should create a service type with 'cluster'", func() {
				mockSTStore.createFunc = func(ctx context.Context, st model.ServiceType) (*model.ServiceType, error) {
					Expect(st.ServiceType).To(Equal("cluster"))
					return &st, nil
				}

				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "cluster",
					Spec:        map[string]any{"nodes": 3},
				}

				result, err := svc.ServiceType().Create(ctx, req)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.ServiceType).To(Equal("cluster"))
			})

			It("should create a service type with 'db'", func() {
				mockSTStore.createFunc = func(ctx context.Context, st model.ServiceType) (*model.ServiceType, error) {
					Expect(st.ServiceType).To(Equal("db"))
					return &st, nil
				}

				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "db",
					Spec:        map[string]any{"engine": "postgres"},
				}

				result, err := svc.ServiceType().Create(ctx, req)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.ServiceType).To(Equal("db"))
			})
		})

		Context("with invalid service types", func() {
			It("should reject 'VM' (uppercase)", func() {
				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "VM",
					Spec:        map[string]any{"vcpu": 2},
				}

				_, err := svc.ServiceType().Create(ctx, req)
				Expect(err).To(Equal(service.ErrInvalidServiceType))
			})

			It("should reject 'database'", func() {
				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "database",
					Spec:        map[string]any{"engine": "mysql"},
				}

				_, err := svc.ServiceType().Create(ctx, req)
				Expect(err).To(Equal(service.ErrInvalidServiceType))
			})

			It("should reject 'invalid-type'", func() {
				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "invalid-type",
					Spec:        map[string]any{"foo": "bar"},
				}

				_, err := svc.ServiceType().Create(ctx, req)
				Expect(err).To(Equal(service.ErrInvalidServiceType))
			})
		})

		Context("with empty spec", func() {
			It("should reject nil spec", func() {
				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "vm",
					Spec:        nil,
				}

				_, err := svc.ServiceType().Create(ctx, req)
				Expect(err).To(Equal(service.ErrEmptySpec))
			})

			It("should reject empty spec map", func() {
				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "vm",
					Spec:        map[string]any{},
				}

				_, err := svc.ServiceType().Create(ctx, req)
				Expect(err).To(Equal(service.ErrEmptySpec))
			})
		})

		Context("with ID validation", func() {
			It("should generate UUID when ID is not provided", func() {
				var capturedID string
				mockSTStore.createFunc = func(ctx context.Context, st model.ServiceType) (*model.ServiceType, error) {
					capturedID = st.ID
					Expect(st.ID).ToNot(BeEmpty())
					Expect(st.Path).To(Equal("service-types/" + st.ID))
					return &st, nil
				}

				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "vm",
					Spec:        map[string]any{"vcpu": 2},
				}

				result, err := svc.ServiceType().Create(ctx, req)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Uid).ToNot(BeNil())
				Expect(*result.Uid).To(Equal(capturedID))
				Expect(*result.Path).To(Equal("service-types/" + capturedID))
			})

			It("should use valid user-provided ID (DNS-1123)", func() {
				userID := "my-service-type"
				mockSTStore.createFunc = func(ctx context.Context, st model.ServiceType) (*model.ServiceType, error) {
					Expect(st.ID).To(Equal(userID))
					Expect(st.Path).To(Equal("service-types/" + userID))
					return &st, nil
				}

				req := &service.CreateServiceTypeRequest{
					ID:          &userID,
					ApiVersion:  "v1alpha1",
					ServiceType: "vm",
					Spec:        map[string]any{"vcpu": 2},
				}

				result, err := svc.ServiceType().Create(ctx, req)
				Expect(err).ToNot(HaveOccurred())
				Expect(*result.Uid).To(Equal(userID))
			})

			It("should reject invalid ID (uppercase)", func() {
				invalidID := "MyServiceType"
				req := &service.CreateServiceTypeRequest{
					ID:          &invalidID,
					ApiVersion:  "v1alpha1",
					ServiceType: "vm",
					Spec:        map[string]any{"vcpu": 2},
				}

				_, err := svc.ServiceType().Create(ctx, req)
				Expect(err).To(Equal(service.ErrInvalidID))
			})

			It("should reject invalid ID (starts with hyphen)", func() {
				invalidID := "-invalid"
				req := &service.CreateServiceTypeRequest{
					ID:          &invalidID,
					ApiVersion:  "v1alpha1",
					ServiceType: "vm",
					Spec:        map[string]any{"vcpu": 2},
				}

				_, err := svc.ServiceType().Create(ctx, req)
				Expect(err).To(Equal(service.ErrInvalidID))
			})

			It("should reject invalid ID (ends with hyphen)", func() {
				invalidID := "invalid-"
				req := &service.CreateServiceTypeRequest{
					ID:          &invalidID,
					ApiVersion:  "v1alpha1",
					ServiceType: "vm",
					Spec:        map[string]any{"vcpu": 2},
				}

				_, err := svc.ServiceType().Create(ctx, req)
				Expect(err).To(Equal(service.ErrInvalidID))
			})
		})

		Context("with store errors", func() {
			It("should map ErrServiceTypeIDTaken", func() {
				mockSTStore.createFunc = func(ctx context.Context, st model.ServiceType) (*model.ServiceType, error) {
					return nil, store.ErrServiceTypeIDTaken
				}

				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "vm",
					Spec:        map[string]any{"vcpu": 2},
				}

				_, err := svc.ServiceType().Create(ctx, req)
				Expect(err).To(Equal(service.ErrServiceTypeIDTaken))
			})

			It("should map ErrServiceTypeServiceTypeTaken", func() {
				mockSTStore.createFunc = func(ctx context.Context, st model.ServiceType) (*model.ServiceType, error) {
					return nil, store.ErrServiceTypeServiceTypeTaken
				}

				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "vm",
					Spec:        map[string]any{"vcpu": 2},
				}

				_, err := svc.ServiceType().Create(ctx, req)
				Expect(err).To(Equal(service.ErrServiceTypeNameTaken))
			})

			It("should propagate unknown errors", func() {
				unknownErr := errors.New("database connection failed")
				mockSTStore.createFunc = func(ctx context.Context, st model.ServiceType) (*model.ServiceType, error) {
					return nil, unknownErr
				}

				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "vm",
					Spec:        map[string]any{"vcpu": 2},
				}

				_, err := svc.ServiceType().Create(ctx, req)
				Expect(err).To(Equal(unknownErr))
			})
		})

		Context("with metadata", func() {
			It("should handle metadata with labels", func() {
				mockSTStore.createFunc = func(ctx context.Context, st model.ServiceType) (*model.ServiceType, error) {
					Expect(st.Metadata.Labels).To(HaveKeyWithValue("env", "prod"))
					Expect(st.Metadata.Labels).To(HaveKeyWithValue("team", "platform"))
					return &st, nil
				}

				labels := map[string]string{"env": "prod", "team": "platform"}
				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "vm",
					Metadata: &struct {
						Labels *map[string]string `json:"labels,omitempty"`
					}{
						Labels: &labels,
					},
					Spec: map[string]any{"vcpu": 2},
				}

				result, err := svc.ServiceType().Create(ctx, req)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Metadata).ToNot(BeNil())
				Expect(result.Metadata.Labels).ToNot(BeNil())
				Expect(*result.Metadata.Labels).To(HaveKeyWithValue("env", "prod"))
			})

			It("should handle nil metadata", func() {
				mockSTStore.createFunc = func(ctx context.Context, st model.ServiceType) (*model.ServiceType, error) {
					Expect(st.Metadata.Labels).To(BeNil())
					return &st, nil
				}

				req := &service.CreateServiceTypeRequest{
					ApiVersion:  "v1alpha1",
					ServiceType: "vm",
					Metadata:    nil,
					Spec:        map[string]any{"vcpu": 2},
				}

				_, err := svc.ServiceType().Create(ctx, req)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Get", func() {
		It("should retrieve a service type", func() {
			mockSTStore.getFunc = func(ctx context.Context, id string) (*model.ServiceType, error) {
				Expect(id).To(Equal("test-id"))
				return &model.ServiceType{
					ID:          "test-id",
					ApiVersion:  "v1alpha1",
					ServiceType: "vm",
					Spec:        map[string]any{"vcpu": 2},
					Path:        "service-types/test-id",
				}, nil
			}

			result, err := svc.ServiceType().Get(ctx, "test-id")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(*result.Uid).To(Equal("test-id"))
			Expect(result.ServiceType).To(Equal("vm"))
		})

		It("should map ErrServiceTypeNotFound", func() {
			mockSTStore.getFunc = func(ctx context.Context, id string) (*model.ServiceType, error) {
				return nil, store.ErrServiceTypeNotFound
			}

			_, err := svc.ServiceType().Get(ctx, "non-existent")
			Expect(err).To(Equal(service.ErrServiceTypeNotFound))
		})
	})

	Describe("List", func() {
		It("should list service types", func() {
			mockSTStore.listFunc = func(ctx context.Context, opts *store.ServiceTypeListOptions) (*store.ServiceTypeListResult, error) {
				return &store.ServiceTypeListResult{
					ServiceTypes: []model.ServiceType{
						{ID: "id1", ServiceType: "vm", ApiVersion: "v1alpha1", Spec: map[string]any{"vcpu": 2}, Path: "service-types/id1"},
						{ID: "id2", ServiceType: "container", ApiVersion: "v1alpha1", Spec: map[string]any{"image": "nginx"}, Path: "service-types/id2"},
					},
					NextPageToken: "token123",
				}, nil
			}

			result, err := svc.ServiceType().List(ctx, &service.ServiceTypeListOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(result.ServiceTypes).To(HaveLen(2))
			Expect(result.NextPageToken).To(Equal("token123"))
		})

		It("should handle empty list", func() {
			mockSTStore.listFunc = func(ctx context.Context, opts *store.ServiceTypeListOptions) (*store.ServiceTypeListResult, error) {
				return &store.ServiceTypeListResult{
					ServiceTypes:  []model.ServiceType{},
					NextPageToken: "",
				}, nil
			}

			result, err := svc.ServiceType().List(ctx, &service.ServiceTypeListOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(result.ServiceTypes).To(BeEmpty())
			Expect(result.NextPageToken).To(BeEmpty())
		})
	})
})
