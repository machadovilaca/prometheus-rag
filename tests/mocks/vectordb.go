package mocks

import (
	"github.com/machadovilaca/prometheus-rag/pkg/prometheus"
	"github.com/machadovilaca/prometheus-rag/pkg/vectordb"
)

type VectorDBMock struct {
	AddMetricMetadataFunc      func(metadata *prometheus.MetricMetadata) error
	BatchAddMetricMetadataFunc func(metadata []*prometheus.MetricMetadata) error
	CreateCollectionFunc       func() error
	DeleteCollectionFunc       func() error
	SearchMetricsFunc          func(query string, limit uint64) ([]*prometheus.MetricMetadata, error)
	CloseFunc                  func() error
}

func NewVectorDBMock() *VectorDBMock {
	return &VectorDBMock{}
}

func (v *VectorDBMock) AddMetricMetadata(metadata *prometheus.MetricMetadata) error {
	if v.AddMetricMetadataFunc != nil {
		return v.AddMetricMetadataFunc(metadata)
	}
	return nil
}

func (v *VectorDBMock) BatchAddMetricMetadata(metadata []*prometheus.MetricMetadata) error {
	if v.BatchAddMetricMetadataFunc != nil {
		return v.BatchAddMetricMetadataFunc(metadata)
	}
	return nil
}

func (v *VectorDBMock) CreateCollection() error {
	if v.CreateCollectionFunc != nil {
		return v.CreateCollectionFunc()
	}
	return nil
}

func (v *VectorDBMock) DeleteCollection() error {
	if v.DeleteCollectionFunc != nil {
		return v.DeleteCollectionFunc()
	}
	return nil
}

func (v *VectorDBMock) SearchMetrics(query string, limit uint64) ([]*prometheus.MetricMetadata, error) {
	if v.SearchMetricsFunc != nil {
		return v.SearchMetricsFunc(query, limit)
	}
	return nil, nil
}

func (v *VectorDBMock) Close() error {
	if v.CloseFunc != nil {
		return v.CloseFunc()
	}
	return nil
}

var _ vectordb.Client = &VectorDBMock{}
