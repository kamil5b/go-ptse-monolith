package adapters

import (
	"testing"
	"time"

	productDomain "github.com/kamil5b/go-ptse-monolith/internal/modules/product/domain"
	productv1 "github.com/kamil5b/go-ptse-monolith/internal/modules/product/proto"
	"github.com/stretchr/testify/assert"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func TestDomainProductToPBProduct(t *testing.T) {
	now := time.Now()

	domainProduct := &productDomain.Product{
		ID:          "prod-123",
		Name:        "Test Product",
		Description: "A test product",
		CreatedBy:   "user-1",
		CreatedAt:   now,
		UpdatedBy:   ptr("user-2"),
		UpdatedAt:   ptr(now.Add(1 * time.Hour)),
	}

	pbProduct := DomainProductToPBProduct(domainProduct)

	assert.NotNil(t, pbProduct)
	assert.Equal(t, "prod-123", pbProduct.GetId())
	assert.Equal(t, "Test Product", pbProduct.GetName())
	assert.Equal(t, "A test product", pbProduct.GetDescription())
	assert.Equal(t, "user-1", pbProduct.GetCreatedBy())
	assert.NotNil(t, pbProduct.GetCreatedAt())
	assert.NotNil(t, pbProduct.GetUpdatedAt())
	assert.Equal(t, "user-2", pbProduct.GetUpdatedBy())
}

func TestDomainProductToPBProductNil(t *testing.T) {
	pbProduct := DomainProductToPBProduct(nil)
	assert.Nil(t, pbProduct)
}

func TestPBProductToDomainProduct(t *testing.T) {
	now := timestamppb.Now()

	pbProduct := &productv1.Product{
		Id:          "prod-123",
		Name:        "Test Product",
		Description: "A test product",
		CreatedBy:   "user-1",
		CreatedAt:   now,
		UpdatedBy:   ptr("user-2"),
		UpdatedAt:   now,
	}

	domainProduct := PBProductToDomainProduct(pbProduct)

	assert.NotNil(t, domainProduct)
	assert.Equal(t, "prod-123", domainProduct.ID)
	assert.Equal(t, "Test Product", domainProduct.Name)
	assert.Equal(t, "A test product", domainProduct.Description)
	assert.Equal(t, "user-1", domainProduct.CreatedBy)
	assert.NotZero(t, domainProduct.CreatedAt)
	assert.NotNil(t, domainProduct.UpdatedAt)
	assert.Equal(t, "user-2", *domainProduct.UpdatedBy)
}

func TestPBProductToDomainProductNil(t *testing.T) {
	domainProduct := PBProductToDomainProduct(nil)
	assert.Nil(t, domainProduct)
}

func TestPBCreateProductRequestToDomainRequest(t *testing.T) {
	pbReq := &productv1.CreateProductRequest{
		Name:        "New Product",
		Description: "Description",
	}

	domainReq := PBCreateProductRequestToDomainRequest(pbReq)

	assert.NotNil(t, domainReq)
	assert.Equal(t, "New Product", domainReq.Name)
	assert.Equal(t, "Description", domainReq.Description)
}

func TestPBCreateProductRequestToDomainRequestNil(t *testing.T) {
	domainReq := PBCreateProductRequestToDomainRequest(nil)
	assert.Nil(t, domainReq)
}

func TestPBUpdateProductRequestToDomainRequest(t *testing.T) {
	newName := "Updated Name"
	pbReq := &productv1.UpdateProductRequest{
		Id:   "prod-123",
		Name: &newName,
	}

	domainReq := PBUpdateProductRequestToDomainRequest(pbReq)

	assert.NotNil(t, domainReq)
	assert.Equal(t, "prod-123", domainReq.ID)
	assert.Equal(t, "Updated Name", domainReq.Name)
}

func TestPBUpdateProductRequestToDomainRequestNil(t *testing.T) {
	domainReq := PBUpdateProductRequestToDomainRequest(nil)
	assert.Nil(t, domainReq)
}

// Helper function for pointer conversion
func ptr[T any](v T) *T {
	return &v
}
