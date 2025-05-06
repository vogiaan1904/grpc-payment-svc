package interceptors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vogiaan1904/payment-svc/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func redact(data interface{}, fields []string) string {
	if data == nil {
		return "null"
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return "Error marshaling data"
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(dataBytes, &dataMap); err != nil {
		return string(dataBytes)
	}

	for _, field := range fields {
		if _, exists := dataMap[field]; exists {
			dataMap[field] = "[Redacted]"
		}
	}

	redactedBytes, err := json.MarshalIndent(dataMap, "", "  ")
	if err != nil {
		return "Error marshaling redacted data"
	}

	return string(redactedBytes)
}

func GrpcClientLoggingInterceptor(logger log.Logger, redactedFields []string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		fmt.Printf("\n➡️  gRPC client request - Method: %s\n", method)
		fmt.Printf("📤 Request: %+v\n", req)

		startTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(startTime)

		if err != nil {
			st, _ := status.FromError(err)
			fmt.Printf("❌ gRPC client error - Method: %s, Code: %s, Message: %s, Duration: %v\n",
				method, st.Code(), st.Message(), duration)
		} else {
			fmt.Printf("✅ gRPC client response - Method: %s, Duration: %v\n", method, duration)
			fmt.Printf("📥 Response: %+v\n", reply)
		}

		return err
	}
}
