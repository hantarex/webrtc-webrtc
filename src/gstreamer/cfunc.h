#define GST_USE_UNSTABLE_API
#include <gst/webrtc/webrtc.h>
#include <glib.h>
#include <gst/sdp/sdp.h>
#include <gst/gst.h>
//#include <gst/gstbin.h>
#include <json-glib/json-glib.h>
#include <string.h>
#include <types.h>
#include <stdbool.h>

gboolean bus_call (GstBus *bus, GstMessage *msg, void *data);
void on_negotiation_needed_wrap (GstElement * webrtc, gpointer user_data);
void on_offer_created_wrap (GstPromise *promise, GstElement *webrtc);
void on_offer_set_wrap (GstPromise * promise, gpointer user_data);
void on_answer_created_wrap (GstPromise * promise, gpointer user_data);
gboolean bus_call_wrap (GstBus *bus, GstMessage *msg, void *data);
void send_ice_candidate_message_wrap (GstElement * webrtc G_GNUC_UNUSED, guint mlineindex, gchar * candidate, gpointer user_data G_GNUC_UNUSED);
void on_incoming_stream_wrap (GstElement * webrtc, GstPad * pad, GstElement * pipe);
GstSDPResult gst_sdp_message_parse_buffer_wrap(gchar *data, ulong size, GstSDPMessage *msg);
GstWebRTCRTPTransceiver *g_array_index_wrap(GArray *a, int i);
void g_object_set_fec(GstWebRTCRTPTransceiver* trans);
void gst_caps_set_simple_wrap(GstCaps *caps, char *field, int type, void *value);
void g_object_set_wrap(gpointer object_type, gchar *first_property_name, void *three);
void g_object_set_bool_wrap(gpointer object_type, gchar *first_property_name, bool three);
GstCaps *gst_caps_set_format();
void sendKeyFrame(GstPad * pad);
void g_object_set_int_wrap(gpointer object_type, gchar *first_property_name, int three);
GArray* g_array_index_zero(GstElement *webrtc);