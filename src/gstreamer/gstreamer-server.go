package gstreamer

/*
#cgo pkg-config: gstreamer-plugins-bad-1.0 gstreamer-rtp-1.0 gstreamer-webrtc-1.0 gstreamer-plugins-base-1.0 glib-2.0 libsoup-2.4 json-glib-1.0
#cgo CFLAGS: -Wall
#cgo CFLAGS: -Wno-deprecated-declarations -Wimplicit-function-declaration -Wformat-security
#cgo LDFLAGS: -lgstsdp-1.0
#include <cfunc.h>
*/
import "C"
import "unsafe"

func (g *GStreamer) InitGstServer() {
	C.gst_init(nil, nil)
	C.gst_debug_set_default_threshold(C.GST_LEVEL_WARNING)
	pipeStr := C.CString("webrtcbin latency=5000 stun-server=stun://stun.l.google.com:19302 name=webrtcbin message-forward=true webrtcbin. ! " +
		"rtph264depay request-keyframe=true ! " +
		"h264parse ! queue2 use-buffering=true ! mux. webrtcbin. ! " +
		"rtpopusdepay ! " +
		"opusdec max-errors=-1 ! audioconvert ! avenc_aac ! queue2 use-buffering=true ! mux. flvmux latency=2000 min-upstream-latency=2000 name=mux emit-signals=true streamable=true ! " +
		"filesink location=test.flv")
	defer C.free(unsafe.Pointer(pipeStr))
	g.pipeline = C.gst_parse_launch(pipeStr, &g.gError)

	webrtcName := C.CString("webrtcbin")
	defer C.free(unsafe.Pointer(webrtcName))
	g.Webrtc = C.gst_bin_get_by_name(GST_BIN(g.pipeline), C.CString("webrtcbin"))

	//capsStr := C.CString("application/x-rtp,media=video,encoding-name=H264,clock-rate=90000")
	//defer C.free(unsafe.Pointer(capsStr))
	//var caps *C.GstCaps = C.gst_caps_from_string(capsStr)
	//
	//g_signal_emit_by_name_trans(g.Webrtc, "add-transceiver", C.GST_WEBRTC_RTP_TRANSCEIVER_DIRECTION_RECVONLY, unsafe.Pointer(caps))

	g_signal_connect(unsafe.Pointer(g.Webrtc), "pad-added", C.on_incoming_stream_wrap, unsafe.Pointer(g))

	g_signal_connect(unsafe.Pointer(g.Webrtc), "on-negotiation-needed", C.on_negotiation_needed_wrap, unsafe.Pointer(g))
	g_signal_connect(unsafe.Pointer(g.Webrtc), "on-ice-candidate", C.send_ice_candidate_message_wrap, unsafe.Pointer(g))

	g.loadBus()
	C.gst_element_set_state(g.pipeline, C.GST_STATE_PLAYING)
}
