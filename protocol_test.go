package websocket

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestProtocol(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Protocol Test Suite")
}

var (
	singleFrameUnmaskedText        = []byte{0x81, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f}
	singleFrameUnmaskedTextPayload = []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}

	singleFrameMaskedText        = []byte{0x81, 0x85, 0x37, 0xfa, 0x21, 0x3d, 0x7f, 0x9f, 0x4d, 0x51, 0x58}
	singleFrameMaskedTextMask    = []byte{0x37, 0xfa, 0x21, 0x3d}
	singleFrameMaskedTextPayload = []byte{0x7f, 0x9f, 0x4d, 0x51, 0x58}

	fragmentedUnmaskedText1        = []byte{0x01, 0x03, 0x48, 0x65, 0x6c}
	fragmentedUnmaskedText1Payload = []byte{0x48, 0x65, 0x6c}
	fragmentedUnmaskedText2        = []byte{0x80, 0x02, 0x6c, 0x6f}
	fragmentedUnmaskedText2Payload = []byte{0x6c, 0x6f}

	singleFrameUnmaskedPingRequest        = []byte{0x89, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f}
	singleFrameUnmaskedPingRequestPayload = []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}

	singleFrameMaskedPongResponse        = []byte{0x8a, 0x85, 0x37, 0xfa, 0x21, 0x3d, 0x7f, 0x9f, 0x4d, 0x51, 0x58}
	singleFrameMaskedPongResponseMask    = []byte{0x37, 0xfa, 0x21, 0x3d}
	singleFrameMaskedPongResponsePayload = []byte{0x7f, 0x9f, 0x4d, 0x51, 0x58}

	singleFrameBinaryUnmasked256BytesLongHeader = []byte{0x82, 0x7E, 0x01, 0x00}
	singleFrameBinaryUnmasked64KBytesLongHeader = []byte{0x82, 0x7F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00}

	singleFrameMaskedFlatedText        = []byte{0xc1, 0x86, 0xda, 0x7f, 0x9b, 0x17, 0xf0, 0x36, 0xb6, 0x39, 0xdb, 0x7f}
	singleFrameMaskedFlatedTextMask    = []byte{0xda, 0x7f, 0x9b, 0x17}
	singleFrameMaskedFlatedTextPayload = []byte{0xf0, 0x36, 0xb6, 0x39, 0xdb, 0x7f}
)

