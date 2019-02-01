package controllers

// import (
// 	proto "github.com/golang/protobuf/proto"
// 	"github.com/labstack/echo/v4"
// 	"net/http"
// )

// func GRPC(c echo.Context) error {

// 	// sc, ok := c.ServiceClient.
// 	// if !ok {
// 	// 	return echo.NewHTTPError(http.StatusBadRequest, "コンテキストが取得できません")
// 	// }

// 	rep, err := c.ServiceClient.GetHello(netCtx.Background(), &proto.Empty{})
// 	if err != nil {
// 		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}

// 	return c.JSON(http.StatusOK, map[string]interface{}{
// 		"reply": rep.Result,
// 	})

// 	return &pb.HelloReply{Message: "Hello again " + in.Name}, nil

// }

// // func (s *server) SayHelloAgain(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
// // 	return &pb.HelloReply{Message: "Hello again " + in.Name}, nil
// // }
