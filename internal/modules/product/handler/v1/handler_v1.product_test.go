package v1

import (
	"context"
	"errors"
	"net/http"
	"testing"

	domain "github.com/kamil5b/go-ptse-monolith/internal/modules/product/domain"
	mockdomain "github.com/kamil5b/go-ptse-monolith/internal/modules/product/domain/mocks"
	ctxmocks "github.com/kamil5b/go-ptse-monolith/internal/shared/context/mocks"

	gomock "github.com/golang/mock/gomock"
)

func TestHandler_Create(t *testing.T) {
	cases := []struct {
		name  string
		setup func(svc *mockdomain.MockService, mc *ctxmocks.MockContext)
	}{
		{
			name: "ok",
			setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
				mc.EXPECT().Bind(gomock.AssignableToTypeOf(&domain.CreateProductRequest{})).Return(nil)
				mc.EXPECT().GetContext().Return(context.Background())
				mc.EXPECT().GetUserID().Return("user1")
				svc.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(&domain.CreateProductRequest{}), "user1").Return(&domain.Product{ID: "p1"}, nil)
				mc.EXPECT().JSON(http.StatusCreated, gomock.Any()).Return(nil)
			},
		},
		{
			name: "bind error",
			setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
				mc.EXPECT().GetContext().Return(context.Background())
				mc.EXPECT().Bind(gomock.AssignableToTypeOf(&domain.CreateProductRequest{})).Return(errors.New("bad input"))
				mc.EXPECT().JSON(http.StatusBadRequest, gomock.Any()).Return(nil)
			},
		},
		{
			name: "service error",
			setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
				mc.EXPECT().Bind(gomock.AssignableToTypeOf(&domain.CreateProductRequest{})).Return(nil)
				mc.EXPECT().GetContext().Return(context.Background())
				mc.EXPECT().GetUserID().Return("user1")
				svc.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(&domain.CreateProductRequest{}), "user1").Return(nil, errors.New("boom"))
				mc.EXPECT().JSON(http.StatusInternalServerError, gomock.Any()).Return(nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mockdomain.NewMockService(ctrl)
			mc := ctxmocks.NewMockContext(ctrl)
			h := NewHandler(svc)

			tc.setup(svc, mc)

			if err := h.Create(mc); err != nil {
				t.Fatalf("Create returned error: %v", err)
			}
		})
	}
}

func TestHandler_Get(t *testing.T) {
	cases := []struct {
		name  string
		setup func(svc *mockdomain.MockService, mc *ctxmocks.MockContext)
	}{
		{
			name: "ok",
			setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
				mc.EXPECT().GetContext().Return(context.Background())
				mc.EXPECT().Param("id").Return("p1")
				svc.EXPECT().Get(gomock.Any(), "p1").Return(&domain.Product{ID: "p1"}, nil)
				mc.EXPECT().JSON(http.StatusOK, gomock.Any()).Return(nil)
			},
		},
		{
			name: "not found",
			setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
				mc.EXPECT().GetContext().Return(context.Background())
				mc.EXPECT().Param("id").Return("p1")
				svc.EXPECT().Get(gomock.Any(), "p1").Return(nil, errors.New("not found"))
				mc.EXPECT().JSON(http.StatusNotFound, gomock.Any()).Return(nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mockdomain.NewMockService(ctrl)
			mc := ctxmocks.NewMockContext(ctrl)
			h := NewHandler(svc)

			tc.setup(svc, mc)

			if err := h.Get(mc); err != nil {
				t.Fatalf("Get returned error: %v", err)
			}
		})
	}
}

func TestHandler_List(t *testing.T) {
	cases := []struct {
		name  string
		setup func(svc *mockdomain.MockService, mc *ctxmocks.MockContext)
	}{
		{
			name: "ok",
			setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
				mc.EXPECT().GetContext().Return(context.Background())
				svc.EXPECT().List(gomock.Any()).Return([]domain.Product{{ID: "p1"}}, nil)
				mc.EXPECT().JSON(http.StatusOK, gomock.Any()).Return(nil)
			},
		},
		{
			name: "service error",
			setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
				mc.EXPECT().GetContext().Return(context.Background())
				svc.EXPECT().List(gomock.Any()).Return(nil, errors.New("boom"))
				mc.EXPECT().JSON(http.StatusInternalServerError, gomock.Any()).Return(nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mockdomain.NewMockService(ctrl)
			mc := ctxmocks.NewMockContext(ctrl)
			h := NewHandler(svc)

			tc.setup(svc, mc)

			if err := h.List(mc); err != nil {
				t.Fatalf("List returned error: %v", err)
			}
		})
	}
}

