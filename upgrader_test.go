package websocket

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"github.com/valyala/fasthttp"
	"net"
	"testing"
)

func TestFasthttpWebsocket(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FasthttpWebsocket Suite")
}

func buildValidCtx() *fasthttp.RequestCtx {
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.Header.Add("Upgrade", "websocket")
	ctx.Request.Header.Add("Connection", "Upgrade")
	ctx.Request.Header.Add("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	ctx.Request.Header.Add("Sec-WebSocket-Protocol", "chat, superchat")
	ctx.Request.Header.Add("Sec-WebSocket-Version", "13")
	return ctx
}

func noopHandler(conn net.Conn) {}

var _ = Describe("Upgrader", func() {
	It("should check generating accepts for keys", func() {
		Expect(generateAcceptFromKey([]byte("bu+Y8NrjcUTuR6rM3u8KQg=="))).To(Equal([]byte("53BdvpuochdSseQfEcKbg7HYqVo=")))
		Expect(generateAcceptFromKey([]byte("OUjxNBsFsoezcd+uPmRgrg=="))).To(Equal([]byte("BoQ1VcNcPY+IX+IqFi6m7HFgB3A=")))
	})

	It("should upgrade the connection", func() {
		ctx := buildValidCtx()
		upgrader := &Upgrader{}
		Expect(upgrader.Upgrade(ctx, noopHandler)).To(BeNil())
	})

	It("should fail upgrading due to wrong method", func() {
		ctx := buildValidCtx()
		ctx.Request.Header.SetMethod("POST")

		upgrader := &Upgrader{}
		err := upgrader.Upgrade(ctx, nil)
		Expect(fmt.Sprintf("%s", err)).To(Equal("Method not allowed"))
	})

	It("should fail upgrading due to wrong connection type", func() {
		ctx := buildValidCtx()
		ctx.Request.Header.Set("Connection", "another type")

		upgrader := &Upgrader{}
		err := upgrader.Upgrade(ctx, nil)
		Expect(fmt.Sprintf("%s", err)).To(Equal("Invalid connection type"))
	})

	It("should fail upgrading due to no provided connection type", func() {
		ctx := buildValidCtx()
		ctx.Request.Header.Del("Connection")

		upgrader := &Upgrader{}
		err := upgrader.Upgrade(ctx, nil)
		Expect(fmt.Sprintf("%s", err)).To(Equal("Invalid connection type"))
	})

	It("should fail upgrading due to wrong upgrade value", func() {
		ctx := buildValidCtx()
		ctx.Request.Header.Set("Upgrade", "invalid")

		upgrader := &Upgrader{}
		err := upgrader.Upgrade(ctx, nil)
		Expect(fmt.Sprintf("%s", err)).To(Equal("This connection cannot be upgraded to 'invalid'"))
	})

	It("should fail upgrading due to missing key", func() {
		ctx := buildValidCtx()
		ctx.Request.Header.Del("Sec-WebSocket-Key")

		upgrader := &Upgrader{}
		err := upgrader.Upgrade(ctx, nil)
		Expect(fmt.Sprintf("%s", err)).To(Equal("The key is missing."))
	})

	It("should fail upgrading due to missing version", func() {
		ctx := buildValidCtx()
		ctx.Request.Header.Del("Sec-WebSocket-Version")

		upgrader := &Upgrader{}
		err := upgrader.Upgrade(ctx, nil)
		Expect(fmt.Sprintf("%s", err)).To(Equal("No version provided."))
	})

	It("should fail upgrading due to wrong version", func() {
		ctx := buildValidCtx()
		ctx.Request.Header.Set("Sec-WebSocket-Version", "12")

		upgrader := &Upgrader{}
		err := upgrader.Upgrade(ctx, nil)
		Expect(fmt.Sprintf("%s", err)).To(Equal("The version is not supported."))
	})

	Describe("headerVisit", func() {
		It("should not visit any value on an empty string", func() {
			list := make([]string, 0)
			headerVisit([]byte(""), func(name, value []byte) bool {
				Expect(string(name)).To(BeEmpty())
				list = append(list, string(value))
				return true
			})
			Expect(list).To(BeEmpty())
		})

		It("should visit a single value", func() {
			headerVisit([]byte("foo"), func(name, value []byte) bool {
				Expect(string(name)).To(BeEmpty())
				Expect(string(value)).To(Equal("foo"))
				return true
			})
		})

		It("should visit a single named value", func() {
			headerVisit([]byte("foo=bar"), func(name, value []byte) bool {
				Expect(string(name)).To(Equal("foo"))
				Expect(string(value)).To(Equal("bar"))
				return true
			})
		})

		It("should visit a two values sticked", func() {
			list := make([]string, 0)
			headerVisit([]byte("foo,bar"), func(name, value []byte) bool {
				Expect(string(name)).To(BeEmpty())
				list = append(list, string(value))
				return true
			})
			Expect(list).To(HaveLen(2))
			Expect(list[0]).To(Equal("foo"))
			Expect(list[1]).To(Equal("bar"))
		})

		type keyValue struct {
			key   string
			value string
		}

		It("should visit a two named values sticked", func() {
			list := make([]keyValue, 0)
			headerVisit([]byte("foo=bar,john=doe"), func(name, value []byte) bool {
				list = append(list, keyValue{string(name), string(value)})
				return true
			})
			Expect(list).To(HaveLen(2))
			Expect(list[0].key).To(Equal("foo"))
			Expect(list[0].value).To(Equal("bar"))
			Expect(list[1].key).To(Equal("john"))
			Expect(list[1].value).To(Equal("doe"))
		})

		It("should visit an empty named", func() {
			list := make([]keyValue, 0)
			headerVisit([]byte("foo=,john=doe"), func(name, value []byte) bool {
				list = append(list, keyValue{string(name), string(value)})
				return true
			})
			Expect(list).To(HaveLen(2))
			Expect(list[0].key).To(Equal("foo"))
			Expect(list[0].value).To(Equal(""))
			Expect(list[1].key).To(Equal("john"))
			Expect(list[1].value).To(Equal("doe"))
		})

		It("should visit an empty named at the end", func() {
			list := make([]keyValue, 0)
			headerVisit([]byte("foo=bar,john="), func(name, value []byte) bool {
				list = append(list, keyValue{string(name), string(value)})
				return true
			})
			Expect(list).To(HaveLen(2))
			Expect(list[0].key).To(Equal("foo"))
			Expect(list[0].value).To(Equal("bar"))
			Expect(list[1].key).To(Equal("john"))
			Expect(list[1].value).To(Equal(""))
		})

		It("should visit an empty value between valid values", func() {
			list := make([]string, 0)
			headerVisit([]byte("foo,,bar, , final"), func(name, value []byte) bool {
				Expect(string(name)).To(BeEmpty())
				list = append(list, string(value))
				return true
			})
			Expect(list).To(HaveLen(5))
			Expect(list[0]).To(Equal("foo"))
			Expect(list[1]).To(Equal(""))
			Expect(list[2]).To(Equal("bar"))
			Expect(list[3]).To(Equal(""))
			Expect(list[4]).To(Equal("final"))
		})

		It("should visit an empty value between valid named values", func() {
			list := make([]keyValue, 0)
			headerVisit([]byte("foo=bar,,john=doe, , fin=al"), func(name, value []byte) bool {
				list = append(list, keyValue{string(name), string(value)})
				return true
			})
			Expect(list).To(HaveLen(5))
			Expect(list[0].key).To(Equal("foo"))
			Expect(list[0].value).To(Equal("bar"))
			Expect(list[1].key).To(Equal(""))
			Expect(list[1].value).To(Equal(""))
			Expect(list[2].key).To(Equal("john"))
			Expect(list[2].value).To(Equal("doe"))
			Expect(list[3].key).To(Equal(""))
			Expect(list[3].value).To(Equal(""))
			Expect(list[4].key).To(Equal("fin"))
			Expect(list[4].value).To(Equal("al"))
		})

		It("should visit multiple values", func() {
			list := make([]string, 0)
			headerVisit([]byte("foo, bar, john, doe"), func(name, value []byte) bool {
				Expect(string(name)).To(BeEmpty())
				list = append(list, string(value))
				return true
			})
			Expect(list).To(HaveLen(4))
			Expect(list[0]).To(Equal("foo"))
			Expect(list[1]).To(Equal("bar"))
			Expect(list[2]).To(Equal("john"))
			Expect(list[3]).To(Equal("doe"))
		})

		It("should visit mix of values and named and empty values", func() {
			list := make([]keyValue, 0)
			headerVisit([]byte("foo=bar, bar, ,, john=,     , doe, john=doe"), func(name, value []byte) bool {
				list = append(list, keyValue{string(name), string(value)})
				return true
			})
			Expect(list).To(HaveLen(8))
			Expect(list[0].key).To(Equal("foo"))
			Expect(list[0].value).To(Equal("bar"))
			Expect(list[1].key).To(Equal(""))
			Expect(list[1].value).To(Equal("bar"))
			Expect(list[2].key).To(Equal(""))
			Expect(list[2].value).To(Equal(""))
			Expect(list[3].key).To(Equal(""))
			Expect(list[3].value).To(Equal(""))
			Expect(list[4].key).To(Equal("john"))
			Expect(list[4].value).To(Equal(""))
			Expect(list[5].key).To(Equal(""))
			Expect(list[5].value).To(Equal(""))
			Expect(list[6].key).To(Equal(""))
			Expect(list[6].value).To(Equal("doe"))
			Expect(list[7].key).To(Equal("john"))
			Expect(list[7].value).To(Equal("doe"))
		})
	})
})
