package gsnmpgo

// Copyright 2013 Sonia Hamilton <sonia@snowfrog.net>. All rights
// reserved.  Use of this source code is governed by a 3-clause BSD
// license that can be found in the LICENSE file.

// glib typedefs - http://developer.gnome.org/glib/2.35/glib-Basic-Types.html
// glib tutorial - http://www.dlhoffman.com/publiclibrary/software/gtk+-html-docs/gtk_tut-17.html
// gsnmp sourcecode browser - http://sourcecodebrowser.com/gsnmp/0.3.0/index.html

/*
#cgo pkg-config: glib-2.0 gsnmp
#include <gsnmp/ber.h>
#include <gsnmp/pdu.h>
#include <gsnmp/dispatch.h>
#include <gsnmp/message.h>
#include <gsnmp/security.h>
#include <gsnmp/session.h>
#include <gsnmp/transport.h>
#include <gsnmp/utils.h>
#include <gsnmp/gsnmp.h>
#include <stdlib.h>

// convenience wrapper for gnet_snmp_enum_get_label()
gchar const *
get_err_label(gint32 const id) {
	return gnet_snmp_enum_get_label(gnet_snmp_enum_error_table, id);
}
*/
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"
)

var _ = reflect.DeepEqual(0, 1) // dummy

type QueryResult struct {
	Oid   string
	Value Varbinder
}

type QueryResults []QueryResult

// Snmp Query - main entry point to library.
func Query(uri string, version SnmpVersion) (results QueryResults, err error) {
	parsed_uri, err := ParseURI(uri)
	if err != nil {
		return nil, err
	}

	vbl, uritype, err := ParsePath(uri, parsed_uri)
	defer UriDelete(parsed_uri)
	if err != nil {
		return nil, err
	}

	session, err := NewUri(uri, version, parsed_uri)
	if err != nil {
		return nil, err
	}

	switch UriType(uritype) {
	case GNET_SNMP_URI_GET:
		vbl_results, err := CGet(session, vbl)
		if err != nil {
			return nil, err
		}
		return ConvertResults(vbl_results), nil // TODO no err from decode?
	case GNET_SNMP_URI_NEXT:
		fmt.Println("doing GNET_SNMP_URI_NEXT")
	case GNET_SNMP_URI_WALK:
		fmt.Println("doing GNET_SNMP_URI_WALK")
	}
	panic(fmt.Sprintf("%s: Query(): fell out of switch", libname()))
}

// dump results - convenience function
func Dump(results QueryResults) {
	fmt.Println("Dump:")
	for _, result := range results {
		fmt.Printf("%T:%s:%s\n", result.Value, result.Oid, result.Value)
	}
}

// ParseURI parses an SNMP URI into fields.
//
// The generic URI parsing is done by gnet_uri_new(), and the SNMP specific
// portions by gnet_snmp_parse_uri(). Only basic URI validation is done here,
// more is done by ParsePath()
//
// Example:
//
//    uri := `snmp://public@192.168.1.10//(1.3.6.1.2.1.1.1.0,1.3.6.1.2.1.1.2.0)`
//    parsed_uri, err := gsnmpgo.ParseURI(uri)
//    if err != nil {
//    	fmt.Println(err)
//    	os.Exit(1)
//    }
//    fmt.Println("ParseURI():", parsed_uri)
func ParseURI(uri string) (parsed_uri *_Ctype_GURI, err error) {
	curi := (*C.gchar)(C.CString(uri))
	defer C.free(unsafe.Pointer(curi))

	var gerror *C.GError
	parsed_uri = C.gnet_snmp_parse_uri(curi, &gerror)
	if parsed_uri == nil {
		return nil, fmt.Errorf("%s: invalid snmp uri: %s", libname(), uri)
	}
	return parsed_uri, nil
}

// ParsePath parses an SNMP URI.
//
// The uritype will default to GNET_SNMP_URI_GET. If the uri ends in:
//
// '*' the uritype will be GNET_SNMP_URI_WALK
//
// '+' the uritype will be GNET_SNMP_URI_NEXT
//
// See RFC 4088 "Uniform Resource Identifier (URI) Scheme for the Simple
// Network Management Protocol (SNMP)" for further documentation.
func ParsePath(uri string, parsed_uri *_Ctype_GURI) (vbl *_Ctype_GList, uritype _Ctype_GNetSnmpUriType, err error) {
	var gerror *C.GError
	rv := C.gnet_snmp_parse_path(parsed_uri.path, &vbl, &uritype, &gerror)
	if rv == 0 {
		err_string := C.GoString((*_Ctype_char)(gerror.message))
		return vbl, uritype, fmt.Errorf("%s: %s: <%s>", libname(), err_string, uri)
	}
	return vbl, uritype, nil
}

