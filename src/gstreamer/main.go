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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
	"time"
	"unsafe"
)

type PassWebrtc struct {
	g      *GStreamer
	webrtc *C.GstElement
}

type GStreamer struct {
	webrtc, webrtc1, pipeline, avdec_h264, rtph264depay, teeVideo, videoconvert, queue, autovideosink *C.GstElement
	gError                                                                                            *C.GError
	//send_channel *C.GObject
	bus *C.GstBus
	//loop         *C.GMainLoop
	ret         C.GstStateChangeReturn
	c           *websocket.Conn
	trans       *C.GstWebRTCRTPTransceiver
	RtmpAddress string
	RtmpKey     string
	Iter        int
	ctx         context.Context
	cancel      func()
}

func (g *GStreamer) teeLink(source *C.GstElement, target *C.GstElement, srcStrName string, tgtSrtName string) (err error) {
	srcStr := C.CString(srcStrName)
	sinkStr := C.CString(tgtSrtName)
	defer func() {
		C.free(unsafe.Pointer(srcStr))
		C.free(unsafe.Pointer(sinkStr))
	}()
	//fmt.Printf("Obtained request pad %s.\n", string(C.GoString(C.gst_pad_get_name_wrap(target))))
	tee_pad := C.gst_element_get_request_pad(source, srcStr)
	target_pad := C.gst_element_get_static_pad(target, sinkStr)

	reason := C.gst_pad_link(tee_pad, target_pad)
	if reason != C.GST_PAD_LINK_OK {
		return errors.New(strconv.Itoa(int(reason)))
	}
	return
}

func (g *GStreamer) Close() {
	g.cancel()
	g.c.Close()
	log.Println("Connection closed: ", g.c.RemoteAddr().String(), " ", g.c.RemoteAddr().Network())
	C.gst_element_set_state(g.pipeline, C.GST_STATE_NULL)
	//C.g_main_loop_quit(g.loop)
	if g.trans != nil {
		C.gst_object_unref(C.gpointer(g.trans))
	}
	C.gst_object_unref(C.gpointer(g.bus))
	//C.gst_object_unref(C.gpointer(g.send_channel))
	C.gst_object_unref(C.gpointer(g.pipeline))
	//C.g_main_loop_unref(g.loop)
}

type IceCandidate struct {
	Candidate     string `json:"candidate"`
	SdpMid        string `json:"sdpMid,omitempty"`
	SdpMLineIndex int    `json:"sdpMLineIndex"`
}

type Message struct {
	SdpAnswer string       `json:"sdpAnswer,omitempty"`
	SdpOffer  string       `json:"sdpOffer,omitempty"`
	Candidate IceCandidate `json:"candidate,omitempty"`
	Id        string       `json:"id,omitempty"`
	Key       string       `json:"key,omitempty"`
}

func (g *GStreamer) InitConnection(c *websocket.Conn) {
	g.c = c
	log.Println("Connected: ", g.c.RemoteAddr().String(), " ", g.c.RemoteAddr().Network())
	ctx, cancel := context.WithCancel(context.Background())
	g.ctx = ctx
	g.cancel = cancel
	g.InitGst()
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
			if err := g.c.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				g.Close()
			}
		}
	}(g.ctx)
	go g.readMessages()
}

