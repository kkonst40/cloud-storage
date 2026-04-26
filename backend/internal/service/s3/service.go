package s3

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	p "path"
	"slices"
	"strings"

	"github.com/kkonst40/cloud-storage/backend/internal/api/dto/resource"
	"github.com/kkonst40/cloud-storage/backend/internal/config"
	errs "github.com/kkonst40/cloud-storage/backend/internal/errors"
	"github.com/minio/minio-go/v7"
)

const pkg = "S3Service"

type Service struct {
	bucket   string
	s3Client *minio.Client
}

func NewService(cfg *config.Config, s3Client *minio.Client) *Service {
	return &Service{
		bucket:   cfg.S3Bucket,
		s3Client: s3Client,
	}
}

func (s *Service) ObjectInfo(ctx context.Context, userId int64, path string) (resource.Response, error) {
	const op = "Object"

	object, err := s.s3Client.GetObject(
		ctx,
		s.bucket,
		p.Clean(pathWithUserPrefix(path, userId)),
		minio.GetObjectOptions{},
	)
	if err != nil {
		return resource.Response{}, errs.Wrap(pkg, op, err)
	}

	objInfo, err := object.Stat()
	if err != nil {
		return resource.Response{}, errs.Wrap(pkg, op, err)
	}

	return resource.Response{
		Path: objectDirWithSlash(objInfo.Key),
		Name: objectName(objInfo.Key),
		Size: objInfo.Size,
		Type: objectType(objInfo.Key),
	}, nil
}

func (s *Service) Object(ctx context.Context, userId int64, path string) (io.Reader, string, error) {
	const op = "Object"

	object, err := s.s3Client.GetObject(
		ctx,
		s.bucket,
		p.Clean(pathWithUserPrefix(path, userId)),
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, "", errs.Wrap(pkg, op, err)
	}

	objInfo, err := object.Stat()
	if err != nil {
		return nil, "", errs.Wrap(pkg, op, err)
	}

	return object, objectName(objInfo.Key), nil
}

func (s *Service) StoreObject(
	ctx context.Context,
	userId int64,
	files []resource.FilePayload,
	path string,
) ([]resource.Response, error) {
	opts := minio.PutObjectOptions{}
	data := make([]resource.Response, 0, len(files))

	for _, file := range files {
		resp, err := s.uploadFile(ctx, userId, file, path, opts)
		if err != nil {
			return nil, err
		}

		data = append(data, resp)
	}

	return data, nil
}

func (s *Service) Delete(ctx context.Context, userId int64, path string) error {
	const op = "Delete"

	// if the resource is a directory - remove all data inside
	if strings.HasSuffix(path, "/") {
		s.deleteRecursive(ctx, pathWithUserPrefix(path, userId))

		return nil
	}

	err := s.deleteObject(ctx, pathWithUserPrefix(path, userId), s.removeOptions())
	if err != nil {
		return errs.Wrap(pkg, op, err)
	}

	return nil
}

func (s *Service) Search(ctx context.Context, userId int64, query string) []resource.Response {
	const op = "Search"

	path := pathWithUserPrefix("", userId)
	opts := minio.ListObjectsOptions{
		Prefix:     path,
		StartAfter: path,
		Recursive:  true,
	}

	data := []resource.Response{}

	for v := range s.s3Client.ListObjects(ctx, s.bucket, opts) {
		if v.Err != nil {
			slog.Error(v.Err.Error(), "pkg", pkg, "op", op)

			continue
		}

		if strings.Contains(v.Key, query) {
			data = append(data, resource.Response{
				Path: objectRelDirWithSlash(v.Key, userId),
				Name: objectName(v.Key),
				Size: v.Size,
				Type: objectType(v.Key),
			})
		}
	}

	return data
}

