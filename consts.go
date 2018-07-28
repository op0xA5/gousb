package gousb

/*
	#cgo CFLAGS: -I/usr/local/include/libusb-1.0
	#include <libusb.h>
*/
import "C"

// UsbSpeed type
type UsbSpeed int

// UsbSpeed values
const (
	UsbSpeedUnknown = UsbSpeed(C.LIBUSB_SPEED_UNKNOWN)
	UsbSpeedLow     = UsbSpeed(C.LIBUSB_SPEED_LOW)
	UsbSpeedFull    = UsbSpeed(C.LIBUSB_SPEED_FULL)
	UsbSpeedHigh    = UsbSpeed(C.LIBUSB_SPEED_HIGH)
	UsbSpeedSuper   = UsbSpeed(C.LIBUSB_SPEED_SUPER)
)

func (us UsbSpeed) String() string {
	switch us {
	case UsbSpeedLow:
		return "low"
	case UsbSpeedFull:
		return "full"
	case UsbSpeedHigh:
		return "high"
	case UsbSpeedSuper:
		return "super"
	}
	return "unknown"
}

// RequestType type
type RequestType uint8

// RequestType values
const (
	RequestTypeStandart = RequestType(0x00 << 5)
	RequestTypeClass    = RequestType(0x01 << 5)
	RequestTypeVendor   = RequestType(0x02 << 5)
	RequestTypeReserved = RequestType(0x03 << 5)
)

// RequestType values
const (
	RecipientDevice    = RequestType(0x00)
	RecipientInterface = RequestType(0x01)
	RecipientEndpoint  = RequestType(0x02)
	RecipientOther     = RequestType(0x02)
)

// RequestType values
const (
	EndpointIn  = RequestType(0x80)
	EndpointOut = RequestType(0x00)
)

// descriptor types
const (
	DescriptorTypeDevice    = uint8(0x01)
	DescriptorTypeConfig    = uint8(0x02)
	DescriptorTypeString    = uint8(0x03)
	DescriptorTypeInterface = uint8(0x04)
	DescriptorTypeEndpoint  = uint8(0x05)
	DescriptorTypeHid       = uint8(0x21)
	DescriptorTypeReport    = uint8(0x22)
	DescriptorTypePhysical  = uint8(0x23)
	DescriptorTypeHub       = uint8(0x29)
)

// ControlTransfer Requests
const (
	RequestGetStatus        = uint8(0x00)
	RequestClearFeature     = uint8(0x01)
	RequestSetFeature       = uint8(0x03)
	RequestSetAddress       = uint8(0x05)
	RequestGetDescriptor    = uint8(0x06)
	RequestSetDescriptor    = uint8(0x07)
	RequestGetConfiguration = uint8(0x08)
	RequestSetConfiguration = uint8(0x09)
	RequestGetInterface     = uint8(0x0A)
	RequestSetInterface     = uint8(0x0B)
	RequestSynchFrame       = uint8(0x0C)
)

// TransferType type
type TransferType int

// TransferType values
const (
	TransferTypeControl     = TransferType(0)
	TransferTypeIsochronous = TransferType(1)
	TransferTypeBulk        = TransferType(2)
	TransferTypeInterrupt   = TransferType(3)
)

// Error type
type Error int

// Error values
const (
	ErrSuccess      = Error(C.LIBUSB_SUCCESS)
	ErrIo           = Error(C.LIBUSB_ERROR_IO)
	ErrInvalidParam = Error(C.LIBUSB_ERROR_INVALID_PARAM)
	ErrAccess       = Error(C.LIBUSB_ERROR_ACCESS)
	ErrNoDevice     = Error(C.LIBUSB_ERROR_NO_DEVICE)
	ErrNotFound     = Error(C.LIBUSB_ERROR_NOT_FOUND)
	ErrBusy         = Error(C.LIBUSB_ERROR_BUSY)
	ErrTimeout      = Error(C.LIBUSB_ERROR_TIMEOUT)
	ErrOverflow     = Error(C.LIBUSB_ERROR_OVERFLOW)
	ErrPipe         = Error(C.LIBUSB_ERROR_PIPE)
	ErrInterrupted  = Error(C.LIBUSB_ERROR_INTERRUPTED)
	ErrNoMem        = Error(C.LIBUSB_ERROR_NO_MEM)
	ErrNotSupported = Error(C.LIBUSB_ERROR_NOT_SUPPORTED)
	ErrOther        = Error(C.LIBUSB_ERROR_OTHER)
)

func (err Error) Error() string {
	switch err {
	case ErrSuccess:
		return "success"
	case ErrIo:
		return "io error"
	case ErrInvalidParam:
		return "invalid param"
	case ErrAccess:
		return "error access"
	case ErrNoDevice:
		return "no device"
	case ErrNotFound:
		return "resource not found"
	case ErrBusy:
		return "busy"
	case ErrTimeout:
		return "timeout"
	case ErrOverflow:
		return "overflow"
	case ErrPipe:
		return "error pipe"
	case ErrInterrupted:
		return "interrupted"
	case ErrNoMem:
		return "no mem"
	case ErrNotSupported:
		return "not supported"
	case ErrOther:
		return "other"
	}
	return ""
}

// USB Class Codes
const (
	ClassInterfaceSpecific   = uint8(0x00)
	ClassAudio               = uint8(0x01)
	ClassCDCControl          = uint8(0x02)
	ClassHID                 = uint8(0x03)
	ClassPhysical            = uint8(0x05)
	ClassImage               = uint8(0x06)
	ClassPrinter             = uint8(0x07)
	ClassMassStorage         = uint8(0x08)
	ClassHub                 = uint8(0x09)
	ClassCDCData             = uint8(0x0a)
	ClassSmartCard           = uint8(0x0b)
	ClassContentSecurity     = uint8(0x0d)
	ClassVideo               = uint8(0x0e)
	ClassersonalHealthcare   = uint8(0x0f)
	ClassAudioVideoDevices   = uint8(0x10)
	ClassBillboardDevice     = uint8(0x11)
	ClassUSBTypeCBridge      = uint8(0x12)
	ClassDiagnosticDevice    = uint8(0xdc)
	ClassWirelessController  = uint8(0xe0)
	ClassMiscellaneous       = uint8(0xef)
	ClassApplicationSpecific = uint8(0xfe)
	ClassVendorSpecific      = uint8(0xff)
)