func (g *GStreamer) InitGst() {
	C.gst_init(nil, nil)
	C.gst_debug_set_default_threshold(C.GST_LEVEL_WARNING)
	//pipeStr := C.CString("webrtcbin bundle-policy=max-bundle ice-tcp=false name=recv recv. ! rtph264depay ! queue ! avdec_h264 ! videoconvert ! queue ! autovideosink")
	//pipeStr := C.CString("webrtcbin stun-server=stun://stun.l.google.com:19302 name=recv recv. ! queue2 max-size-buffers=0 max-size-time=0 max-size-bytes=0 ! rtph264depay ! queue2 ! h264parse ! video/x-h264,stream-format=(string)avc ! queue2 ! avdec_h264 ! queue2 ! videoconvert ! queue ! autovideosink")
	//pipeStr := C.CString("webrtcbin stun-server=stun://stun.l.google.com:19302 name=recv recv. ! queue2 max-size-buffers=0 max-size-time=0 max-size-bytes=0 ! rtph264depay ! queue2 ! h264parse ! flvmux ! rtmp2sink sync=false location=rtmp://localhost:1935/hls_dash/${name}_mid")
	//pipeStr := C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtph264depay ! avdec_h264 ! queue ! x264enc ! flvmux ! filesink location=xyz.flv")
	//defer C.free(unsafe.Pointer(pipeStr))
	//g.pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtpvp8depay ! vp8dec ! videoconvert ! queue ! autovideosink"), &g.gError)
	//g.pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtph264depay ! avdec_h264 ! queue ! autovideosink"), &g.gError)
	//g.pipeline = C.gst_parse_launch(C.CString("webrtcbin bundle-policy=max-bundle stun-server=stun://stun.l.google.com:19302 name=recv recv. ! rtph264depay request-keyframe=1 ! avdec_h264 ! queue ! x264enc ! flvmux ! filesink location=xyz.flv"), &g.gError)
	pipeName := C.CString("j2c_webrtc")
	defer C.free(unsafe.Pointer(pipeName))
	g.pipeline = C.gst_pipeline_new(pipeName)
	// webrtcbin
	webrtcName := C.CString("webrtcbin")
	defer C.free(unsafe.Pointer(webrtcName))
	g.webrtc = C.gst_element_factory_make(webrtcName, webrtcName)
	g_object_set(C.gpointer(g.webrtc), "stun-server", unsafe.Pointer(C.CString("stun://stun.l.google.com:19302")))
	//g_object_set(C.gpointer(g.webrtc), "stun-server", unsafe.Pointer(C.CString("stun://stun.l.google.com:19302")))
	// tee video
	teeVideoName := C.CString("tee")
	defer C.free(unsafe.Pointer(teeVideoName))
	teeVideoNameDesc := C.CString("teeVideo")
	defer C.free(unsafe.Pointer(teeVideoNameDesc))
	g.teeVideo = C.gst_element_factory_make(teeVideoName, teeVideoNameDesc)
	// rtph264depay
	rtph264depayName := C.CString("rtph264depay")
	defer C.free(unsafe.Pointer(rtph264depayName))
	rtph264depayDescName := C.CString("rtph264depay")
	defer C.free(unsafe.Pointer(rtph264depayDescName))
	g.rtph264depay = C.gst_element_factory_make(rtph264depayName, rtph264depayDescName)
	// h264parse
	avdec_h264Name := C.CString("avdec_h264")
	defer C.free(unsafe.Pointer(avdec_h264Name))
	avdec_h264DescName := C.CString("avdec_h264")
	defer C.free(unsafe.Pointer(avdec_h264DescName))
	g.avdec_h264 = C.gst_element_factory_make(avdec_h264Name, avdec_h264DescName)
	// videoconvert
	videoconvertName := C.CString("videoconvert")
	defer C.free(unsafe.Pointer(videoconvertName))
	videoconvertDescName := C.CString("videoconvert")
	defer C.free(unsafe.Pointer(videoconvertDescName))
	g.videoconvert = C.gst_element_factory_make(videoconvertName, videoconvertDescName)
	// autovideosink
	autovideosinkName := C.CString("autovideosink")
	defer C.free(unsafe.Pointer(autovideosinkName))
	autovideosinkDescName := C.CString("autovideosink")
	defer C.free(unsafe.Pointer(autovideosinkDescName))
	g.autovideosink = C.gst_element_factory_make(autovideosinkName, autovideosinkDescName)
	// queue
	queueName := C.CString("queue2")
	defer C.free(unsafe.Pointer(queueName))
	g.queue = C.gst_element_factory_make(queueName, queueName)
	// webrtcbin1
	webrtcName1 := C.CString("webrtcbin")
	defer C.free(unsafe.Pointer(webrtcName1))
	webrtcName1Desc := C.CString("webrtcName1")
	defer C.free(unsafe.Pointer(webrtcName1Desc))
	g.webrtc1 = C.gst_element_factory_make(webrtcName1, webrtcName1Desc)
	g_object_set(C.gpointer(g.webrtc1), "stun-server", unsafe.Pointer(C.CString("stun://stun.l.google.com:19302")))
	//g_object_set(C.gpointer(g.webrtc1), "bundle-policy", unsafe.Pointer(C.CString("max-compat")))

	C.gst_bin_add(GST_BIN(g.pipeline), g.webrtc)
	C.gst_bin_add(GST_BIN(g.pipeline), g.rtph264depay)
	C.gst_bin_add(GST_BIN(g.pipeline), g.avdec_h264)
	C.gst_bin_add(GST_BIN(g.pipeline), g.videoconvert)
	C.gst_bin_add(GST_BIN(g.pipeline), g.autovideosink)
	C.gst_bin_add(GST_BIN(g.pipeline), g.teeVideo)
	C.gst_bin_add(GST_BIN(g.pipeline), g.queue)
	C.gst_bin_add(GST_BIN(g.pipeline), g.webrtc1)

	C.gst_element_link(g.rtph264depay, g.avdec_h264)
	C.gst_element_link(g.avdec_h264, g.videoconvert)
	C.gst_element_link(g.videoconvert, g.autovideosink)

	//C.gst_element_link(g.queue, g.rtph264depay)
	//C.gst_element_link(g.rtph264depay, g.avdec_h264)
	//C.gst_element_link(g.avdec_h264, g.videoconvert)
	//C.gst_element_link(g.videoconvert, g.autovideosink)

	if err := g.teeLink(g.teeVideo, g.rtph264depay, "src_%u", "sink"); err != nil {
		fmt.Println("Tee video not linked: " + err.Error())
	}

	if err := g.teeLink(g.teeVideo, g.queue, "src_%u", "sink"); err != nil {
		fmt.Println("Tee video not linked: " + err.Error())
	}
	//
	//if err := g.teeLink(g.queue, g.webrtc1, "src", "sink_%u"); err != nil {
	//	fmt.Println("Tee queue not webrtc1: " + err.Error())
	//}

	g_signal_connect(unsafe.Pointer(g.webrtc), "pad-added", C.on_incoming_stream_wrap, unsafe.Pointer(g))
	g_signal_connect(unsafe.Pointer(g.webrtc1), "pad-added", C.on_incoming_stream_wrap, unsafe.Pointer(g))

	//g_signal_connect(unsafe.Pointer(g.webrtc), "on-negotiation-needed", C.on_negotiation_needed_wrap, unsafe.Pointer(g))
	g_signal_connect(unsafe.Pointer(g.webrtc), "on-ice-candidate", C.send_ice_candidate_message_wrap, unsafe.Pointer(g))
	g_signal_connect(unsafe.Pointer(g.webrtc1), "on-ice-candidate", C.send_ice_candidate_message_wrap, unsafe.Pointer(g))

	//C.gst_element_set_state(g.pipeline, C.GST_STATE_READY)

	//g_signal_emit_by_name(g.webrtc, "create-data-channel", unsafe.Pointer(C.CString("channel")), nil, unsafe.Pointer(&g.send_channel))
	//g_signal_emit_by_name(g.webrtc, "add-local-ip-address", unsafe.Pointer(C.CString("127.0.0.1")), nil, nil)

	capsStr := C.CString("application/x-rtp,media=video,encoding-name=H264,clock-rate=90000")
	defer C.free(unsafe.Pointer(capsStr))
	var caps *C.GstCaps = C.gst_caps_from_string(capsStr)
	//C.gst_caps_set_simple_wrap(caps,  C.CString("extmap"), C.G_TYPE_STRING, unsafe.Pointer(C.CString("http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time")))

	//g.trans = new(C.GstWebRTCRTPTransceiver)
	g_signal_emit_by_name_trans(g.webrtc, "add-transceiver", C.GST_WEBRTC_RTP_TRANSCEIVER_DIRECTION_RECVONLY, unsafe.Pointer(caps))
	g_signal_emit_by_name_trans(g.webrtc1, "add-transceiver", C.GST_WEBRTC_RTP_TRANSCEIVER_DIRECTION_SENDONLY, unsafe.Pointer(caps))
	//C.g_object_set_fec(g.trans)

	//if g.send_channel != nil {
	//	fmt.Println("Created data channel")
	//} else {
	//	fmt.Println("Could not create data channel, is usrsctp available?")
	//}

	//g.loop = C.g_main_loop_new(nil, 0)

	g.bus = gst_pipeline_get_bus(unsafe.Pointer(g.pipeline))
	C.gst_bus_add_signal_watch(g.bus)
	g_signal_connect(unsafe.Pointer(g.bus), "message", C.bus_call_wrap, unsafe.Pointer(g))
	//C.g_main_loop_run(g.loop)
}

