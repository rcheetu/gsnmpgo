// Copyright 2012 Sonia Hamilton <sonia@snowfrog.net>. All rights
// reserved.  Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.
//
// stringers.go contains stringers for C enums and other types. To help with the
// generation of the boilerplate code, github.com/natefinch/gocog is used.
//
// AFTER EDITING any gocog sections (between gocog open and close square
// brackets), you MUST run:
//
// rm -f stringers.go_cog; $GOPATH/bin/gocog stringers.go
//
package gsnmp

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
#include <stdio.h>
#define MAX_OIDS_STR_LEN 1000

static void
oid_to_str(GList *list, char result[MAX_OIDS_STR_LEN]) {
	result[0] = '\0';
	while (list) {
		// assume an oid isn't longer than 200 characters
		if (strlen(result) > (MAX_OIDS_STR_LEN - 200)) {
			// run out of space, just append ...
			strcat(result, "...");
			return;
		}

		GList *next = list->next;
		GNetSnmpVarBind *vb = list->data;

		gint i;
		// assume an oid octet isn't longer than 20 characters
		char some_digits[20];
		for (i = 0; i < vb->oid_len; i++) {
			strcat(result, ".");
			sprintf(some_digits, "%i", vb->oid[i]);
			strcat(result, some_digits);
			some_digits[0] = '\0';
		}
		if (next != NULL) {
			strcat(result, ":");
		}
		list = next;
	}
}
*/
import "C"

import (
	"fmt"
	"strconv"
	"unsafe"
)

// returns a string represention of OIDs in vbl (var bind list)
//
// C:
//     /usr/include/glib-2.0/glib/glist.h
//     typedef struct _GList GList;
//     struct _GList
//     {
// 	       gpointer data;
// 	       GList *next;
// 	       GList *prev;
//     };
func OidToString(vbl *_Ctype_GList) string {
	// allocate "char result[MAX_OIDS_STR_LEN]"
	const MAX_OIDS_STR_LEN = 1000 // same as C code define
	result_go := fmt.Sprintf("%"+strconv.Itoa(MAX_OIDS_STR_LEN)+"s", " ")
	var result_c *C.char = C.CString(result_go)
	defer C.free(unsafe.Pointer(result_c))

	C.oid_to_str(vbl, result_c)
	return C.GoString(result_c)
}

// Stringer for *_Ctype_GURI
//
// Example:
//    fmt.Println("ParseURI():", parsed_uri)
//
// C:
//    /usr/include/gnet-2.0/uri.h
//    struct _GURI
//    {
//      gchar* scheme;
//      gchar* userinfo;
//      gchar* hostname;
//      gint   port;
//      gchar* path;
//      gchar* query;
//      gchar* fragment;
//    };
//
func (parsed_uri *_Ctype_GURI) String() string {
	scheme := C.GoString((*C.char)(parsed_uri.scheme))
	userinfo := C.GoString((*C.char)(parsed_uri.userinfo))
	hostname := C.GoString((*C.char)(parsed_uri.hostname))
	port := int(parsed_uri.port)
	path := C.GoString((*C.char)(parsed_uri.path))
	query := C.GoString((*C.char)(parsed_uri.query))
	fragment := C.GoString((*C.char)(parsed_uri.fragment))

	result := "{"
	result += "scheme:" + scheme + " "
	result += "userinfo:" + userinfo + " "
	result += "hostname:" + hostname + " "
	result += "port:" + strconv.Itoa(port) + " "
	result += "path:" + path + " "
	result += "query:" + query + " "
	result += "fragment:" + fragment
	result += "}"

	return result
}

// type and values for GNetSnmpUriType
type UriType int

const (
	GNET_SNMP_URI_GET UriType = iota
	GNET_SNMP_URI_NEXT
	GNET_SNMP_URI_WALK
)

// Stringer for _Ctype_GNetSnmpUriType
//
//    /usr/include/gsnmp/utils.h
//    typedef enum
//    {
//        GNET_SNMP_URI_GET,
//        GNET_SNMP_URI_NEXT,
//        GNET_SNMP_URI_WALK
//    } GNetSnmpUriType;
//
func (uritype _Ctype_GNetSnmpUriType) String() string {
	switch UriType(uritype) {
	case GNET_SNMP_URI_GET:
		return "GNET_SNMP_URI_GET"
	case GNET_SNMP_URI_NEXT:
		return "GNET_SNMP_URI_NEXT"
	case GNET_SNMP_URI_WALK:
		return "GNET_SNMP_URI_WALK"
	}
	return "UNKNOWN GNetSnmpUriType"
}

// Stringer for GString
//
//     glib/gstring.h
//     typedef struct _GString GString;
//     struct _GString {
//         gchar  *str;
//         gsize len;
//         gsize allocated_len;
//     };
//
// http://developer.gnome.org/glib/2.34/glib-Strings.html
// gchar *str - points to the character data. It may move as text is added. The
// str field is null-terminated and so can be used as an ordinary C string.
//
func (s _Ctype_GString) String() string {
	return C.GoString((*_Ctype_char)(s.str))
}