// UriDelete frees the memory used by a parsed_uri.
//
// A defered call to UriDelete should be made after ParsePath().
func UriDelete(parsed_uri *_Ctype_GURI) {
	C.gnet_uri_delete(parsed_uri)
}

// NewUri creates a session from a parsed uri.
func NewUri(uri string, version SnmpVersion, parsed_uri *_Ctype_GURI) (session *_Ctype_GNetSnmp, err error) {
	var gerror *C.GError
	session = C.gnet_snmp_new_uri(parsed_uri, &gerror)

	// error handling
	if gerror != nil {
		err_string := C.GoString((*_Ctype_char)(gerror.message))
		C.g_clear_error(&gerror)
		return session, fmt.Errorf("%s: %s", libname(), err_string)
	}
	if session == nil {
		return session, fmt.Errorf("%s: unable to create session", libname())
	}
	session.version = (_Ctype_GNetSnmpVersion)(version)

	// results
	return session, nil
}

// Do an SNMP Get.
//
// Results are returned in C form, use ConvertResults() to convert to a Go struct.
func CGet(session *_Ctype_GNetSnmp, vbl *_Ctype_GList) (*_Ctype_GList, error) {
	var gerror *C.GError
	out := C.gnet_snmp_sync_get(session, vbl, &gerror)

	// error handling
	if gerror != nil {
		err_string := C.GoString((*_Ctype_char)(gerror.message))
		C.g_clear_error(&gerror)
		return out, fmt.Errorf("%s: %s", libname(), err_string)
	}
	if PduError(session.error_status) != GNET_SNMP_PDU_ERR_NOERROR {
		es := C.get_err_label(session.error_status)
		err_string := C.GoString((*_Ctype_char)(es))
		return out, fmt.Errorf("%s: %s for uri %d", libname(), err_string, session.error_index)
	}

	// results
	return out, nil
}

// ConvertResults converts C results to a Go struct.
func ConvertResults(out *_Ctype_GList) (results QueryResults) {
	for {
		if out == nil {
			// finished
			return results
		}

		// another result: initialise
		data := (*C.GNetSnmpVarBind)(out.data)
		oid := GIntArrayOidString(data.oid, data.oid_len)
		result := QueryResult{Oid: oid}
		var value Varbinder

		// convert C values to Go values
		vbt := VarBindType(data._type)
		switch vbt {

		case GNET_SNMP_VARBIND_TYPE_NULL:
			value = new(VBT_Null)

		case GNET_SNMP_VARBIND_TYPE_OCTETSTRING:
			value = VBT_OctetString(union_ui8v_string(data.value, data.value_len))

		case GNET_SNMP_VARBIND_TYPE_OBJECTID:
			guint32_ptr := union_ui32v(data.value)
			value = VBT_ObjectID(GIntArrayOidString(guint32_ptr, data.value_len))

		case GNET_SNMP_VARBIND_TYPE_IPADDRESS:
			value = VBT_IPAddress(union_ui8v_ipaddress(data.value, data.value_len))

		case GNET_SNMP_VARBIND_TYPE_INTEGER32:
			value = VBT_Integer32(union_i32(data.value))

		case GNET_SNMP_VARBIND_TYPE_UNSIGNED32:
			value = VBT_Unsigned32(union_ui32(data.value))

		case GNET_SNMP_VARBIND_TYPE_COUNTER32:
			value = VBT_Counter32(union_ui32(data.value))

		case GNET_SNMP_VARBIND_TYPE_TIMETICKS:
			value = VBT_Timeticks(union_ui32(data.value))

		case GNET_SNMP_VARBIND_TYPE_OPAQUE:
			value = VBT_Opaque(union_ui8v_hexstring(data.value, data.value_len))

		case GNET_SNMP_VARBIND_TYPE_COUNTER64:
			value = VBT_Counter64(union_ui64(data.value))

		case GNET_SNMP_VARBIND_TYPE_NOSUCHOBJECT:
			value = new(VBT_NoSuchObject)

		case GNET_SNMP_VARBIND_TYPE_NOSUCHINSTANCE:
			value = new(VBT_NoSuchInstance)

		case GNET_SNMP_VARBIND_TYPE_ENDOFMIBVIEW:
			value = new(VBT_EndOfMibView)
		}
		result.Value = value
		results = append(results, result)

		// move on to next element in list
		out = out.next
	}
	panic(fmt.Sprintf("%s: ConvertResults(): fell out of for loop", libname()))
}