func (g GStreamer) sendSpdToPeer(desc *C.GstWebRTCSessionDescription) {
	//if (app_state < PEER_CALL_NEGOTIATING) {
	//	cleanup_and_quit_loop ("Can't send SDP to peer, not in call",
	//		APP_STATE_ERROR);
	//	return;
	//}

	//media := C.gst_sdp_message_get_media(desc.sdp, 1)
	//
	//var caps *C.GstCaps = new(C.GstCaps)
	//C.gst_caps_set_simple_wrap(caps,  C.CString("extmap"), C.G_TYPE_STRING, unsafe.Pointer(C.CString("http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time")))
	//C.gst_sdp_media_attributes_to_caps(media, caps)

	text := C.gst_sdp_message_as_text(desc.sdp)

	if desc._type == C.GST_WEBRTC_SDP_TYPE_OFFER {
		//fmt.Printf("Sending offer:\n%s\n", C.GoString(text))
		fmt.Println("Sending offer")
	} else if desc._type == C.GST_WEBRTC_SDP_TYPE_ANSWER {
		//fmt.Printf("Sending answer:\n%s\n", C.GoString(text))
		fmt.Println("Sending answer offer")
	} else {
		log.Println("sendSpdToPeer:", "type not found")
		g.c.Close()
		return
	}
	fmt.Println(C.GoString(text))
	err := g.c.WriteJSON(Message{
		Id:        "startResponse",
		SdpAnswer: C.GoString(text),
	})
	C.g_free(C.gpointer(text))
	if err != nil {
		log.Println("sendSpdToPeer:", err)
		g.c.Close()
	}
}

