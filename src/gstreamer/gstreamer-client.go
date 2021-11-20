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
	"unsafe"
)

func (g *GStreamer) InitGstClient(server *GStreamer) {
	C.gst_init(nil, nil)
	C.gst_debug_set_default_threshold(C.GST_LEVEL_WARNING)
	pipeStr := C.CString("queue2 name=queueAudio ! rtpopuspay ! webrtc. queue2 name=queueVideo ! rtph264pay pt=96 ! webrtc. webrtcbin name=webrtc stun-server=stun://stun.l.google.com:19302")
	defer C.free(unsafe.Pointer(pipeStr))
	g.pipeline = C.gst_parse_launch(pipeStr, &g.GError)

	fmt.Println(g.pipeline)

	webrtcName := C.CString("webrtc")
	defer C.free(unsafe.Pointer(webrtcName))
	g.Webrtc = C.gst_bin_get_by_name(GST_BIN(g.pipeline), webrtcName)

	queueAudioName := C.CString("queueAudio")
	defer C.free(unsafe.Pointer(queueAudioName))
	g.queue = C.gst_bin_get_by_name(GST_BIN(g.pipeline), queueAudioName)

	queueVideoName := C.CString("queueVideo")
	defer C.free(unsafe.Pointer(queueVideoName))
	g.queue1 = C.gst_bin_get_by_name(GST_BIN(g.pipeline), queueVideoName)

	capsStr := C.CString("application/x-rtp,media=video,encoding-name=H264,clock-rate=90000")
	defer C.free(unsafe.Pointer(capsStr))
	var caps *C.GstCaps = C.gst_caps_from_string(capsStr)
	g_signal_emit_by_name_trans(g.Webrtc, "add-transceiver", C.GST_WEBRTC_RTP_TRANSCEIVER_DIRECTION_SENDONLY, unsafe.Pointer(caps))

	//var reason C.GstPadLinkReturn
	//srcStr := C.CString("src_%u")
	//sinkStr := C.CString("sink")
	//defer func() {
	//	C.free(unsafe.Pointer(srcStr))
	//	C.free(unsafe.Pointer(sinkStr))
	//}()
	//tee_audio := C.gst_element_get_request_pad(server.TeeAudio, srcStr)
	//webrtc_audio := C.gst_element_get_static_pad(g.queue, sinkStr)
	//reason = C.gst_pad_link(tee_audio, webrtc_audio)
	//if reason != C.GST_PAD_LINK_OK {
	//	fmt.Println(strconv.Itoa(int(reason)))
	//}
	//
	//tee_video := C.gst_element_get_request_pad(server.TeeVideo, srcStr)
	//webrtc_video := C.gst_element_get_static_pad(g.queue1, sinkStr)
	//reason = C.gst_pad_link(tee_video, webrtc_video)
	//if reason != C.GST_PAD_LINK_OK {
	//	fmt.Println(strconv.Itoa(int(reason)))
	//}

	g_signal_connect(unsafe.Pointer(g.Webrtc), "pad-added", C.on_incoming_stream_wrap, unsafe.Pointer(g))
	g_signal_connect(unsafe.Pointer(g.Webrtc), "on-negotiation-needed", C.on_negotiation_needed_wrap, unsafe.Pointer(g))
	g_signal_connect(unsafe.Pointer(g.Webrtc), "on-ice-candidate", C.send_ice_candidate_message_wrap, unsafe.Pointer(g))

	g.loadBus()
	C.gst_element_set_state(g.pipeline, C.GST_STATE_PLAYING)
}
