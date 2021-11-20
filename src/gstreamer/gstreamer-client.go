package gstreamer

/*
#cgo pkg-config: gstreamer-plugins-bad-1.0 gstreamer-rtp-1.0 gstreamer-webrtc-1.0 gstreamer-plugins-base-1.0 glib-2.0 libsoup-2.4 json-glib-1.0
#cgo CFLAGS: -Wall
#cgo CFLAGS: -Wno-deprecated-declarations -Wimplicit-function-declaration -Wformat-security
#cgo LDFLAGS: -lgstsdp-1.0
#include <cfunc.h>
*/
import "C"
import (
	"fmt"
	"strconv"
	"unsafe"
)

func (g *GStreamer) InitGstClient(server *GStreamer) {
	C.gst_element_set_state(server.pipeline, C.GST_STATE_PAUSED)
	webrtcName := C.CString("webrtcbin")
	defer C.free(unsafe.Pointer(webrtcName))
	webrtcNameDesc := C.CString("webrtcbin1")
	defer C.free(unsafe.Pointer(webrtcNameDesc))
	g.Webrtc = C.gst_element_factory_make(webrtcName, webrtcNameDesc)
	g_object_set(C.gpointer(g.Webrtc), "stun-server", unsafe.Pointer(C.CString("stun://stun.l.google.com:19302")))

	capsStr := C.CString("application/x-rtp,media=video,encoding-name=H264,clock-rate=90000")
	defer C.free(unsafe.Pointer(capsStr))
	var caps *C.GstCaps = C.gst_caps_from_string(capsStr)

	g_signal_emit_by_name_trans(g.Webrtc, "add-transceiver", C.GST_WEBRTC_RTP_TRANSCEIVER_DIRECTION_SENDONLY, unsafe.Pointer(caps))

	C.gst_bin_add(GST_BIN(server.pipeline), g.Webrtc)

	var reason C.GstPadLinkReturn
	srcStr := C.CString("src_%u")
	sinkStr := C.CString("sink_%u")
	defer func() {
		C.free(unsafe.Pointer(srcStr))
		C.free(unsafe.Pointer(sinkStr))
	}()
	tee_audio := C.gst_element_get_request_pad(server.TeeAudio, srcStr)
	webrtc_audio := C.gst_element_get_request_pad(g.Webrtc, sinkStr)
	reason = C.gst_pad_link(tee_audio, webrtc_audio)
	if reason != C.GST_PAD_LINK_OK {
		fmt.Println(strconv.Itoa(int(reason)))
	}
	tee_video := C.gst_element_get_request_pad(server.TeeAudio, srcStr)
	webrtc_video := C.gst_element_get_request_pad(g.Webrtc, sinkStr)
	reason = C.gst_pad_link(tee_video, webrtc_video)
	if reason != C.GST_PAD_LINK_OK {
		fmt.Println(strconv.Itoa(int(reason)))
	}

	fmt.Println("LINKED")

	g_signal_connect(unsafe.Pointer(g.Webrtc), "pad-added", C.on_incoming_stream_wrap, unsafe.Pointer(g))
	g_signal_connect(unsafe.Pointer(g.Webrtc), "on-negotiation-needed", C.on_negotiation_needed_wrap, unsafe.Pointer(g))
	g_signal_connect(unsafe.Pointer(g.Webrtc), "on-ice-candidate", C.send_ice_candidate_message_wrap, unsafe.Pointer(g))

	//g.loadBus()
	C.gst_element_set_state(server.pipeline, C.GST_STATE_PLAYING)
}