func (g GStreamer) sendIceCandidate(ice string) {
	var msg Message
	if err := json.Unmarshal([]byte(ice), &msg); err != nil {
		log.Printf("Сбой демаршалинга JON: %s\n", err)
		g.c.Close()
	}
	err := g.c.WriteJSON(Message{
		Id:        "iceCandidate",
		Candidate: msg.Candidate,
	})
	if err != nil {
		log.Println("iceCandidate:", err)
		g.c.Close()
	}
}

func (g *GStreamer) readMessages() {
	defer g.Close()
	for {
		var msg Message
		_, message, err := g.c.ReadMessage()

		if err != nil {
			log.Println("read:", err)
			break
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Сбой демаршалинга JON: %s\n", err)
			break
		}
		switch msg.Id {
		case "start":
			if err := g.on_offer_received(msg, g.webrtc); err != nil {
				log.Println(err.Error())
			}
			break
		case "client_start":
			if err := g.on_offer_received(msg, g.webrtc1); err != nil {
				log.Println(err.Error())
			}
			break
		case "onIceCandidate":
			g.iceCandidateReceived(msg, g.webrtc)
			break
		case "onIceCandidateClient":
			g.iceCandidateReceived(msg, g.webrtc1)
			break
		default:
			log.Println("Error readMessages")
		}
	}
}

func (g *GStreamer) on_offer_received(msg Message, dst *C.GstElement) (err error) {
	fmt.Println("on_offer_received")
	if msg.Key == "" {
		err = errors.New("key of stream does not exists")
	}
	g.setRTMPKey(msg.Key)

	var sdp *C.GstSDPMessage
	C.gst_sdp_message_new(&sdp)
	spdStr := C.CString(msg.SdpOffer)
	defer C.free(unsafe.Pointer(spdStr))
	C.gst_sdp_message_parse_buffer_wrap(spdStr, C.strlen(spdStr), sdp)

	var offer *C.GstWebRTCSessionDescription
	var promise *C.GstPromise

	offer = C.gst_webrtc_session_description_new(C.GST_WEBRTC_SDP_TYPE_OFFER, sdp)
	promise = C.gst_promise_new_with_change_func(C.GCallback(C.on_offer_set_wrap), C.gpointer(&PassWebrtc{
		g:      g,
		webrtc: dst,
	}), nil)
	g_signal_emit_by_name_offer_remote(dst, "set-remote-description", offer, promise)
	return
}

func (g *GStreamer) iceCandidateReceived(msg Message, webrtc *C.GstElement) {
	if msg.Candidate.Candidate == "" {
		//g_signal_emit_by_name(g.webrtc, "add-ice-candidate", nil, nil, nil)
		return
	}
	fmt.Println(msg)
	canStr := C.CString(msg.Candidate.Candidate)
	defer C.free(unsafe.Pointer(canStr))
	g_signal_emit_by_name_recv(webrtc, "add-ice-candidate", msg.Candidate.SdpMLineIndex, unsafe.Pointer(C.gchararray(canStr)), nil)
}

func (g *GStreamer) setRTMPKey(key string) {
	g.RtmpKey = key
	g.ret = C.gst_element_set_state(g.pipeline, C.GST_STATE_PLAYING)
	if g.ret == C.GST_STATE_CHANGE_FAILURE {
		fmt.Println("Unable to set the pipeline to the playing state (check the bus for error messages).")
	}
}
