package adapters

import (
	productDomain "github.com/kamil5b/go-ptse-monolith/internal/modules/product/domain"
	productv1 "github.com/kamil5b/go-ptse-monolith/internal/modules/product/proto"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

// DomainProductToPBProduct converts a domain Product to a protobuf Product
func DomainProductToPBProduct(domain *productDomain.Product) *productv1.Product {
	if domain == nil {
		return nil
	}

	pb := &productv1.Product{
		Id:          domain.ID,
		Name:        domain.Name,
		Description: domain.Description,
		CreatedBy:   domain.CreatedBy,
	}

	if !domain.CreatedAt.IsZero() {
		pb.CreatedAt = &timestamppb.Timestamp{
			Seconds: domain.CreatedAt.Unix(),
			Nanos:   int32(domain.CreatedAt.Nanosecond()),
		}
	}

	if domain.UpdatedAt != nil && !domain.UpdatedAt.IsZero() {
		pb.UpdatedAt = &timestamppb.Timestamp{
			Seconds: domain.UpdatedAt.Unix(),
			Nanos:   int32(domain.UpdatedAt.Nanosecond()),
		}
	}

	if domain.UpdatedBy != nil {
		pb.UpdatedBy = domain.UpdatedBy
	}

	if domain.DeletedAt != nil && !domain.DeletedAt.IsZero() {
		pb.DeletedAt = &timestamppb.Timestamp{
			Seconds: domain.DeletedAt.Unix(),
			Nanos:   int32(domain.DeletedAt.Nanosecond()),
		}
	}

	if domain.DeletedBy != nil {
		pb.DeletedBy = domain.DeletedBy
	}

	return pb
}

// PBProductToDomainProduct converts a protobuf Product to a domain Product
func PBProductToDomainProduct(pb *productv1.Product) *productDomain.Product {
	if pb == nil {
		return nil
	}

	product := &productDomain.Product{
		ID:          pb.GetId(),
		Name:        pb.GetName(),
		Description: pb.GetDescription(),
		CreatedBy:   pb.GetCreatedBy(),
	}

	if pb.CreatedAt != nil {
		product.CreatedAt = pb.CreatedAt.AsTime()
	}

	if pb.UpdatedAt != nil {
		updatedAt := pb.UpdatedAt.AsTime()
		product.UpdatedAt = &updatedAt
	}

	if pb.UpdatedBy != nil {
		product.UpdatedBy = pb.UpdatedBy
	}

	if pb.DeletedAt != nil {
		deletedAt := pb.DeletedAt.AsTime()
		product.DeletedAt = &deletedAt
	}

	if pb.DeletedBy != nil {
		product.DeletedBy = pb.DeletedBy
	}

	return product
}

// PBCreateProductRequestToDomainRequest converts protobuf request to domain request
func PBCreateProductRequestToDomainRequest(pb *productv1.CreateProductRequest) *productDomain.CreateProductRequest {
	if pb == nil {
		return nil
	}

	return &productDomain.CreateProductRequest{
		Name:        pb.GetName(),
		Description: pb.GetDescription(),
	}
}

// PBUpdateProductRequestToDomainRequest converts protobuf request to domain request
func PBUpdateProductRequestToDomainRequest(pb *productv1.UpdateProductRequest) *productDomain.UpdateProductRequest {
	if pb == nil {
		return nil
	}

	req := &productDomain.UpdateProductRequest{
		ID: pb.GetId(),
	}

	if pb.Name != nil {
		req.Name = *pb.Name
	}

	if pb.Description != nil {
		req.Description = *pb.Description
	}

	return req
}
