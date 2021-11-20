package gstreamer

/*
#include <gst/gst.h>
#include <cfunc.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

//export on_negotiation_needed
func on_negotiation_needed(webrtc *C.GstElement, user_data unsafe.Pointer) {
	g := (*GStreamer)(user_data)
	fmt.Println("on_negotiation_needed")
	//capsStr := C.CString("application/x-rtp,media=video,encoding-name=H264,clock-rate=90000")
	//defer C.free(unsafe.Pointer(capsStr))
	//var caps *C.GstCaps = C.gst_caps_from_string(capsStr)
	//g_signal_emit_by_name_trans(g.Webrtc, "add-transceiver", C.GST_WEBRTC_RTP_TRANSCEIVER_DIRECTION_SENDONLY, unsafe.Pointer(caps))
	//transceivers := C.g_array_index_zero(g.Webrtc)
	//t := C.g_array_index_wrap(transceivers, 0)
	//fmt.Println(t)
	promise := C.gst_promise_new_with_change_func(C.GCallback(C.on_offer_created_wrap), C.gpointer(user_data), nil)
	g_signal_emit_by_name(g.Webrtc, "create-offer", nil, unsafe.Pointer(promise), nil)
}

//export on_offer_set
func on_offer_set(promise *C.GstPromise, user_data unsafe.Pointer) {
	C.gst_promise_unref(promise)
	//var transceiver *C.GstWebRTCRTPTransceiver
	//transceiver = C.g_array_index_zero(g.g.Webrtc)
	//fmt.Println(transceiver)
	//C.g_array_index_zero(g.g.Webrtc)

	promise = C.gst_promise_new_with_change_func(C.GCallback(C.on_answer_created_wrap), C.gpointer(user_data), nil)
	g_signal_emit_by_name((*PassWebrtc)(user_data).webrtc, "create-answer", nil, unsafe.Pointer(promise), nil)

}

//export on_offer_set_client
func on_offer_set_client(promise *C.GstPromise, user_data unsafe.Pointer) {
	C.gst_promise_unref(promise)
}

//export on_answer_created
func on_answer_created(promise *C.GstPromise, user_data unsafe.Pointer) {
	fmt.Println("on_answer_created")
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
	g_signal_emit_by_name_offer(g.Webrtc, "set-local-description", offer)
	g.sendSpdToPeer(offer)
	///* Implement this and send offer to peer using signalling */
	//g.sendSpdToPeer (offer);
}

//export bus_call
func bus_call(bus *C.GstBus, msg *C.GstMessage, data unsafe.Pointer) C.gboolean {
	//g := (*GStreamer)(data)
	//switch msg._type {
	//case C.GST_MESSAGE_ERROR:
	//	{
	//		var debug *C.gchar
	//		var gError *C.GError
	//
	//		C.gst_message_parse_error(msg, &gError, &debug)
	//		log.Printf("Error: %s\n", C.GoString(gError.message))
	//		C.g_error_free(gError)
	//		g.c.Close()
	//		break
	//	}
	//default:
	//	break
	//}
	return 1
}

func g_object_int(object C.gpointer, f1 string, f2 int) {
	f1Name := C.CString(f1)
	defer C.free(unsafe.Pointer(f1Name))
	C.g_object_set_int_wrap(object, f1Name, C.int(f2))
}

//export on_incoming_stream
func on_incoming_stream(webrtc *C.GstElement, pad *C.GstPad, user_data unsafe.Pointer) {
	new_pad_caps := C.gst_pad_get_current_caps(pad)
	new_pad_struct := C.gst_caps_get_structure(new_pad_caps, 0)
	media := C.CString("media")
	defer C.free(unsafe.Pointer(media))
	typePad := C.GoString(C.gst_structure_get_string(new_pad_struct, media))
	fmt.Println("receive pad " + typePad)
}

//export send_ice_candidate_message
func send_ice_candidate_message(webrtc *C.GstElement, mlineindex C.long, candidate *C.gchar, user_data unsafe.Pointer) {
	fmt.Println("send_ice_candidate_message " + C.GoString(webrtc.object.name))
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