func TestHandler_Update(t *testing.T) {
	cases := []struct {
		name  string
		setup func(svc *mockdomain.MockService, mc *ctxmocks.MockContext)
	}{
		{
			name: "ok",
			setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
				mc.EXPECT().Param("id").Return("p1")
				mc.EXPECT().Bind(gomock.AssignableToTypeOf(&domain.UpdateProductRequest{})).Return(nil)
				mc.EXPECT().GetContext().Return(context.Background())
				mc.EXPECT().GetUserID().Return("user1")
				svc.EXPECT().Update(gomock.Any(), gomock.AssignableToTypeOf(&domain.UpdateProductRequest{}), "user1").Return(&domain.Product{ID: "p1"}, nil)
				mc.EXPECT().JSON(http.StatusOK, gomock.Any()).Return(nil)
			},
		},
		{
			name: "bind error",
			setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
				mc.EXPECT().GetContext().Return(context.Background())
				mc.EXPECT().Param("id").Return("p1")
				mc.EXPECT().Bind(gomock.AssignableToTypeOf(&domain.UpdateProductRequest{})).Return(errors.New("bad"))
				mc.EXPECT().JSON(http.StatusBadRequest, gomock.Any()).Return(nil)
			},
		},
		{
			name: "service error",
			setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
				mc.EXPECT().Param("id").Return("p1")
				mc.EXPECT().Bind(gomock.AssignableToTypeOf(&domain.UpdateProductRequest{})).Return(nil)
				mc.EXPECT().GetContext().Return(context.Background())
				mc.EXPECT().GetUserID().Return("user1")
				svc.EXPECT().Update(gomock.Any(), gomock.AssignableToTypeOf(&domain.UpdateProductRequest{}), "user1").Return(nil, errors.New("boom"))
				mc.EXPECT().JSON(http.StatusInternalServerError, gomock.Any()).Return(nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mockdomain.NewMockService(ctrl)
			mc := ctxmocks.NewMockContext(ctrl)
			h := NewHandler(svc)

			tc.setup(svc, mc)

			if err := h.Update(mc); err != nil {
				t.Fatalf("Update returned error: %v", err)
			}
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	cases := []struct {
		name  string
		setup func(svc *mockdomain.MockService, mc *ctxmocks.MockContext)
	}{
		{
			name: "ok_with_user",
			setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
				mc.EXPECT().GetContext().Return(context.Background())
				mc.EXPECT().Param("id").Return("p1")
				mc.EXPECT().Get("user_id").Return("user1")
				svc.EXPECT().Delete(gomock.Any(), "p1", "user1").Return(nil)
				mc.EXPECT().JSON(http.StatusOK, gomock.Any()).Return(nil)
			},
		},
		{
			name: "ok_without_user",
			setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
				mc.EXPECT().GetContext().Return(context.Background())
				mc.EXPECT().Param("id").Return("p1")
				mc.EXPECT().Get("user_id").Return(nil)
				svc.EXPECT().Delete(gomock.Any(), "p1", "").Return(nil)
				mc.EXPECT().JSON(http.StatusOK, gomock.Any()).Return(nil)
			},
		},
		{
			name: "service error",
			setup: func(svc *mockdomain.MockService, mc *ctxmocks.MockContext) {
				mc.EXPECT().GetContext().Return(context.Background())
				mc.EXPECT().Param("id").Return("p1")
				mc.EXPECT().Get("user_id").Return("user1")
				svc.EXPECT().Delete(gomock.Any(), "p1", "user1").Return(errors.New("boom"))
				mc.EXPECT().JSON(http.StatusInternalServerError, gomock.Any()).Return(nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mockdomain.NewMockService(ctrl)
			mc := ctxmocks.NewMockContext(ctrl)
			h := NewHandler(svc)

			tc.setup(svc, mc)

			if err := h.Delete(mc); err != nil {
				t.Fatalf("Delete returned error: %v", err)
			}
		})
	}
}
