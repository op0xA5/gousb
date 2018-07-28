package gousb

/*
#cgo CFLAGS: -I/usr/local/include/libusb-1.0
#cgo LDFLAGS: /usr/local/lib/libusb-1.0.dylib
#include <stdlib.h>
#include <libusb.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"io"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)

type Context struct {
	handle *C.struct_libusb_context
}

var libusb_ctx *Context

func Init() *Context {
	if libusb_ctx != nil {
		return libusb_ctx
	}
	libusb_ctx = &Context{
		handle: (*C.struct_libusb_context)(unsafe.Pointer(C.malloc(C.size_t(unsafe.Sizeof(C.struct_libusb_context{}))))),
	}
	rc := C.libusb_init(&libusb_ctx.handle)
	if rc < 0 {
		return nil
	}
	return libusb_ctx
}
func Exit() {
	if libusb_ctx != nil {
		C.libusb_exit(libusb_ctx.handle)
		C.free(unsafe.Pointer(libusb_ctx.handle))
		libusb_ctx = nil
	}
}

type Device struct {
	list **C.struct_libusb_device
	ptr  *C.struct_libusb_device

	Bus     uint8
	Port    uint8
	Address uint8
	Speed   UsbSpeed

	DeviceDescriptor
}

func newDevice(list **C.struct_libusb_device, ptr *C.struct_libusb_device) *Device {
	var desc C.struct_libusb_device_descriptor
	C.libusb_get_device_descriptor(ptr, &desc)
	return &Device{
		list: list,
		ptr:  ptr,

		Bus:     uint8(C.libusb_get_bus_number(ptr)),
		Port:    uint8(C.libusb_get_port_number(ptr)),
		Address: uint8(C.libusb_get_device_address(ptr)),
		Speed:   UsbSpeed(int(C.libusb_get_device_speed(ptr))),

		DeviceDescriptor: DeviceDescriptor{
			Length:          uint8(desc.bLength),
			DescriptorType:  uint8(desc.bDescriptorType),
			BcdUSB:          uint16(desc.bcdUSB),
			DeviceClass:     uint8(desc.bDeviceClass),
			DeviceSubClass:  uint8(desc.bDeviceSubClass),
			DeviceProtol:    uint8(desc.bDeviceProtocol),
			MaxPacketSize0:  uint8(desc.bMaxPacketSize0),
			IDVender:        uint16(desc.idVendor),
			IDProduct:       uint16(desc.idProduct),
			BcdDevice:       uint16(desc.bcdDevice),
			IdxManufacturer: uint8(desc.iManufacturer),
			IdxProduct:      uint8(desc.iProduct),
			IdxSerialNumber: uint8(desc.iSerialNumber),
			NumConfiguation: uint8(desc.bNumConfigurations),
		},
	}
}

type DeviceList []*Device

func GetDeviceList() (DeviceList, error) {
	var devs **C.struct_libusb_device
	ctx := Init()
	rc := int(C.libusb_get_device_list(ctx.handle, &devs))
	if rc == 0 {
		return nil, nil
	}
	if rc < 0 {
		return nil, Error(rc)
	}

	list := make([]*Device, rc)
	dev_ptr := uintptr(unsafe.Pointer(devs))
	for i := range list {
		list[i] = newDevice(devs, *(**C.struct_libusb_device)(unsafe.Pointer(dev_ptr)))
		dev_ptr += unsafe.Sizeof(dev_ptr)
	}
	return list, nil
}
func (list DeviceList) Close() {
	if list == nil {
		return
	}
	C.libusb_free_device_list(list[0].list, C.int(1))
}

const maxPortDepth = 8

func (dev *Device) GetPortNumbers() ([]byte, error) {
	ports := [maxPortDepth]byte{}
	rc := int(C.libusb_get_port_numbers(dev.ptr, (*C.uint8_t)(&ports[0]), (C.int)(len(ports))))
	if rc < 0 {
		return nil, Error(rc)
	}
	return ports[:rc], nil
}
func (dev *Device) GetParent() *Device {
	return newDevice(nil, C.libusb_get_parent(dev.ptr))
}
func (dev *Device) GetMaxPacketSize(endpoint uint8) int {
	return int(C.libusb_get_max_packet_size(dev.ptr, (C.uchar)(endpoint)))
}
func (dev *Device) MatchVidPid(vendor_id, product_id uint16) bool {
	return dev.IDVender == vendor_id && dev.IDProduct == product_id
}
func (dev *Device) String() string {
	return fmt.Sprintf("Bus=%d, Port=%d, Addr=%d, Pid:Vid=%04x:%04x", dev.Bus, dev.Port, dev.Address, dev.IDProduct, dev.IDVender)
}

type Handle struct {
	dev *Device
	ptr *C.struct_libusb_device_handle

	timeout uint
}

func (dev *Device) Open() (*Handle, error) {
	var h Handle
	rc := int(C.libusb_open(dev.ptr, (**C.struct_libusb_device_handle)(&h.ptr)))
	if rc < 0 {
		return nil, Error(rc)
	}
	h.dev = dev
	return &h, nil
}
func OpenDeviceWithPidVid(vendor_id, product_id uint16) (*Handle, error) {
	ctx := Init()
	ptr := C.libusb_open_device_with_vid_pid(ctx.handle, (C.uint16_t)(vendor_id), (C.uint16_t)(product_id))
	if ptr == nil {
		return nil, ErrNoDevice
	}
	return &Handle{
		dev: newDevice(nil, C.libusb_get_device(ptr)),
		ptr: ptr,
	}, nil
}
func (h *Handle) Close() {
	C.libusb_close(h.ptr)
}
func (h *Handle) GetDevice() *Device {
	return h.dev
}
func (h *Handle) ClaimInterface(interface_number int) error {
	rc := int(C.libusb_claim_interface(h.ptr, (C.int)(interface_number)))
	if rc < 0 {
		return Error(rc)
	}
	return nil
}
func (h *Handle) ReleaseInterface(interface_number int) error {
	rc := int(C.libusb_release_interface(h.ptr, (C.int)(interface_number)))
	if rc < 0 {
		return Error(rc)
	}
	return nil
}
func (h *Handle) ResetDevice() error {
	rc := int(C.libusb_reset_device(h.ptr))
	if rc < 0 {
		return Error(rc)
	}
	return nil
}
func (h *Handle) KernelDriverActive(interface_number int) (bool, error) {
	rc := int(C.libusb_kernel_driver_active(h.ptr, (C.int)(interface_number)))
	if rc < 0 {
		return false, Error(rc)
	}
	return rc != 0, nil
}
func (h *Handle) DetachKernelDriver(interface_number int) error {
	rc := int(C.libusb_detach_kernel_driver(h.ptr, (C.int)(interface_number)))
	if rc < 0 {
		return Error(rc)
	}
	return nil
}
func (h *Handle) AttachKernelDriver(interface_number int) error {
	rc := int(C.libusb_attach_kernel_driver(h.ptr, (C.int)(interface_number)))
	if rc < 0 {
		return Error(rc)
	}
	return nil
}
func (h *Handle) SetAutoDetachKernelDriver(enable bool) error {
	enable_int := 0
	if enable {
		enable_int = 1
	}
	rc := int(C.libusb_set_auto_detach_kernel_driver(h.ptr, (C.int)(enable_int)))
	if rc < 0 {
		return Error(rc)
	}
	return nil
}

func (h *Handle) SetTimeout(timeout uint) {
	h.timeout = timeout
}

func (h *Handle) ControlTransferTimeout(typ RequestType, req uint8, value, index uint16, p []byte, timeout uint) (n int, err error) {
	var data_ptr *C.uchar
	if len(p) > 0 {
		data_ptr = (*C.uchar)(&p[0])
	}
	rc := int(C.libusb_control_transfer(h.ptr,
		C.uint8_t(typ),
		C.uint8_t(req),
		C.uint16_t(value),
		C.uint16_t(index),
		data_ptr,
		C.uint16_t(uint16(len(p))),
		C.uint(timeout)))
	if rc < 0 {
		return 0, Error(rc)
	}
	return rc, nil
}
func (h *Handle) ControlTransfer(typ RequestType, req uint8, value, index uint16, p []byte) (n int, err error) {
	return h.ControlTransferTimeout(typ, req, value, index, p, h.timeout)
}
func (h *Handle) ControlRead(req uint8, value, index uint16, p []byte) (n int, err error) {
	return h.ControlTransfer(EndpointIn|RequestTypeVendor|RecipientDevice, req, value, index, p)
}
func (h *Handle) ControlWrite(req uint8, value, index uint16, p []byte) (n int, err error) {
	return h.ControlTransfer(EndpointOut|RequestTypeVendor|RecipientDevice, req, value, index, p)
}
func (h *Handle) Command(req uint8, value, index uint16) error {
	_, err := h.ControlWrite(req, value, index, nil)
	return err
}

func (h *Handle) bulkTransferTimeout(ep uint8, p []byte, timeout uint) (n int, err error) {
	var transferred C.int
	rc := int(C.libusb_bulk_transfer(h.ptr,
		C.uchar(ep),
		(*C.uchar)(&p[0]),
		C.int(len(p)),
		&transferred,
		C.uint(timeout)))
	if rc != 0 {
		return 0, Error(rc)
	}
	return int(transferred), nil
}
func (h *Handle) BulkRead(ep uint8, p []byte) (n int, err error) {
	return h.bulkTransferTimeout(ep&0x07|uint8(EndpointIn), p, h.timeout)
}
func (h *Handle) BulkWrite(ep uint8, p []byte) (n int, err error) {
	return h.bulkTransferTimeout(ep&0x07|uint8(EndpointOut), p, h.timeout)
}

func (h *Handle) interruptTransferTimeout(ep uint8, p []byte, timeout uint) (n int, err error) {
	var transferred C.int
	rc := int(C.libusb_interrupt_transfer(h.ptr,
		C.uchar(ep),
		(*C.uchar)(&p[0]),
		C.int(len(p)),
		&transferred,
		C.uint(timeout)))
	if rc != 0 {
		return 0, Error(rc)
	}
	return int(transferred), nil
}
func (h *Handle) InterruptRead(ep uint8, p []byte) (n int, err error) {
	return h.interruptTransferTimeout(ep&0x07|uint8(EndpointIn), p, h.timeout)
}
func (h *Handle) InterruptWrite(ep uint8, p []byte) (n int, err error) {
	return h.interruptTransferTimeout(ep&0x07|uint8(EndpointOut), p, h.timeout)
}

type BulkTransfer struct {
	h           *Handle
	epIn, epOut uint8
	timeout     uint
	flag        int
}

const (
	bulkTransferCanRead = 1 << iota
	bulkTransferCanWrite
)

func (bt *BulkTransfer) SetTimeout(timeout uint) {
	bt.timeout = timeout
}
func (bt *BulkTransfer) Read(p []byte) (n int, err error) {
	if bt.flag&bulkTransferCanRead == 0 {
		return 0, errors.New("bulk transfer: cannot read")
	}
	return bt.h.bulkTransferTimeout(bt.epIn&0x07|uint8(EndpointIn), p, bt.timeout)
}
func (bt *BulkTransfer) Write(p []byte) (n int, err error) {
	if bt.flag&bulkTransferCanWrite == 0 {
		return 0, errors.New("bulk transfer: cannot write")
	}
	return bt.h.bulkTransferTimeout(bt.epOut&0x07|uint8(EndpointOut), p, bt.timeout)
}
func (h *Handle) GetBulkTransfer(epIn uint8, epOut uint8) *BulkTransfer {
	return &BulkTransfer{
		h:       h,
		epIn:    epIn,
		epOut:   epOut,
		timeout: h.timeout,
		flag:    bulkTransferCanRead | bulkTransferCanWrite,
	}
}
func (h *Handle) GetBulkReader(ep uint8) io.Reader {
	return &BulkTransfer{
		h:       h,
		epIn:    ep,
		timeout: h.timeout,
		flag:    bulkTransferCanRead,
	}
}
func (h *Handle) GetBulkWriter(ep uint8) io.Writer {
	return &BulkTransfer{
		h:       h,
		epOut:   ep,
		timeout: h.timeout,
		flag:    bulkTransferCanWrite,
	}
}

type InterruptTransfer struct {
	h       *Handle
	ep      uint8
	timeout uint
}

func (it *InterruptTransfer) SetTimeout(timeout uint) {
	it.timeout = timeout
}
func (it *InterruptTransfer) Read(p []byte) (n int, err error) {
	return it.h.interruptTransferTimeout(it.ep&0x07|uint8(EndpointIn), p, it.timeout)
}
func (it *InterruptTransfer) Write(p []byte) (n int, err error) {
	return it.h.interruptTransferTimeout(it.ep&0x07|uint8(EndpointOut), p, it.timeout)
}
func (h *Handle) GetInterruptTransfer(ep uint8) *InterruptTransfer {
	return &InterruptTransfer{
		h:       h,
		ep:      ep,
		timeout: h.timeout,
	}
}
func (h *Handle) GetInterruptReader(ep uint8) io.Reader {
	return h.GetInterruptTransfer(ep)
}
func (h *Handle) GetInterruptWriter(ep uint8) io.Writer {
	return h.GetInterruptTransfer(ep)
}

func (h *Handle) GetDescriptorBuffer(desc_type, desc_index uint8, data []byte) ([]byte, error) {
	rc := int(C.libusb_get_descriptor(h.ptr,
		C.uint8_t(desc_type),
		C.uint8_t(desc_index),
		(*C.uchar)(&data[0]),
		C.int(len(data))))
	if rc < 0 {
		return nil, Error(rc)
	}
	return data[:rc], nil
}
func (h *Handle) GetDescriptor(desc_type, desc_index uint8) ([]byte, error) {
	buf := make([]byte, 1024)
	return h.GetDescriptorBuffer(desc_type, desc_index, buf)
}

func (h *Handle) GetStringDescriptor(desc_index uint8, langid uint16) (string, error) {
	if desc_index == 0 {
		return "", nil
	}
	p := make([]byte, 256)
	rc := int(C.libusb_get_string_descriptor(h.ptr,
		C.uint8_t(desc_index),
		C.uint16_t(langid),
		(*C.uchar)(&p[0]),
		C.int(len(p))))
	if rc < 0 {
		return "", Error(rc)
	}
	if rc < 4 {
		return "", nil
	}
	u16s := make([]uint16, 1)
	u8buf, n := make([]byte, rc*2), 0
	for i := 2; i < rc; i += 2 {
		u16s[0] = uint16(p[i]) + (uint16(p[i+1]) << 8)
		r := utf16.Decode(u16s)
		n += utf8.EncodeRune(u8buf[n:], r[0])
	}
	return string(u8buf[0:n]), nil
}
func (h *Handle) GetManufacturerString() (string, error) {
	return h.GetStringDescriptor(h.dev.IdxManufacturer, 0)
}
func (h *Handle) GetProductString() (string, error) {
	return h.GetStringDescriptor(h.dev.IdxProduct, 0)
}
func (h *Handle) GetSerialNumberString() (string, error) {
	return h.GetStringDescriptor(h.dev.IdxSerialNumber, 0)
}
