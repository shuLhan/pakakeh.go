// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package xmlrpc provide an implementation of XML-RPC specification,
// http://xmlrpc.com/spec.md.
package xmlrpc

const (
	schemeIsHTTPS     = "https"
	tagXML            = "xml"
	timeLayoutISO8601 = "20060102T15:04:05"
)

// Kind define the known type in Value.
//
// This is looks like the reflect.Kind but limited only to specific types
// defined in XML-RPC.
type Kind int

// List of available Kind.
const (
	Unset    Kind = iota
	String        // represent Go string type.
	Boolean       // represent Go bool type.
	Integer       // represent Go int8, int16, int32, uint8, and uint16 types.
	Double        // represent Go uint32, uint64, float32, and float64 types.
	DateTime      // represent Go time.Time type.
	Base64        // represent Go string type.
	Struct        // represent Go struct type.
	Array         // represent Go array and slice types.
)

const (
	elNameMethodCall     = "methodCall"
	elNameMethodName     = "methodName"
	elNameMethodResponse = "methodResponse"

	elNameData   = "data"
	elNameFault  = "fault"
	elNameMember = "member"
	elNameName   = "name"
	elNameParam  = "param"
	elNameParams = "params"
	elNameValue  = "value"

	memberNameFaultCode   = "faultCode"
	memberNameFaultString = "faultString"

	typeNameArray    = "array"
	typeNameBase64   = "base64"
	typeNameBoolean  = "boolean"
	typeNameDateTime = "datetime.iso8601"
	typeNameDouble   = "double"
	typeNameInteger  = "int"
	typeNameInteger4 = "i4"
	typeNameString   = "string"
	typeNameStruct   = "struct"
)
