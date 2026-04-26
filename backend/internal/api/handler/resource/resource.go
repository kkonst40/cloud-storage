package resource

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"path"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/kkonst40/cloud-storage/backend/internal/api/dto"
	"github.com/kkonst40/cloud-storage/backend/internal/api/dto/resource"
	"github.com/kkonst40/cloud-storage/backend/internal/api/handler"
)

const pkg = "ResourceHandler"

type Handler struct {
	s3Service S3Service
}

type S3Service interface {
	ObjectInfo(ctx context.Context, userId int64, path string) (resource.Response, error)
	Object(ctx context.Context, userId int64, path string) (io.Reader, string, error)
	StoreObject(ctx context.Context, userId int64, files []resource.FilePayload, path string) ([]resource.Response, error)
	Delete(ctx context.Context, userId int64, path string) error

	Move(ctx context.Context, userId int64, to, from string) (resource.Response, error)
	Search(ctx context.Context, userId int64, query string) []resource.Response
	MakeZip(ctx context.Context, userId int64, path string) (*bytes.Buffer, error)

	StoreDirectory(ctx context.Context, userId int64, path string) (resource.Response, error)
	PaginateDirectory(ctx context.Context, userId int64, path string) []resource.Response
}

func New(s3Service S3Service) *Handler {
	return &Handler{
		s3Service: s3Service,
	}
}

func (h *Handler) Show(ctx fiber.Ctx) error {
	const op = "Show"

	handler.SetCommonHeaders(ctx)
	userId := handler.RequestedUserId(ctx)

	path := sanitizePath(ctx.Query("path"))

	data, err := h.s3Service.ObjectInfo(ctx.RequestCtx(), userId, path)
	if err != nil {
		slog.Error(err.Error(), "pkg", pkg, "op", op)

		return ctx.Status(fiber.StatusNotFound).JSON(&dto.ErrorResponse{Message: handler.MessageNotFound})
	}

	ctx.Status(fiber.StatusOK)

	return ctx.JSON(data)
}

func (h *Handler) Store(ctx fiber.Ctx) error {
	const op = "Store"

	handler.SetCommonHeaders(ctx)
	userId := handler.RequestedUserId(ctx)

	form, err := ctx.MultipartForm()
	if err != nil {
		slog.Error(err.Error(), "pkg", pkg, "op", op)

		return ctx.Status(fiber.StatusBadRequest).JSON(&dto.ErrorResponse{Message: handler.MessageBadRequest})
	}

	fileHeaders := form.File["object"]
	if len(fileHeaders) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(&dto.ErrorResponse{Message: handler.MessageBadRequest})
	}

	files := make([]resource.FilePayload, 0, len(fileHeaders))

	for _, fileHeader := range fileHeaders {
		fn1 := strings.Split(fileHeader.Header.Get("Content-Disposition"), "; ")[2]
		fn2, _ := strings.CutPrefix(fn1, "filename=")
		relPath := sanitizePath(strings.Trim(fn2, "\""))

		f, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer f.Close()

		files = append(files, resource.FilePayload{
			Name:    relPath,
			Size:    fileHeader.Size,
			Content: f,
		})
	}

	path := sanitizePath(ctx.FormValue("path"))

	data, err := h.s3Service.StoreObject(ctx.RequestCtx(), userId, files, path)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(&dto.ErrorResponse{Message: handler.MessageServerError})
	}

	ctx.Status(fiber.StatusCreated)

	return ctx.JSON(data)
}

func (h *Handler) Delete(ctx fiber.Ctx) error {
	const op = "Delete"

	handler.SetCommonHeaders(ctx)
	userId := handler.RequestedUserId(ctx)

	path := sanitizePath(ctx.Query("path"))

	err := h.s3Service.Delete(ctx.RequestCtx(), userId, path)
	if err != nil {
		slog.Error(err.Error(), "pkg", pkg, "op", op)

		return ctx.Status(fiber.StatusNotFound).JSON(&dto.ErrorResponse{Message: handler.MessageNotFound})
	}

	ctx.Status(fiber.StatusNoContent)

	return nil
}

