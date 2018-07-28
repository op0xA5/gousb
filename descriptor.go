package gousb

import (
	"encoding/binary"
	"errors"
)

var usbEncoding = binary.LittleEndian

type Descriptor interface {
	Len() int
	Type() uint8
}

type DeviceDescriptor struct {
	Length          uint8
	DescriptorType  uint8
	BcdUSB          uint16
	DeviceClass     uint8
	DeviceSubClass  uint8
	DeviceProtol    uint8
	MaxPacketSize0  uint8
	IDVender        uint16
	IDProduct       uint16
	BcdDevice       uint16
	IdxManufacturer uint8
	IdxProduct      uint8
	IdxSerialNumber uint8
	NumConfiguation uint8
}

func (desc *DeviceDescriptor) Len() int {
	return int(desc.Length)
}
func (desc *DeviceDescriptor) Type() uint8 {
	return desc.DescriptorType
}

const deviceDescriptorLength = 18

func (desc *DeviceDescriptor) Put(b []byte) {
	desc.Length = b[0]
	desc.DescriptorType = b[1]
	desc.BcdUSB = usbEncoding.Uint16(b[2:])
	desc.DeviceClass = b[4]
	desc.DeviceSubClass = b[5]
	desc.DeviceProtol = b[6]
	desc.MaxPacketSize0 = b[7]
	desc.IDVender = usbEncoding.Uint16(b[8:])
	desc.IDProduct = usbEncoding.Uint16(b[10:])
	desc.BcdDevice = usbEncoding.Uint16(b[12:])
	desc.IdxManufacturer = b[14]
	desc.IdxProduct = b[15]
	desc.IdxSerialNumber = b[16]
	desc.NumConfiguation = b[17]
}

type ConfigurationDescriptor struct {
	Length             uint8
	DescriptorType     uint8
	TotalLength        uint16
	NumInterfaces      uint8
	ConfigurationValue uint8
	IdxConfiguration   uint8
	Attributes         uint8
	MaxPower           uint8
}

func (desc *ConfigurationDescriptor) Len() int {
	return int(desc.Length)
}
func (desc *ConfigurationDescriptor) Type() uint8 {
	return desc.DescriptorType
}

const configurationDescriptorLength = 9

func (desc *ConfigurationDescriptor) Put(b []byte) {
	desc.Length = b[0]
	desc.DescriptorType = b[1]
	desc.TotalLength = usbEncoding.Uint16(b[2:])
	desc.NumInterfaces = b[4]
	desc.ConfigurationValue = b[5]
	desc.IdxConfiguration = b[6]
	desc.Attributes = b[7]
	desc.MaxPower = b[8]
}

type InterfaceDescriptor struct {
	Length            uint8
	DescriptorType    uint8
	InterfaceNumber   uint8
	AlternateSetting  uint8
	NumEndpoint       uint8
	InterfaceClass    uint8
	InterfaceSubClass uint8
	InterfaceProtocol uint8
	IdxInterface      uint8
}

func (desc *InterfaceDescriptor) Len() int {
	return int(desc.Length)
}
func (desc *InterfaceDescriptor) Type() uint8 {
	return desc.DescriptorType
}

const interfaceDescriptorLength = 9

func (desc *InterfaceDescriptor) Put(b []byte) {
	desc.Length = b[0]
	desc.DescriptorType = b[1]
	desc.InterfaceNumber = b[2]
	desc.AlternateSetting = b[3]
	desc.NumEndpoint = b[4]
	desc.InterfaceClass = b[5]
	desc.InterfaceSubClass = b[6]
	desc.InterfaceProtocol = b[7]
	desc.IdxInterface = b[8]
}

type EndpointDescriptor struct {
	Length          uint8
	DescriptorType  uint8
	EndpointAddress uint8
	Attributes      uint8
	MaxPacketSize   uint16
	Interval        uint8
}

func (desc *EndpointDescriptor) Len() int {
	return int(desc.Length)
}
func (desc *EndpointDescriptor) Type() uint8 {
	return desc.DescriptorType
}

const endpointDescriptorLength = 7

func (desc *EndpointDescriptor) Put(b []byte) {
	desc.Length = b[0]
	desc.DescriptorType = b[1]
	desc.EndpointAddress = b[2]
	desc.Attributes = b[3]
	desc.MaxPacketSize = usbEncoding.Uint16(b[4:])
	desc.Interval = b[6]
}

func (desc *EndpointDescriptor) InOut() RequestType {
	return RequestType(desc.EndpointAddress & 0x80)
}
func (desc *EndpointDescriptor) Ep() uint8 {
	return desc.EndpointAddress & 0x0F
}
func (desc *EndpointDescriptor) TransferType() TransferType {
	return TransferType(desc.Attributes & 0x03)
}

type StringDescriptor struct {
	Length         uint8
	DescriptorType uint8
	String         string
}

func (desc *StringDescriptor) Len() int {
	return int(desc.Length)
}
func (desc *StringDescriptor) Type() uint8 {
	return desc.DescriptorType
}

func (desc *StringDescriptor) Put(b []byte) {
	desc.Length = b[0]
	desc.DescriptorType = b[1]
	desc.String = string(b[2:desc.Length])
}

func ParseDescriptor(b []byte) (Descriptor, error) {
	if len(b) < 2 {
		return nil, errors.New("too less bytes")
	}
	length, typ := b[0], uint8(b[1])
	if len(b) < int(length) {
		return nil, errors.New("no enough data")
	}
	switch typ {
	case DescriptorTypeDevice:
		if length != deviceDescriptorLength {
			return nil, errors.New("descriptor length mismatch")
		}
		desc := new(DeviceDescriptor)
		desc.Put(b)
		return desc, nil
	case DescriptorTypeConfig:
		if length != configurationDescriptorLength {
			return nil, errors.New("descriptor length mismatch")
		}
		desc := new(ConfigurationDescriptor)
		desc.Put(b)
		return desc, nil
	case DescriptorTypeInterface:
		if length != interfaceDescriptorLength {
			return nil, errors.New("descriptor length mismatch")
		}
		desc := new(InterfaceDescriptor)
		desc.Put(b)
		return desc, nil
	case DescriptorTypeEndpoint:
		if length != endpointDescriptorLength {
			return nil, errors.New("descriptor length mismatch")
		}
		desc := new(EndpointDescriptor)
		desc.Put(b)
		return desc, nil
	case DescriptorTypeString:
		desc := new(StringDescriptor)
		desc.Put(b)
		return desc, nil
	}
	return nil, errors.New("unknown descriptor")
}
