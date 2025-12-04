package grpc

import (
	"context"

	productDomain "github.com/kamil5b/go-ptse-monolith/internal/modules/product/domain"
	productv1 "github.com/kamil5b/go-ptse-monolith/internal/modules/product/proto"
	"github.com/kamil5b/go-ptse-monolith/internal/modules/product/proto/adapters"
	grpcAdapter "github.com/kamil5b/go-ptse-monolith/internal/transports/grpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GRPCHandler implements the Product gRPC service
type GRPCHandler struct {
	service productDomain.Service
	productv1.UnimplementedProductServiceServer
}

// NewGRPCHandler creates a new GRPCHandler
func NewGRPCHandler(service productDomain.Service) *GRPCHandler {
	return &GRPCHandler{service: service}
}

// Create creates a new product
func (h *GRPCHandler) Create(ctx context.Context, req *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
	createdBy := ""
	if uid := ctx.Value("user_id"); uid != nil {
		createdBy = uid.(string)
	}

	createReq := adapters.PBCreateProductRequestToDomainRequest(req)

	product, err := h.service.Create(ctx, createReq, createdBy)
	if err != nil {
		return nil, err
	}

	return &productv1.CreateProductResponse{
		Product: adapters.DomainProductToPBProduct(product),
	}, nil
}

// Get retrieves a product by ID
func (h *GRPCHandler) Get(ctx context.Context, req *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	product, err := h.service.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &productv1.GetProductResponse{
		Product: adapters.DomainProductToPBProduct(product),
	}, nil
}

// List retrieves all products
func (h *GRPCHandler) List(ctx context.Context, _ *emptypb.Empty) (*productv1.ListProductResponse, error) {
	products, err := h.service.List(ctx)
	if err != nil {
		return nil, err
	}

	pbProducts := make([]*productv1.Product, len(products))
	for i, p := range products {
		pbProducts[i] = adapters.DomainProductToPBProduct(&p)
	}

	return &productv1.ListProductResponse{
		Products: pbProducts,
	}, nil
}

// Update updates an existing product
func (h *GRPCHandler) Update(ctx context.Context, req *productv1.UpdateProductRequest) (*productv1.UpdateProductResponse, error) {
	updatedBy := ""
	if uid := ctx.Value("user_id"); uid != nil {
		updatedBy = uid.(string)
	}

	updateReq := adapters.PBUpdateProductRequestToDomainRequest(req)

	product, err := h.service.Update(ctx, updateReq, updatedBy)
	if err != nil {
		return nil, err
	}

	return &productv1.UpdateProductResponse{
		Product: adapters.DomainProductToPBProduct(product),
	}, nil
}

// Delete deletes a product
func (h *GRPCHandler) Delete(ctx context.Context, req *productv1.DeleteProductRequest) (*emptypb.Empty, error) {
	deletedBy := ""
	if uid := ctx.Value("user_id"); uid != nil {
		deletedBy = uid.(string)
	}

	err := h.service.Delete(ctx, req.GetId(), deletedBy)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RegisterService registers the Product service with the gRPC server
func RegisterService(h *GRPCHandler) grpcAdapter.ServiceRegistrar {
	return func(s *grpc.Server) {
		productv1.RegisterProductServiceServer(s, h)
	}
}
