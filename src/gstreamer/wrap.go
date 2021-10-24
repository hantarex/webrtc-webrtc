package gstreamer

/*
#define GST_USE_UNSTABLE_API
#include <gst/gst.h>
#include <json-glib/json-glib.h>
#include <gst/webrtc/webrtc.h>

void g_assert_nonnull_wrap(gpointer expr) {
	g_assert_nonnull(expr);
}

GstBin *GST_BIN_WRAP(GstElement *r) {
	return GST_BIN(r);
}

gulong g_signal_connect_wrap(gpointer instance, gchar *detailed_signal, GCallback c_handler, gpointer data) {
	return g_signal_connect(instance, detailed_signal, c_handler, data);
}

void g_signal_emit_by_name_wrap(GstElement *instance,char* signal,void* one,void* two,void* three) {
	g_signal_emit_by_name(instance, signal, one, two, three);
}

void g_signal_emit_by_name_offer_wrap(GstElement *instance,char* signal,GstWebRTCSessionDescription* one) {
	g_signal_emit_by_name(instance, signal, one, NULL);
}

void g_signal_emit_by_name_offer_remote_wrap(GstElement *instance,char* signal,GstWebRTCSessionDescription* one, GstPromise* two) {
	g_signal_emit_by_name(instance, signal, one, two, NULL);
}

GstSDPResult gst_sdp_message_parse_buffer_wrap(gchar *data, ulong size, GstSDPMessage *msg) {
	return gst_sdp_message_parse_buffer((guint8 *) data, size, msg);
}

void g_signal_emit_by_name_recv_wrap(GstElement *instance,char* signal,int one,void* two,void* three) {
	g_signal_emit_by_name(instance, signal, one, two, three);
}

void g_signal_emit_by_name_trans(GstElement *instance,char* signal,int one,void* two) {
	GstWebRTCRTPTransceiver *trans = NULL;
	g_signal_emit_by_name(instance, signal, one, two, &trans);
	g_object_set(trans, "fec-type", GST_WEBRTC_FEC_TYPE_ULP_RED, "do-nack", TRUE, NULL);
	if( trans != NULL ) {
		gst_object_unref (trans);
	}
}

void g_print_wrap(gchar *str) {
	g_print(str, NULL);
}

GstBus *gst_pipeline_get_bus_wrap(void *pipeline) {
	return gst_pipeline_get_bus(GST_PIPELINE(pipeline));
}

gboolean gst_structure_get_wrap(GstStructure  *structure,char * first_fieldname, ulong one, GstWebRTCSessionDescription** two,void* three) {
	return gst_structure_get(structure, first_fieldname, one, &*two, three, NULL);
}
*/
import "C"
import (
	"unsafe"
)

func g_assert_nonnull(r C.gpointer) {
	C.g_assert_nonnull_wrap(r)
}

func GST_BIN(r *C.GstElement) *C.GstBin {
	return C.GST_BIN_WRAP(r)
}

func g_signal_connect(instance unsafe.Pointer, detailed_signal string, c_handler unsafe.Pointer, data unsafe.Pointer) C.gulong {
	instance_c := C.gpointer(instance)
	detailed_signal_c := C.CString(detailed_signal)
	defer C.free(unsafe.Pointer(detailed_signal_c))
	c_handler_c := C.GCallback(c_handler)
	data_c := C.gpointer(data)
	return C.g_signal_connect_wrap(instance_c, detailed_signal_c, c_handler_c, data_c)
}

func g_signal_emit_by_name(instance *C.GstElement, signal string, one unsafe.Pointer, two unsafe.Pointer, three unsafe.Pointer) {
	sigC := C.CString(signal)
	defer C.free(unsafe.Pointer(sigC))
	C.g_signal_emit_by_name_wrap(instance, sigC, one, two, three)
}

func g_signal_emit_by_name_offer(instance *C.GstElement, signal string, one *C.GstWebRTCSessionDescription) {
	sigC := C.CString(signal)
	defer C.free(unsafe.Pointer(sigC))
	C.g_signal_emit_by_name_offer_wrap(instance, sigC, one)
}

func g_signal_emit_by_name_offer_remote(instance *C.GstElement, signal string, one *C.GstWebRTCSessionDescription, two *C.GstPromise) {
	sigC := C.CString(signal)
	defer C.free(unsafe.Pointer(sigC))
	C.g_signal_emit_by_name_offer_remote_wrap(instance, sigC, one, two)
}

func g_signal_emit_by_name_recv(instance *C.GstElement, signal string, one int, two unsafe.Pointer, three unsafe.Pointer) {
	sigC := C.CString(signal)
	defer C.free(unsafe.Pointer(sigC))
	C.g_signal_emit_by_name_recv_wrap(instance, sigC, C.int(one), two, three)
}

func g_signal_emit_by_name_trans(instance *C.GstElement, signal string, one int, two unsafe.Pointer) {
	sigC := C.CString(signal)
	defer C.free(unsafe.Pointer(sigC))
	C.g_signal_emit_by_name_trans(instance, sigC, C.int(one), two)
}

func g_print(str string) {
	s := C.CString(str)
	defer C.free(unsafe.Pointer(s))
	C.g_print_wrap(s)
}

func gst_pipeline_get_bus(r unsafe.Pointer) *C.GstBus {
	return C.gst_pipeline_get_bus_wrap(r)
}

func gst_structure_get(a1 *C.GstStructure, a2 string, a3 C.ulong, a4 *C.GstWebRTCSessionDescription, a5 unsafe.Pointer) C.gboolean {
	offer := new(C.GstWebRTCSessionDescription)
	a2c := C.CString(a2)
	defer C.free(unsafe.Pointer(a2c))
	r := C.gst_structure_get_wrap(a1, a2c, a3, &offer, a5)
	*a4 = *offer
	return r
}

func get_string_from_json_object(object *C.JsonObject) *C.gchar {
	var root *C.JsonNode
	var generator *C.JsonGenerator
	var text *C.gchar

	/* Make it the root node */
	root = C.json_node_init_object(C.json_node_alloc(), object)
	generator = C.json_generator_new()
	C.json_generator_set_root(generator, root)
	text = C.json_generator_to_data(generator, nil)

	/* Release everything */
	C.g_object_unref(C.gpointer(generator))
	C.json_node_free(root)
	return text
}
