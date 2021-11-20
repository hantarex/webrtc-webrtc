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
	"unsafe"
)

type WsStore interface {
	AddServerUser(key string) (err error)
}

type PassWebrtc struct {
	g      *GStreamer
	webrtc *C.GstElement
}

type GStreamer struct {
	Webrtc, pipeline, videotestsrc, teeVideo, nvh264enc, rtph264pay, queue, h264parse, queue1, autovideosink *C.GstElement
	gError                                                                                                   *C.GError
	//send_channel *C.GObject
	bus *C.GstBus
	//loop         *C.GMainLoop
	ret         C.GstStateChangeReturn
	C           *websocket.Conn
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

func (g *GStreamer) Close(code int, text string) (err error) {
	g.cancel()
	log.Println("Connection closed: ", g.C.RemoteAddr().String(), " ", g.C.RemoteAddr().Network())
	C.gst_element_set_state(g.pipeline, C.GST_STATE_NULL)
	//C.g_main_loop_quit(g.loop)
	if g.trans != nil {
		C.gst_object_unref(C.gpointer(g.trans))
	}
	C.gst_object_unref(C.gpointer(g.bus))
	//C.gst_object_unref(C.gpointer(g.send_channel))
	C.gst_object_unref(C.gpointer(g.pipeline))
	//C.g_main_loop_unref(g.loop)
	return
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

func (g *GStreamer) loadBus() {
	g.bus = gst_pipeline_get_bus(unsafe.Pointer(g.pipeline))
	go func(bus *C.GstBus) {
		for {
			msg := C.gst_bus_timed_pop_filtered(bus, C.GST_CLOCK_TIME_NONE,
				C.GST_MESSAGE_STATE_CHANGED|C.GST_MESSAGE_ERROR|C.GST_MESSAGE_WARNING|C.GST_MESSAGE_EOS|C.GST_MESSAGE_STREAM_STATUS)
			if msg != nil {
				switch msg._type {
				case C.GST_MESSAGE_ERROR:
					{
						var debug *C.gchar
						var gError *C.GError

						C.gst_message_parse_error(msg, &gError, &debug)
						fmt.Printf("Error: %s\n", C.GoString(gError.message))
						C.g_error_free(gError)
						break
					}

				case C.GST_MESSAGE_STATE_CHANGED:
					{
						break
					}
				case C.GST_MESSAGE_BUFFERING:
					{
						break
					}
				case C.GST_MESSAGE_ELEMENT:
					{
						break
					}
				case C.GST_MESSAGE_STREAM_STATUS:
					{
						break

					}
				case C.GST_MESSAGE_STREAM_START:
					{
						break

					}
				default:
					fmt.Println(msg._type)
					break
				}
				C.gst_message_unref(msg)
			}
		}
	}(g.bus)
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
		return
	}
	fmt.Println(C.GoString(text))
	err := g.C.WriteJSON(Message{
		Id:        "startResponse",
		SdpAnswer: C.GoString(text),
	})
	C.g_free(C.gpointer(text))
	if err != nil {
		log.Println("sendSpdToPeer:", err)
	}
}

func (g GStreamer) sendIceCandidate(ice string) {
	var msg Message
	if err := json.Unmarshal([]byte(ice), &msg); err != nil {
		log.Printf("Сбой демаршалинга JON: %s\n", err)
	}
	err := g.C.WriteJSON(Message{
		Id:        "iceCandidate",
		Candidate: msg.Candidate,
	})
	if err != nil {
		log.Println("iceCandidate:", err)
	}
}

func (g *GStreamer) On_offer_received(msg Message, dst *C.GstElement, ws WsStore, noAnswer bool) (err error) {
	fmt.Println("on_offer_received")
	if msg.Key == "" {
		err = errors.New("key of stream does not exists")
	}

	g.setRTMPKey(msg.Key)
	ws.AddServerUser(msg.Key)

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

func (g *GStreamer) On_answer_received(msg Message, dst *C.GstElement) (err error) {
	fmt.Println("On_answer_received")
	fmt.Println(msg)

	var sdp *C.GstSDPMessage
	C.gst_sdp_message_new(&sdp)
	spdStr := C.CString(msg.SdpOffer)
	defer C.free(unsafe.Pointer(spdStr))
	C.gst_sdp_message_parse_buffer_wrap(spdStr, C.strlen(spdStr), sdp)
	//
	var offer *C.GstWebRTCSessionDescription
	var promise *C.GstPromise
	//
	offer = C.gst_webrtc_session_description_new(C.GST_WEBRTC_SDP_TYPE_ANSWER, sdp)
	promise = C.gst_promise_new_with_change_func(C.GCallback(C.on_offer_set_wrap), C.gpointer(&PassWebrtc{
		g:      g,
		webrtc: dst,
	}), nil)
	g_signal_emit_by_name_offer_remote(dst, "set-remote-description", offer, promise)
	return
}

func (g *GStreamer) IceCandidateReceived(msg Message, webrtc *C.GstElement) {
	fmt.Println("IceCandidateReceived")
	if msg.Candidate.Candidate == "" {
		//g_signal_emit_by_name(webrtc, "add-ice-candidate", nil, nil, nil)
		return
	}
	fmt.Println(msg)
	canStr := C.CString(msg.Candidate.Candidate)
	defer C.free(unsafe.Pointer(canStr))
	g_signal_emit_by_name_recv(webrtc, "add-ice-candidate", msg.Candidate.SdpMLineIndex, unsafe.Pointer(C.gchararray(canStr)), nil)
}

func (g *GStreamer) setRTMPKey(key string) (err error) {
	g.RtmpKey = key
	g.ret = C.gst_element_set_state(g.pipeline, C.GST_STATE_PLAYING)
	if g.ret == C.GST_STATE_CHANGE_FAILURE {
		err = errors.New("Unable to set the pipeline to the playing state (check the bus for error messages).")
	}
	return
}