var _ = Describe("Protocol", func() {
	Describe("DecodePacket", func() {
		It("should parse a single-frame unmasked text message", func() {
			fin, rsv1, rsv2, rsv3, opcode, payloadLen, maskingKey, payload, err := DecodePacket(singleFrameUnmaskedText)
			Expect(err).To(BeNil())
			Expect(fin).To(BeTrue())
			Expect(rsv1).To(BeFalse())
			Expect(rsv2).To(BeFalse())
			Expect(rsv3).To(BeFalse())
			Expect(opcode).To(Equal(byte(OPCodeTextFrame)))
			Expect(payloadLen).To(Equal(uint64(len(singleFrameUnmaskedTextPayload))))
			Expect(maskingKey).To(BeNil())
			Expect(payload).To(HaveLen(int(payloadLen)))
			Expect(payload).To(Equal(singleFrameUnmaskedTextPayload))
		})

		It("should parse a single-frame masked text message", func() {
			fin, rsv1, rsv2, rsv3, opcode, payloadLen, maskingKey, payload, err := DecodePacket(singleFrameMaskedText)
			Expect(err).To(BeNil())
			Expect(fin).To(BeTrue())
			Expect(rsv1).To(BeFalse())
			Expect(rsv2).To(BeFalse())
			Expect(rsv3).To(BeFalse())
			Expect(opcode).To(Equal(byte(OPCodeTextFrame)))
			Expect(payloadLen).To(Equal(uint64(len(singleFrameMaskedTextPayload))))
			Expect(maskingKey).To(Equal(singleFrameMaskedTextMask))
			Expect(payload).To(HaveLen(int(payloadLen)))
			Expect(payload).To(Equal(singleFrameMaskedTextPayload))
		})

		It("should parse a fragmented unmasked text message, first part", func() {
			fin, rsv1, rsv2, rsv3, opcode, payloadLen, maskingKey, payload, err := DecodePacket(fragmentedUnmaskedText1)
			Expect(err).To(BeNil())
			Expect(fin).To(BeFalse())
			Expect(rsv1).To(BeFalse())
			Expect(rsv2).To(BeFalse())
			Expect(rsv3).To(BeFalse())
			Expect(opcode).To(Equal(byte(OPCodeTextFrame)))
			Expect(payloadLen).To(Equal(uint64(len(fragmentedUnmaskedText1Payload))))
			Expect(maskingKey).To(BeNil())
			Expect(payload).To(HaveLen(int(payloadLen)))
			Expect(payload).To(Equal(fragmentedUnmaskedText1Payload))
		})

		It("should parse a fragmented unmasked text message, second part", func() {
			fin, rsv1, rsv2, rsv3, opcode, payloadLen, maskingKey, payload, err := DecodePacket(fragmentedUnmaskedText2)
			Expect(err).To(BeNil())
			Expect(fin).To(BeTrue())
			Expect(rsv1).To(BeFalse())
			Expect(rsv2).To(BeFalse())
			Expect(rsv3).To(BeFalse())
			Expect(opcode).To(Equal(byte(OPCodeContinuationFrame)))
			Expect(payloadLen).To(Equal(uint64(len(fragmentedUnmaskedText2Payload))))
			Expect(maskingKey).To(BeNil())
			Expect(payload).To(HaveLen(int(payloadLen)))
			Expect(payload).To(Equal(fragmentedUnmaskedText2Payload))
		})

		It("should parse a masked ping request", func() {
			fin, rsv1, rsv2, rsv3, opcode, payloadLen, maskingKey, payload, err := DecodePacket(singleFrameUnmaskedPingRequest)
			Expect(err).To(BeNil())
			Expect(fin).To(BeTrue())
			Expect(rsv1).To(BeFalse())
			Expect(rsv2).To(BeFalse())
			Expect(rsv3).To(BeFalse())
			Expect(opcode).To(Equal(byte(OPCodePingFrame)))
			Expect(payloadLen).To(Equal(uint64(len(singleFrameUnmaskedPingRequestPayload))))
			Expect(maskingKey).To(BeNil())
			Expect(payload).To(HaveLen(int(payloadLen)))
			Expect(payload).To(Equal(singleFrameUnmaskedPingRequestPayload))
		})

		It("should parse an unmasked pong response", func() {
			fin, rsv1, rsv2, rsv3, opcode, payloadLen, maskingKey, payload, err := DecodePacket(singleFrameMaskedPongResponse)
			Expect(err).To(BeNil())
			Expect(fin).To(BeTrue())
			Expect(rsv1).To(BeFalse())
			Expect(rsv2).To(BeFalse())
			Expect(rsv3).To(BeFalse())
			Expect(opcode).To(Equal(byte(OPCodePongFrame)))
			Expect(payloadLen).To(Equal(uint64(len(singleFrameMaskedPongResponsePayload))))
			Expect(maskingKey).To(Equal(singleFrameMaskedPongResponseMask))
			Expect(payload).To(HaveLen(int(payloadLen)))
			Expect(payload).To(Equal(singleFrameMaskedPongResponsePayload))
		})

		It("should parse a package with payload 256 bytes long", func() {
			fin, rsv1, rsv2, rsv3, opcode, payloadLen, maskingKey, payload, err := DecodePacket(append(singleFrameBinaryUnmasked256BytesLongHeader, make([]byte, 256)...))
			Expect(err).To(BeNil())
			Expect(fin).To(BeTrue())
			Expect(rsv1).To(BeFalse())
			Expect(rsv2).To(BeFalse())
			Expect(rsv3).To(BeFalse())
			Expect(opcode).To(Equal(byte(OPCodeBinaryFrame)))
			Expect(payloadLen).To(Equal(uint64(256)))
			Expect(maskingKey).To(BeNil())
			Expect(payload).To(HaveLen(int(256)))
		})

		It("should parse a package with payload 64 KBytes long", func() {
			fin, rsv1, rsv2, rsv3, opcode, payloadLen, maskingKey, payload, err := DecodePacket(append(singleFrameBinaryUnmasked64KBytesLongHeader, make([]byte, 1024*64)...))
			Expect(err).To(BeNil())
			Expect(fin).To(BeTrue())
			Expect(rsv1).To(BeFalse())
			Expect(rsv2).To(BeFalse())
			Expect(rsv3).To(BeFalse())
			Expect(opcode).To(Equal(byte(OPCodeBinaryFrame)))
			Expect(payloadLen).To(Equal(uint64(1024 * 64)))
			Expect(maskingKey).To(BeNil())
			Expect(payload).To(HaveLen(int(1024 * 64)))
		})

		It("should parse a single frame masked and flated", func() {
			fin, rsv1, rsv2, rsv3, opcode, payloadLen, maskingKey, payload, err := DecodePacket(singleFrameMaskedFlatedText)
			Expect(err).To(BeNil())
			Expect(fin).To(BeTrue())
			Expect(rsv1).To(BeFalse())
			Expect(rsv2).To(BeFalse())
			Expect(rsv3).To(BeFalse())
			Expect(opcode).To(Equal(byte(OPCodeTextFrame)))
			Expect(payloadLen).To(Equal(uint64(len(singleFrameMaskedFlatedTextPayload))))
			Expect(maskingKey).To(Equal(singleFrameMaskedFlatedTextMask))
			Expect(payload).To(HaveLen(int(payloadLen)))
			Expect(payload).To(Equal(singleFrameMaskedFlatedTextPayload))
		})

		It("should fail parsing a broken single-frame unmasked text message with no minimum requirement", func() {
			_, _, _, _, _, _, _, _, err := DecodePacket(singleFrameUnmaskedText[:2])
			Expect(err).NotTo(BeNil())
			Expect(IsUnexpectedEndOfPacket(err)).To(BeTrue())
		})

		It("should fail parsing a broken single-frame unmasked text message with no 16-bits length", func() {
			_, _, _, _, _, _, _, _, err := DecodePacket(singleFrameBinaryUnmasked256BytesLongHeader[:4])
			Expect(err).NotTo(BeNil())
			Expect(IsUnexpectedEndOfPacket(err)).To(BeTrue())

			_, _, _, _, _, _, _, _, err = DecodePacket(singleFrameBinaryUnmasked256BytesLongHeader[:3])
			Expect(err).NotTo(BeNil())
			Expect(IsUnexpectedEndOfPacket(err)).To(BeTrue())
		})

		It("should fail parsing a broken single-frame unmasked text message with no 64-bits length", func() {
			_, _, _, _, _, _, _, _, err := DecodePacket(singleFrameBinaryUnmasked64KBytesLongHeader[:9])
			Expect(err).NotTo(BeNil())
			Expect(IsUnexpectedEndOfPacket(err)).To(BeTrue())
		})

		It("should fail parsing single-frame masked text message broken at the mask", func() {
			_, _, _, _, _, _, _, _, err := DecodePacket(singleFrameMaskedFlatedText[:6])
			Expect(err).NotTo(BeNil())
			Expect(IsUnexpectedEndOfPacket(err)).To(BeTrue())
		})

		It("should fail parsing single-frame masked text message broken at the payload", func() {
			_, _, _, _, _, _, _, _, err := DecodePacket(singleFrameMaskedFlatedText[:11])
			Expect(err).NotTo(BeNil())
			Expect(IsUnexpectedEndOfPacket(err)).To(BeTrue())
		})
	})

	Describe("EncodePacket", func() {
		It("should parse a single-frame unmasked text message", func() {
			packet, err := EncodePacket(true, false, false, false, OPCodeTextFrame, 5, nil, []byte("Hello"))
			Expect(err).To(BeNil())
			Expect(packet).To(Equal(singleFrameUnmaskedText))
		})

		It("should parse a single-frame masked text message", func() {
			maskedPayload := []byte("Hello")
			Unmask(maskedPayload, singleFrameMaskedTextMask)

			packet, err := EncodePacket(true, false, false, false, OPCodeTextFrame, 5, singleFrameMaskedTextMask, maskedPayload)
			Expect(err).To(BeNil())
			Expect(packet).To(Equal(singleFrameMaskedText))
		})

		It("should parse a fragmented unmasked text message, first part", func() {
			packet, err := EncodePacket(false, false, false, false, OPCodeTextFrame, 3, nil, []byte("Hel"))
			Expect(err).To(BeNil())
			Expect(packet).To(Equal(fragmentedUnmaskedText1))
		})

		It("should parse a fragmented unmasked text message, second part", func() {
			packet, err := EncodePacket(true, false, false, false, OPCodeContinuationFrame, 2, nil, []byte("lo"))
			Expect(err).To(BeNil())
			Expect(packet).To(Equal(fragmentedUnmaskedText2))
		})

		It("should parse an unmasked ping request", func() {
			packet, err := EncodePacket(true, false, false, false, OPCodePingFrame, 5, nil, []byte("Hello"))
			Expect(err).To(BeNil())
			Expect(packet).To(Equal(singleFrameUnmaskedPingRequest))
		})

		It("should parse masked pong response", func() {
			maskedPayload := []byte("Hello")
			Unmask(maskedPayload, singleFrameMaskedPongResponseMask)

			packet, err := EncodePacket(true, false, false, false, OPCodePongFrame, 5, singleFrameMaskedPongResponseMask, maskedPayload)
			Expect(err).To(BeNil())
			Expect(packet).To(Equal(singleFrameMaskedPongResponse))
		})

		It("should parse a package with payload 256 bytes long", func() {
			payload := make([]byte, 256)
			packet, err := EncodePacket(true, false, false, false, OPCodeBinaryFrame, uint64(len(payload)), nil, payload)
			Expect(err).To(BeNil())
			Expect(packet).To(Equal(append(singleFrameBinaryUnmasked256BytesLongHeader, make([]byte, 256)...)))
		})

		It("should parse a package with payload 64 KBytes long", func() {
			payload := make([]byte, 1024*64)
			packet, err := EncodePacket(true, false, false, false, OPCodeBinaryFrame, uint64(len(payload)), nil, payload)
			Expect(err).To(BeNil())
			Expect(packet).To(Equal(append(singleFrameBinaryUnmasked64KBytesLongHeader, make([]byte, 1024*64)...)))
		})

		It("should parse a single frame masked and flated", func() {
			payload := []byte("test")
			payloadFlatted, n, err := Flate(make([]byte, 0, 1024), payload)
			Expect(err).To(BeNil())
			Expect(n).To(Equal(6))
			Expect(payloadFlatted).To(HaveLen(6))
			Unmask(payloadFlatted, singleFrameMaskedFlatedTextMask)

			packet, err := EncodePacket(true, true, false, false, OPCodeTextFrame, 6, singleFrameMaskedFlatedTextMask, payloadFlatted)
			Expect(err).To(BeNil())
			Expect(packet).To(Equal(singleFrameMaskedFlatedText))
		})
	})

	Describe("Unmask", func() {
		It("should unmask a single frame text", func() {
			buff := make([]byte, len(singleFrameMaskedTextPayload))
			copy(buff, singleFrameMaskedTextPayload)
			Unmask(buff, singleFrameMaskedTextMask)
			Expect(string(buff)).To(Equal("Hello"))
		})

		It("should unmask a single frame text", func() {
			buff := make([]byte, len(singleFrameMaskedTextPayload))
			copy(buff, singleFrameMaskedPongResponsePayload)
			Unmask(buff, singleFrameMaskedPongResponseMask)
			Expect(string(buff)).To(Equal("Hello"))
		})
	})

	Describe("Deflate", func() {
		It("should deflate a payload", func() {
			buff := make([]byte, len(singleFrameMaskedFlatedTextPayload))
			copy(buff, singleFrameMaskedFlatedTextPayload)
			Unmask(buff, singleFrameMaskedFlatedTextMask)
			dst, err := Deflate(make([]byte, 0, 1024), buff)
			Expect(err).To(BeNil())
			Expect(string(dst)).To(Equal("test"))
		})
	})
})

func BenchmarkDecodePacket_SingleFrameUnmaskedText(b *testing.B) {
	DecodePacket(singleFrameUnmaskedText)
}

func BenchmarkDecodePacket_SingleFrameMaskedText(b *testing.B) {
	DecodePacket(singleFrameMaskedText)
}

func BenchmarkUnmask(b *testing.B) {
	buff := make([]byte, len(singleFrameMaskedFlatedTextPayload))
	copy(buff, singleFrameMaskedFlatedTextPayload)
	Unmask(buff, singleFrameMaskedFlatedTextMask)
}

func BenchmarkDeflate(b *testing.B) {
	buff := make([]byte, len(singleFrameMaskedFlatedTextPayload))
	copy(buff, singleFrameMaskedFlatedTextPayload)
	Unmask(buff, singleFrameMaskedFlatedTextMask)
	Deflate(make([]byte, 0, 1024), buff)
}