func (s *Service) MakeZip(ctx context.Context, userId int64, path string) (*bytes.Buffer, error) {
	const op = "MakeZip"

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	opts := minio.ListObjectsOptions{
		Prefix:    pathWithUserPrefix(path, userId),
		Recursive: true,
	}

	for v := range s.s3Client.ListObjects(ctx, s.bucket, opts) {
		err := s.putObjectInZip(ctx, v, zipWriter, path)
		if err != nil {
			return nil, errs.Wrap(pkg, op, err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, errs.Wrap(pkg, op, err)
	}

	return buf, nil
}

func (s *Service) Move(ctx context.Context, userId int64, to, from string) (resource.Response, error) {
	const op = "Move"

	// if "from" is a directory, move all items inside "to"
	if strings.HasSuffix(from, "/") {
		fromPath := pathWithUserPrefix(from, userId)
		toPath := pathWithUserPrefix(to, userId)

		err := s.copyRecursive(ctx, toPath, fromPath)
		if err != nil {
			return resource.Response{}, errs.Wrap(pkg, op, err)
		}

		s.deleteRecursive(ctx, fromPath)

		return resource.Response{
			Path: objectDirWithSlash(to),
			Name: objectName(to),
			Size: 0,
			Type: "DIRECTORY",
		}, nil
	}

	objInfo, err := s.s3Client.CopyObject(
		ctx,
		minio.CopyDestOptions{
			Bucket: s.bucket,
			Object: pathWithUserPrefix(to, userId),
		},
		minio.CopySrcOptions{
			Bucket: s.bucket,
			Object: pathWithUserPrefix(from, userId),
		},
	)
	if err != nil {
		return resource.Response{}, errs.Wrap(pkg, op, err)
	}

	err = s.deleteObject(ctx, pathWithUserPrefix(from, userId), s.removeOptions())
	if err != nil {
		return resource.Response{}, errs.Wrap(pkg, op, err)
	}

	return resource.Response{
		Path: objectRelDirWithSlash(objInfo.Key, userId),
		Name: objectName(objInfo.Key),
		Size: objInfo.Size,
		Type: "FILE",
	}, nil
}

func (s *Service) StoreDirectory(ctx context.Context, userId int64, path string) (resource.Response, error) {
	const op = "StoreDirectory"

	object, err := s.s3Client.PutObject(
		ctx,
		s.bucket,
		pathWithUserPrefix(path, userId),
		nil,
		0,
		minio.PutObjectOptions{},
	)
	if err != nil {
		return resource.Response{}, errs.Wrap(pkg, op, err)
	}

	return resource.Response{
		Path: objectRelDirWithSlash(object.Key, userId),
		Name: objectName(object.Key),
		Size: object.Size,
		Type: objectType(object.Key),
	}, nil
}

func (s *Service) PaginateDirectory(ctx context.Context, userId int64, path string) []resource.Response {
	const op = "PaginateDirectory"

	objects := s.s3Client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:     pathWithUserPrefix(path, userId),
		StartAfter: pathWithUserPrefix(path, userId),
	})

	data := []resource.Response{}

	for v := range objects {
		if v.Err != nil {
			slog.Error(v.Err.Error(), "pkg", pkg, "op", op)

			continue
		}

		data = append(data, resource.Response{
			Path: objectRelDirWithSlash(v.Key, userId),
			Name: objectName(v.Key),
			Size: v.Size,
			Type: objectType(v.Key),
		})
	}

	s.sortObjectsList(data)

	return data
}

func (s *Service) uploadFile(
	ctx context.Context,
	userId int64,
	file resource.FilePayload,
	path string,
	opts minio.PutObjectOptions,
) (resource.Response, error) {
	const op = "uploadFile"

	objectKey := p.Join(pathWithUserPrefix(path, userId), file.Name)

	slog.Debug("", "OBJECT_KEY", objectKey)

	object, err := s.s3Client.PutObject(
		ctx,
		s.bucket,
		objectKey,
		file.Content,
		file.Size,
		opts,
	)
	if err != nil {
		return resource.Response{}, errs.Wrap(pkg, op, err)
	}

	return resource.Response{
		Path: objectRelDirWithSlash(object.Key, userId),
		Name: objectName(object.Key),
		Size: object.Size,
		Type: objectType(object.Key),
	}, nil
}

func (s *Service) deleteRecursive(ctx context.Context, path string) {
	const op = "deleteRecursive"

	opts := minio.ListObjectsOptions{
		Prefix:    path,
		Recursive: true,
	}

	removeOpts := s.removeOptions()

	for object := range s.s3Client.ListObjects(ctx, s.bucket, opts) {
		if object.Err != nil {
			slog.Error(object.Err.Error(), "pkg", pkg, "op", op)

			continue
		}

		err := s.deleteObject(ctx, object.Key, removeOpts)
		if err != nil {
			slog.Error(err.Error(), "pkg", pkg, "op", op)

			continue
		}
	}
}

