package trace

import (
	"net/http"

	"github.com/fpay/gopress"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func TracerMiddleware() gopress.MiddlewareFunc {
	return func(next gopress.HandlerFunc) gopress.HandlerFunc {
		return func(c gopress.Context) error {
			r := c.Request()
			opName := "HTTP " + r.Method + " " + r.URL.Path

			tracer := opentracing.GlobalTracer()

			carrier := opentracing.HTTPHeadersCarrier(r.Header)
			spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
			if err != nil {
				c.Logger().Warn("failed to extract spanContext from headers")
			}

			span := opentracing.StartSpan(opName, ext.RPCServerOption(spanCtx))
			defer span.Finish()

			ext.SpanKindRPCClient.Set(span)
			ext.HTTPMethod.Set(span, r.Method)
			ext.HTTPUrl.Set(span, r.URL.String())

			r = r.WithContext(opentracing.ContextWithSpan(r.Context(), span))
			c.SetRequest(r)

			if err := next(c); err != nil {
				span.SetTag("error", true)
				c.Error(err)
			}

			span.SetTag("error", false)
			ext.HTTPStatusCode.Set(span, uint16(c.Response().Status))

			return nil
		}
	}
}

// 将 span 注入到 http headers
func Inject(span *opentracing.Span, carrier interface{}) error {
	s := *span
	return s.Tracer().Inject(s.Context(), opentracing.HTTPHeaders, carrier)
}

// 从请求中提取 span
func Extract(r *http.Request) (opentracing.SpanContext, error) {
	tracer := opentracing.GlobalTracer()
	return tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
}
