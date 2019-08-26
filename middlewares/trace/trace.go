package trace

import (
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

			// 将span注入到 http headers
			if span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier, ) != nil {
				panic("SpanContext Inject Error!")
			}

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
