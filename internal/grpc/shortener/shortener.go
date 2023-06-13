package shortener

import (
	"context"
	"errors"
	"fmt"
	"github.com/LorezV/url-shorter.git/internal/config"
	"github.com/LorezV/url-shorter.git/internal/grpc/auth"
	"github.com/LorezV/url-shorter.git/internal/repository"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/LorezV/url-shorter.git/proto"
)

var errWrongURL = errors.New("wrong url")

type server struct {
	pb.UnimplementedShortenerServer

	https  bool
	domain string
}

// NewGRPCServer - конструктор сервера шортенера.
func NewGRPCServer(https bool, domain string) *server {
	return &server{
		https:  https,
		domain: domain,
	}
}

// Ping - обработчик для проверки связи с хранилищем.
func (s server) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if len(config.AppConfig.DatabaseDsn) == 0 {
		return &emptypb.Empty{}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := config.DB.PingContext(ctx); err != nil {
		return nil, status.Error(codes.Internal, "database ping failed")
	}

	return &emptypb.Empty{}, nil
}

// Short - обработчик для создания короткой ссылки.
func (s server) Short(ctx context.Context, req *pb.ShortRequest) (*pb.ShortResponse, error) {
	user, err := auth.GetUser(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get user: %v", err)
	}

	id, url, err := s.short(ctx, user, req.Url)
	if errors.Is(err, repository.ErrorURLDuplicate) {
		return nil, status.Errorf(codes.AlreadyExists, "short error: %v", err)
	}
	if errors.Is(err, errWrongURL) {
		return nil, status.Errorf(codes.InvalidArgument, "short error: %v", err)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ShortResponse{
		Id:       id,
		Url:      req.Url,
		ShortUrl: url,
	}, nil
}

// Get - обработчик, который получает полную ссылку из id короткой.
func (s server) Get(ctx context.Context, l *pb.GetRequest) (*pb.GetResponse, error) {
	if len(l.Id) == 0 {
		return nil, status.Error(codes.InvalidArgument, "id length should be greater than 0")
	}
	url, ok := repository.GlobalRepository.Get(ctx, l.Id)
	if !ok {
		return nil, status.Error(codes.NotFound, "url not found")
	}

	if url.IsDeleted {
		return nil, status.Error(codes.Unavailable, "link is deleted")
	}

	return &pb.GetResponse{
		Id:       url.ID,
		Url:      url.Original,
		ShortUrl: url.Short,
	}, nil
}

// GetLinks - обработчик возвращающий все ссылки принадлежащие текущему пользователю.
func (s server) GetLinks(ctx context.Context, _ *emptypb.Empty) (*pb.GetLinksResponse, error) {
	user, err := auth.GetUser(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get user: %v", err)
	}

	links, err := repository.GlobalRepository.GetAllByUser(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "server error: %v", err)
	}

	b := &pb.GetLinksResponse{}
	for _, link := range links {
		b.Links = append(b.Links, &pb.GetLinksResponse_Link{
			Id:       link.ID,
			Url:      link.Original,
			ShortUrl: s.genShortLink(link.ID),
		})
	}

	return b, nil
}

// BatchShort - обработчик для создания пачки коротких ссылок.
func (s server) BatchShort(ctx context.Context, in *pb.BatchShortRequest) (*pb.BatchShortResponse, error) {
	user, err := auth.GetUser(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get user: %v", err)
	}

	res := &pb.BatchShortResponse{}

	for _, link := range in.Links {
		id, shortURL, err := s.short(ctx, user, link.Url)
		if err != nil {
			continue
		}
		res.Links = append(res.Links, &pb.BatchShortResponse_Link{
			Id:            id,
			Url:           link.Url,
			ShortUrl:      shortURL,
			CorrelationId: link.CorrelationId,
		})
	}

	return res, nil
}

// Delete - обработчик для удаления ссылок пользователя.
func (s server) Delete(ctx context.Context, b *pb.DeleteRequest) (*emptypb.Empty, error) {
	user, err := auth.GetUser(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get user: %v", err)
	}

	ids := make([]string, 0)
	for _, link := range b.Ids {
		if len(link) == 0 {
			continue
		}
		ids = append(ids, link)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		ok := repository.GlobalRepository.DeleteManyByUser(ctx, ids, user)
		if !ok {
			log.Printf("unable to delete user ids: %v", err)
		}
	}()

	return &emptypb.Empty{}, nil
}

// GetStats - обработчик, который возвращает статистику сервера.
func (s server) GetStats(ctx context.Context, _ *emptypb.Empty) (*pb.GetStatsResponse, error) {
	stats, err := repository.GlobalRepository.GetStats(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "server error: %v", err)
	}

	return &pb.GetStatsResponse{
		Links: uint64(stats.Urls),
		Users: uint64(stats.Users),
	}, nil
}

func (s server) short(ctx context.Context, user string, url string) (id string, shortURL string, err error) {
	if len(url) == 0 {
		return "", "", errWrongURL
	}

	iURL, err := repository.MakeURL(url, user)
	if errors.Is(err, repository.ErrorURLDuplicate) {
		return "", "", err
	}

	if err != nil {
		return "", "", err
	}

	return iURL.Original, iURL.Short, nil
}

func (s server) genShortLink(id string) string {
	if s.https {
		return fmt.Sprintf("https://%s/%s", s.domain, id)
	}
	return fmt.Sprintf("%s/%s", s.domain, id)
}