// Stringer for *_Ctype_GNetSnmp (a session)
//
// C:
//     gsnmp-0.3.0/src/session.h
//     typedef struct _GNetSnmp GNetSnmp;
//     struct _GNetSnmp {
//         GNetSnmpTAddress *taddress;
//         GURI             *uri;
//         gint32           error_status;
//         guint32          error_index;
//         guint            retries;        /* number of retries */
//         guint            timeout;        /* timeout in milliseconds */
//         GNetSnmpVersion  version;        /* message version */
//         GString          *ctxt_name;     /* context name */
//         GString          *sec_name;      /* security name */
//         GNetSnmpSecModel sec_model;      /* security model */
//         GNetSnmpSecLevel sec_level;      /* security level */
//         GNetSnmpDoneFunc done_callback;  /* what to call when complete */
//     }
//
func (s *_Ctype_GNetSnmp) String() string {

	error_status := strconv.Itoa(int(s.error_status))
	error_index := strconv.Itoa(int(s.error_index))
	retries := strconv.Itoa(int(s.retries))
	version := strconv.Itoa(int(s.version))

	result := "{"
	result += "taddress: TODO "
	result += "uri:" + fmt.Sprintf("%s", s.uri) + " "
	result += "error_status:" + error_status + " "
	result += "error_index:" + error_index + " "
	result += "retries:" + retries + " "
	result += "version:" + version + " "
	result += "context_name:" + fmt.Sprintf("%s", s.ctxt_name) + " "
	result += "security_name:" + fmt.Sprintf("%s", s.sec_name) + " "
	result += "security_model:" + fmt.Sprintf("%s", s.sec_model) + " "
	result += "security_level:" + fmt.Sprintf("%s", s.sec_level) + " "
	result += "done_callback: TODO "
	result += "}"
	return result
}

// type and values for GNetSnmpSecModel
type SecModel int

const (
	GNET_SNMP_SECMODEL_ANY SecModel = iota
	GNET_SNMP_SECMODEL_SNMPV1
	GNET_SNMP_SECMODEL_SNMPV2C
	GNET_SNMP_SECMODEL_SNMPV3
)

// Stringer for GNetSnmpSecModel
//
// C:
//    gsnmp-0.3.0/src/security.h
//    typedef enum {
//        GNET_SNMP_SECMODEL_ANY	= 0,
//        GNET_SNMP_SECMODEL_SNMPV1	= 1,
//        GNET_SNMP_SECMODEL_SNMPV2C	= 2,
//        GNET_SNMP_SECMODEL_SNMPV3	= 3
//    } GNetSnmpSecModel;
//
func (secmodel _Ctype_GNetSnmpSecModel) String() string {
	switch SecModel(secmodel) {
	case GNET_SNMP_SECMODEL_ANY:
		return "GNET_SNMP_SECMODEL_ANY"
	case GNET_SNMP_SECMODEL_SNMPV1:
		return "GNET_SNMP_SECMODEL_SNMPV1"
	case GNET_SNMP_SECMODEL_SNMPV2C:
		return "GNET_SNMP_SECMODEL_SNMPV2C"
	case GNET_SNMP_SECMODEL_SNMPV3:
		return "GNET_SNMP_SECMODEL_SNMPV3"
	}
	return "UNKNOWN GNetSnmpSecModel"
}

/*[[[gocog
package main
import ("github.com/soniah/gsnmp/enumconv")
func main() {

	ccode := `gsnmp-0.3.0/src/security.h
typedef enum {
    GNET_SNMP_SECLEVEL_NANP	= 0,
    GNET_SNMP_SECLEVEL_ANP	= 1,
    GNET_SNMP_SECLEVEL_AP	= 2
} GNetSnmpSecLevel;`

	vals := []string{"GNET_SNMP_SECLEVEL_NANP", "GNET_SNMP_SECLEVEL_ANP", "GNET_SNMP_SECLEVEL_AP"}
	enumconv.Write("SecLevel", "GNetSnmpSecLevel", vals, ccode)
}
gocog]]]*/

// type and values for GNetSnmpSecLevel
//
type SecLevel int

const (
	GNET_SNMP_SECLEVEL_NANP SecLevel = iota
	GNET_SNMP_SECLEVEL_ANP
	GNET_SNMP_SECLEVEL_AP
)

// Stringer for GNetSnmpSecLevel
//
// C:
//    gsnmp-0.3.0/src/security.h
//    typedef enum {
//        GNET_SNMP_SECLEVEL_NANP	= 0,
//        GNET_SNMP_SECLEVEL_ANP	= 1,
//        GNET_SNMP_SECLEVEL_AP	= 2
//    } GNetSnmpSecLevel;
//
func (seclevel _Ctype_GNetSnmpSecLevel) String() string {
	switch SecLevel(seclevel) {
	case GNET_SNMP_SECLEVEL_NANP:
		return "GNET_SNMP_SECLEVEL_NANP"
	case GNET_SNMP_SECLEVEL_ANP:
		return "GNET_SNMP_SECLEVEL_ANP"
	case GNET_SNMP_SECLEVEL_AP:
		return "GNET_SNMP_SECLEVEL_AP"
	}
	return "UNKNOWN GNetSnmpSecLevel"
}
//[[[end]]]
