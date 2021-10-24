package gstreamer

/*
#include <gst/gst.h>
#include <cfunc.h>
*/
import "C"
import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
	"unsafe"
)

//export on_negotiation_needed
func on_negotiation_needed(webrtc *C.GstElement, user_data unsafe.Pointer) {
	fmt.Println("on_negotiation_needed")
	promise := C.gst_promise_new_with_change_func(C.GCallback(C.on_offer_created_wrap), C.gpointer(user_data), nil)
	g_signal_emit_by_name((*GStreamer)(user_data).webrtc, "create-offer", nil, unsafe.Pointer(promise), nil)
}

//export on_offer_set
func on_offer_set(promise *C.GstPromise, user_data unsafe.Pointer) {
	C.gst_promise_unref(promise)
	promise = C.gst_promise_new_with_change_func(C.GCallback(C.on_answer_created_wrap), C.gpointer(user_data), nil)
	g_signal_emit_by_name((*PassWebrtc)(user_data).webrtc, "create-answer", nil, unsafe.Pointer(promise), nil)
}

//export on_answer_created
func on_answer_created(promise *C.GstPromise, user_data unsafe.Pointer) {
	g := (*PassWebrtc)(user_data)
	answer := new(C.GstWebRTCSessionDescription)

	reply := C.gst_promise_get_reply(promise)
	gst_structure_get(reply, "answer", C.GST_TYPE_WEBRTC_SESSION_DESCRIPTION, answer, nil)
	C.gst_promise_unref(promise)
	//
	promise = C.gst_promise_new()
	g_signal_emit_by_name(g.webrtc, "set-local-description", unsafe.Pointer(answer), unsafe.Pointer(promise), nil)
	C.gst_promise_unref(promise)
	///* Send answer to peer */
	g.g.sendSpdToPeer(answer)
	//fmt.Println("free")
	//C.gst_webrtc_session_description_free(answer)
}

//export on_offer_created
func on_offer_created(promise *C.GstPromise, webrtc unsafe.Pointer) {
	fmt.Println("on_offer_created")
	g := (*GStreamer)(webrtc)
	offer := new(C.GstWebRTCSessionDescription)
	var reply *C.GstStructure
	//defer C.free(unsafe.Pointer(reply))
	reply = C.gst_promise_get_reply(promise)
	gst_structure_get(reply, "offer", C.GST_TYPE_WEBRTC_SESSION_DESCRIPTION, offer, nil)

	g_signal_emit_by_name_offer(g.webrtc, "set-local-description", offer)
	g.sendSpdToPeer(offer)
	/* Implement this and send offer to peer using signalling */
	//	send_sdp_offer (offer);
}

//export bus_call
func bus_call(bus *C.GstBus, msg *C.GstMessage, data unsafe.Pointer) C.gboolean {
	g := (*GStreamer)(data)
	switch msg._type {
	case C.GST_MESSAGE_ERROR:
		{
			var debug *C.gchar
			var gError *C.GError

			C.gst_message_parse_error(msg, &gError, &debug)
			log.Printf("Error: %s\n", C.GoString(gError.message))
			C.g_error_free(gError)
			g.c.Close()
			break
		}
	default:
		break
	}
	return 1
}