func (h *Handler) Download(ctx fiber.Ctx) error {
	const op = "Download"

	ctx.Accepts("application/json")
	ctx.Set(fiber.HeaderAccept, "application/json")

	userId := handler.RequestedUserId(ctx)
	path := sanitizePath(ctx.Query("path"))

	// zip all in directory
	if strings.HasSuffix(path, "/") {
		buf, err := h.s3Service.MakeZip(ctx.RequestCtx(), userId, path)
		if err != nil {
			slog.Error(err.Error(), "pkg", pkg, "op", op)

			return ctx.Status(fiber.StatusInternalServerError).JSON(
				&dto.ErrorResponse{Message: handler.MessageServerError},
			)
		}

		ctx.Status(fiber.StatusOK)
		ctx.Set(fiber.HeaderContentType, "application/octet-stream")
		ctx.Set(fiber.HeaderContentDisposition, "attachment; filename=\"archive.zip\"")

		return ctx.Send(buf.Bytes())
	}

	object, name, err := h.s3Service.Object(ctx.RequestCtx(), userId, path)
	if err != nil {
		slog.Error(err.Error(), "pkg", pkg, "op", op)

		return ctx.Status(fiber.StatusInternalServerError).JSON(
			&dto.ErrorResponse{Message: handler.MessageServerError},
		)
	}

	ctx.Set(fiber.HeaderContentType, "application/octet-stream")
	ctx.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, name))

	return ctx.SendStream(object)
}

func (h *Handler) Search(ctx fiber.Ctx) error {
	handler.SetCommonHeaders(ctx)
	userId := handler.RequestedUserId(ctx)

	query := ctx.Query("query")
	if query == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(&dto.ErrorResponse{Message: handler.MessageBadRequest})
	}

	data := h.s3Service.Search(ctx.RequestCtx(), userId, query)

	ctx.Status(fiber.StatusOK)

	return ctx.JSON(data)
}

func (h *Handler) Move(ctx fiber.Ctx) error {
	const op = "Move"

	handler.SetCommonHeaders(ctx)
	userId := handler.RequestedUserId(ctx)

	from := sanitizePath(ctx.Query("from"))
	to := sanitizePath(ctx.Query("to"))
	if from == "." || to == "." {
		return ctx.Status(fiber.StatusBadRequest).JSON(&dto.ErrorResponse{Message: handler.MessageBadRequest})
	}

	resp, err := h.s3Service.Move(ctx.RequestCtx(), userId, to, from)
	if err != nil {
		slog.Error(err.Error(), "pkg", pkg, "op", op)

		return ctx.Status(fiber.StatusBadRequest).JSON(&dto.ErrorResponse{Message: handler.MessageBadRequest})
	}

	ctx.Status(fiber.StatusOK)

	return ctx.JSON(resp)
}

func (h *Handler) DirectoryShow(ctx fiber.Ctx) error {
	const op = "DirectoryShow"

	handler.SetCommonHeaders(ctx)
	userId := handler.RequestedUserId(ctx)

	path := sanitizePath(ctx.Query("path"))

	data := h.s3Service.PaginateDirectory(ctx.RequestCtx(), userId, path)
	ctx.Status(fiber.StatusOK)

	return ctx.JSON(data)
}

func (h *Handler) DirectoryStore(ctx fiber.Ctx) error {
	const op = "DirectoryStore"

	handler.SetCommonHeaders(ctx)
	userId := handler.RequestedUserId(ctx)

	path := sanitizePath(ctx.Query("path"))

	resp, err := h.s3Service.StoreDirectory(ctx.RequestCtx(), userId, path)
	if err != nil {
		slog.Error(err.Error(), "pkg", pkg, "op", op)

		return ctx.Status(fiber.StatusBadRequest).JSON(&dto.ErrorResponse{Message: handler.MessageBadRequest})
	}

	ctx.Status(fiber.StatusCreated)

	return ctx.JSON(resp)
}

func sanitizePath(p string) string {
	if p == "" {
		return ""
	}

	cleaned := path.Clean(p)

	if strings.HasPrefix(cleaned, "..") || strings.HasPrefix(cleaned, "/") {
		cleaned = path.Clean("/" + p)
		cleaned = strings.TrimPrefix(cleaned, "/")
	}

	if strings.HasSuffix(p, "/") && !strings.HasSuffix(cleaned, "/") && cleaned != "." {
		cleaned += "/"
	}

	if cleaned == "." {
		return ""
	}

	return cleaned
}
