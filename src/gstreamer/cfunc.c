#define GST_USE_UNSTABLE_API
#include <gst/webrtc/webrtc.h>
#include <glib.h>
#include <gst/gst.h>
#include <gst/gstbin.h>
#include <json-glib/json-glib.h>
#include <string.h>
#include <types.h>
#include <stdio.h>
#include <stdbool.h>

gboolean print_field (GQuark field, const GValue * value, gpointer pfx) {
  gchar *str = gst_value_serialize (value);
  g_print ("%s  %15s: %s\n", (gchar *) pfx, g_quark_to_string (field), str);
  g_free (str);
  return TRUE;
}

void print_caps (const GstCaps * caps, const gchar * pfx) {
  guint i;
  g_return_if_fail (caps != NULL);
  if (gst_caps_is_any (caps)) {
    g_print ("%sANY\n", pfx);
    return;
  }
  if (gst_caps_is_empty (caps)) {
    g_print ("%sEMPTY\n", pfx);
    return;
  }
  for (i = 0; i < gst_caps_get_size (caps); i++) {
    GstStructure *structure = gst_caps_get_structure (caps, i);
    g_print ("%s%s\n", pfx, gst_structure_get_name (structure));
    gst_structure_foreach (structure, print_field, (gpointer) pfx);
  }
}

void print_pad_capabilities (GstElement *element, gchar *pad_name) {
  	GstPad *pad = NULL;
 	GstCaps *caps = NULL;
	pad = gst_element_get_static_pad(element, pad_name);
	if (!pad) {
		g_printerr ("Could not retrieve pad '%s'\n", pad_name);
		return;
	}
	caps = gst_pad_get_current_caps (pad);
	if (!caps)
		caps = gst_pad_query_caps (pad, NULL);
	g_print ("Caps for the %s pad:\n", pad_name);
	print_caps(caps, "      ");
	gst_caps_unref (caps);
	gst_object_unref (pad);
}

extern gboolean bus_call (GstBus *bus, GstMessage *msg, void *data);
gboolean bus_call_wrap (GstBus *bus, GstMessage *msg, void *data)
{
  return bus_call(bus, msg, data);
}

extern void on_answer_created (GstPromise * promise, void * user_data);
void on_answer_created_wrap (GstPromise * promise, gpointer user_data) {
    on_answer_created(promise, user_data);
}

extern void on_negotiation_needed (GstElement * webrtc, void* user_data);
void on_negotiation_needed_wrap (GstElement * webrtc, void* user_data)
{
    on_negotiation_needed(webrtc, user_data);
}

extern void on_offer_set (GstPromise * webrtc, void* user_data);
void on_offer_set_wrap(GstPromise * webrtc, void* user_data)
{
    on_offer_set(webrtc, user_data);
}

extern void on_offer_created (GstPromise * webrtc, void * user_data);
void on_offer_created_wrap (GstPromise *promise, void *user_data)
{
    on_offer_created(promise, user_data);
//	g_print ("on_offer_created:\n");
//	GstWebRTCSessionDescription *offer = NULL;
//	const GstStructure *reply;
//	gchar *desc;
//	reply = gst_promise_get_reply (promise);
//	gst_structure_get (reply, "offer", GST_TYPE_WEBRTC_SESSION_DESCRIPTION, &offer, NULL);
//	g_signal_emit_by_name (webrtc, "set-local-description", offer, NULL);
//	gst_webrtc_session_description_free (offer);
}

extern void send_ice_candidate_message (GstElement * webrtc G_GNUC_UNUSED, guint mlineindex, gchar * candidate, void *user_data);
void send_ice_candidate_message_wrap (GstElement * webrtc G_GNUC_UNUSED, guint mlineindex, gchar * candidate, void *user_data)
{
    send_ice_candidate_message(webrtc, mlineindex, candidate, user_data);
}

extern void on_incoming_stream (GstElement * webrtc, GstPad * pad, GstElement * pipe);
void on_incoming_stream_wrap (GstElement * webrtc, GstPad * pad, GstElement * pipe)
{
    on_incoming_stream(webrtc, pad, pipe);
}

GstWebRTCRTPTransceiver *g_array_index_wrap(GArray *a,int i) {
    return  g_array_index(a, GstWebRTCRTPTransceiver*, i);
}

void g_object_set_fec(GstWebRTCRTPTransceiver* trans) {
    g_object_set(trans, "fec-type", GST_WEBRTC_FEC_TYPE_ULP_RED, "do-nack", TRUE, NULL);
}

void sendKeyFrame(GstPad * pad) {
    gst_pad_send_event(pad, gst_event_new_custom( GST_EVENT_CUSTOM_UPSTREAM, gst_structure_new( "GstForceKeyUnit", "all-headers", G_TYPE_BOOLEAN, TRUE, NULL)));
}

void gst_caps_set_simple_wrap(GstCaps *caps, char *field, int type, void *value) {
    gst_caps_set_simple (caps, field, type, value, NULL);
}

void g_object_set_wrap(gpointer object_type, gchar *first_property_name, void *three) {
    g_object_set(object_type, first_property_name, three, NULL);
}

void g_object_set_bool_wrap(gpointer object_type, gchar *first_property_name, bool three) {
    g_object_set(object_type, first_property_name, three, NULL);
}
GstCaps *gst_caps_set_format() {
    return gst_caps_new_simple("video/x-h264", "stream-format", G_TYPE_STRING, "avc", NULL);
}