//export on_incoming_stream
func on_incoming_stream(webrtc *C.GstElement, pad *C.GstPad, user_data unsafe.Pointer) {
	fmt.Println("on_incoming_stream " + C.GoString(webrtc.object.name))
	g := (*GStreamer)(user_data)
	sinkName := C.CString("sink")
	defer C.free(unsafe.Pointer(sinkName))
	srcName := C.CString("src_%u")
	defer C.free(unsafe.Pointer(srcName))
	new_pad_caps := C.gst_pad_get_current_caps(pad)
	fmt.Println(new_pad_caps)
	if new_pad_caps == nil {
		fmt.Println("Load webrtc client")
		//C.gst_element_link(g.queue, g.rtph264depay)
		srcpad := C.gst_element_get_request_pad(g.teeVideo, srcName)
		//sinkpad := C.gst_element_get_static_pad(g.rtph264depay1, C.CString("sink"))
		//fmt.Println("sinkpad")
		//fmt.Println(sinkpad)
		reason := C.gst_pad_link(srcpad, pad)
		if reason != C.GST_PAD_LINK_OK {
			fmt.Println(errors.New(strconv.Itoa(int(reason))).Error())
		}
		return
	}
	new_pad_struct := C.gst_caps_get_structure(new_pad_caps, 0)
	media := C.CString("media")
	defer C.free(unsafe.Pointer(media))
	typePad := C.GoString(C.gst_structure_get_string(new_pad_struct, media))
	//fmt.Println(C.GoString(C.gst_structure_serialize(new_pad_struct, C.GST_SERIALIZE_FLAG_NONE)))
	//fmt.Println(C.GoString(C.gst_structure_get_string(new_pad_struct, C.CString("media"))))
	if typePad == "video" {
		fmt.Println("receive pad " + typePad)
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second * 2):
				}
				C.sendKeyFrame(pad)
			}

		}(g.ctx)
		sinkpad := C.gst_element_get_static_pad(g.teeVideo, sinkName)
		defer C.gst_object_unref(C.gpointer(sinkpad))
		if C.gst_pad_is_linked(sinkpad) == 1 {
			fmt.Println("We are already linked. Ignoring.\n")
			return
		}
		C.gst_pad_link(pad, sinkpad)
	}

	//if typePad == "audio" {
	//	fmt.Println("receive pad " + typePad)
	//	g.initAudio()
	//	sinkpad := C.gst_element_get_static_pad(g.rtpopusdepay, sinkName)
	//	defer C.gst_object_unref(C.gpointer(sinkpad))
	//	if C.gst_pad_is_linked(sinkpad) == 1 {
	//		fmt.Println("We are already linked. Ignoring.\n")
	//		return
	//	}
	//	C.gst_pad_link(pad, sinkpad)
	//	//g_print ("It has type '%s' is raw audio.\n", new_pad_type);
	//	//sink_pad = gst_element_get_static_pad (targetPad->elementToAudio, "sink");
	//	//if (gst_pad_is_linked (sink_pad)) {
	//	//	g_print ("We are already linked. Ignoring.\n");
	//	//	goto exit;
	//	//}
	//	//ret = gst_pad_link (new_pad, sink_pad);
	//}

	//if C.GST_PAD_DIRECTION(pad) != C.GST_PAD_SRC {
	//	fmt.Println("Pad is not source")
	//}
	//new_pad_caps := C.gst_pad_get_current_caps (pad)
	//new_pad_struct := C.gst_caps_get_structure (new_pad_caps, 0)
	//fmt.Println(pad.direction)
	//fmt.Println(C.GST_PAD_SRC)
	//new_pad_type = gst_structure_get_name (new_pad_struct);
}

//export send_ice_candidate_message
func send_ice_candidate_message(webrtc *C.GstElement, mlineindex C.long, candidate *C.gchar, user_data unsafe.Pointer) {
	g := (*GStreamer)(user_data)
	//
	//   if (app_state < PEER_CALL_NEGOTIATING) {
	//   	g_print ("Can't send ICE, not in call", APP_STATE_ERROR);
	//       return;
	//   }
	//
	ice := C.json_object_new()
	candidateStr := C.CString("candidate")
	defer C.free(unsafe.Pointer(candidateStr))
	sdpMLineIndex := C.CString("sdpMLineIndex")
	defer C.free(unsafe.Pointer(sdpMLineIndex))
	C.json_object_set_string_member(ice, candidateStr, (*C.gchar)(candidate))
	C.json_object_set_int_member(ice, sdpMLineIndex, mlineindex)
	msg := C.json_object_new()
	iceStr := C.CString("candidate")
	defer C.free(unsafe.Pointer(iceStr))
	C.json_object_set_object_member(msg, iceStr, ice)
	text := get_string_from_json_object(msg)
	defer C.g_free(C.gpointer(text))
	//C.g_free(C.gpointer(text))
	g.sendIceCandidate(C.GoString(text))
}

func g_object_set(object C.gpointer, f1 string, f2 unsafe.Pointer) {
	f1Name := C.CString(f1)
	defer C.free(unsafe.Pointer(f1Name))
	C.g_object_set_wrap(object, f1Name, f2)
}

func g_object_set_bool(object C.gpointer, f1 string, f2 bool) {
	f1Name := C.CString(f1)
	defer C.free(unsafe.Pointer(f1Name))
	C.g_object_set_bool_wrap(object, f1Name, C.bool(f2))
}