func (s *Service) putObjectInZip(ctx context.Context, v minio.ObjectInfo, zipWriter *zip.Writer, prefix string) error {
	const op = "putObjectInZip"

	if v.Err != nil {
		return errs.Wrap(pkg, op, v.Err)
	}

	if strings.HasSuffix(v.Key, "/") {
		return nil
	}

	name, _ := strings.CutPrefix(v.Key, prefix)
	entry, err := zipWriter.Create(name)
	if err != nil {
		return errs.Wrap(pkg, op, err)
	}

	obj, err := s.s3Client.GetObject(ctx, s.bucket, v.Key, minio.GetObjectOptions{})
	if err != nil {
		return errs.Wrap(pkg, op, err)
	}

	if _, err := io.Copy(entry, obj); err != nil {
		return errs.Wrap(pkg, op, err)
	}

	if err := obj.Close(); err != nil {
		return errs.Wrap(pkg, op, err)
	}

	return nil
}

func (s *Service) sortObjectsList(data []resource.Response) {
	slices.SortFunc(data, func(a, b resource.Response) int {
		if a.Type == b.Type {
			return 0
		}

		if a.Type == "DIRECTORY" {
			return -1
		}

		return 1
	})
}

func (s *Service) copyRecursive(ctx context.Context, to, from string) error {
	const op = "copyRecursive"

	opts := minio.ListObjectsOptions{
		Prefix:     from,
		StartAfter: from,
		Recursive:  true,
	}

	isChanged := false

	for v := range s.s3Client.ListObjects(ctx, s.bucket, opts) {
		if v.Err != nil {
			return errs.Wrap(pkg, op, v.Err)
		}

		isChanged = true
		copyTo := p.Join(to, p.Base(v.Key))

		if strings.HasSuffix(v.Key, "/") {
			copyTo = fmt.Sprintf("%s/", copyTo)
		}

		_, err := s.s3Client.CopyObject(
			ctx,
			minio.CopyDestOptions{
				Bucket: s.bucket,
				Object: copyTo,
			},
			minio.CopySrcOptions{
				Bucket: s.bucket,
				Object: v.Key,
			},
		)
		if err != nil {
			return errs.Wrap(pkg, op, err)
		}
	}

	if !isChanged {
		_, err := s.s3Client.CopyObject(
			ctx,
			minio.CopyDestOptions{
				Bucket: s.bucket,
				Object: to,
			},
			minio.CopySrcOptions{
				Bucket: s.bucket,
				Object: from,
			},
		)
		if err != nil {
			return errs.Wrap(pkg, op, err)
		}
	}

	return nil
}

func (s *Service) deleteObject(ctx context.Context, path string, removeOpts *minio.RemoveObjectOptions) error {
	const op = "deleteObject"

	err := s.s3Client.RemoveObject(ctx, s.bucket, path, *removeOpts)
	if err != nil {
		return errs.Wrap(pkg, op, err)
	}

	return nil
}

func (s *Service) removeOptions() *minio.RemoveObjectOptions {
	return &minio.RemoveObjectOptions{
		ForceDelete:      true,
		GovernanceBypass: true,
	}
}

func pathWithUserPrefix(path string, userId int64) string {
	return fmt.Sprintf("user-%d-files/%s", userId, path)
}

func objectRelDirWithSlash(path string, userId int64) string {
	dir := objectDirWithSlash(path)
	noPrefix, _ := strings.CutPrefix(dir, fmt.Sprintf("user-%d-files/", userId))

	return noPrefix
}

func objectDirWithSlash(path string) string {
	return fmt.Sprintf("%s/", p.Dir(p.Clean(path)))
}

func objectType(path string) string {
	if strings.HasSuffix(path, "/") {
		return "DIRECTORY"
	}

	return "FILE"
}

func objectName(path string) string {
	if strings.HasSuffix(path, "/") {
		return fmt.Sprintf("%s/", p.Base(path))
	}

	return p.Base(path)
}
