package davinci

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/zanz1n/duvua/internal/errors"
	"github.com/zanz1n/duvua/pkg/pb/davinci"
	staticembed "github.com/zanz1n/duvua/static"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ davinci.DavinciServer = &GrpcServer{}

type GrpcServer struct {
	wg    Generator
	c     *http.Client
	token string
	davinci.UnimplementedDavinciServer
}

func NewGrpcServer(token string, wg Generator, client *http.Client) *GrpcServer {
	if client == nil {
		client = http.DefaultClient
	}

	return &GrpcServer{
		wg:    wg,
		c:     client,
		token: token,
	}
}

// SendWelcome implements davinci.DavinciServer.
func (s *GrpcServer) SendWelcome(
	ctx context.Context,
	req *davinci.WelcomeRequest,
) (*emptypb.Empty, error) {
	r, err := s.loadAvatar(req.ImageUrl)
	if err != nil {
		return nil, status.Error(codes.Aborted, err.Error())
	}
	defer r.Close()

	img, err := s.wg.Generate(r, req.Username, req.GreetingText)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			"failed to generate image: "+err.Error(),
		)
	}

	if err = s.sendMessage(req.Data, img); err != nil {
		return nil, status.Error(
			codes.Internal,
			"failed to send message: "+err.Error(),
		)
	}

	return &emptypb.Empty{}, nil
}

func (s *GrpcServer) loadAvatar(url string) (io.ReadCloser, error) {
	if url == "" {
		path := fmt.Sprintf("avatar/default-%v.png", rand.Intn(6))
		fileData, err := staticembed.Assets.Open(path)
		if err != nil {
			return nil, errors.Unexpected("failed to fetch avatar: " + err.Error())
		}

		return fileData, nil
	}

	req, err := s.c.Get(url)
	if err != nil {
		return nil, errors.Unexpected("failed to fetch avatar: " + err.Error())
	}

	return req.Body, nil
}

func (s *GrpcServer) sendMessage(data *davinci.ImageSendData, img *EncodedImage) error {
	msg := &discordgo.MessageSend{
		Content: data.Message,
	}

	ctype, body, err := discordgo.MultipartBodyWithJSON(msg, []*discordgo.File{{
		Name:        data.FileName + "." + img.Extension,
		ContentType: img.ContentType,
		Reader:      bytes.NewReader(img.Buf),
	}})
	if err != nil {
		return errors.Unexpected("encode multipart data: " + err.Error())
	}

	channelId := strconv.FormatUint(data.ChannelId, 10)
	url := discordgo.EndpointChannelMessages(channelId)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return errors.Unexpected("make http request: " + err.Error())
	}

	req.Header.Set("authorization", s.token)
	req.Header.Set("content-type", ctype)

	res, err := s.c.Do(req)
	if err != nil {
		return errors.Unexpected("send http request: " + err.Error())
	}

	if res.StatusCode > http.StatusIMUsed || res.StatusCode < http.StatusOK {
		return errors.Unexpectedf(
			"send http request: status code %s", res.Status,
		)
	}

	return nil
}
