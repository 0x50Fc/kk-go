package kk

import (
	"bytes"
	"strconv"
)

type Message struct {
	Method  string
	From    string
	To      string
	Type    string
	Content []byte
}

type IReader interface {
	Read(data []byte) (int, error)
}

type IWriter interface {
	Write(data []byte) (int, error)
}

const MessageReaderStateKey = 0
const MessageReaderStateValue = 1
const MessageReaderStateContent = 2

type MessageReader struct {
	_state   int
	_key     *bytes.Buffer
	_value   *bytes.Buffer
	_length  int
	_content *bytes.Buffer
	_message Message
	_unread  []byte
	_data    []byte
}

func NewMessageReader() *MessageReader {
	var v = MessageReader{}
	v._state = MessageReaderStateKey
	v._key = bytes.NewBuffer(nil)
	v._value = bytes.NewBuffer(nil)
	v._length = 0
	v._content = bytes.NewBuffer(nil)
	v._unread = nil
	v._data = make([]byte, 2048)
	return &v
}

func (rd *MessageReader) readBytes(data []byte, n int) (*Message, int) {

	var i = 0

	for i < n {
		var c = data[i]
		switch rd._state {
		case MessageReaderStateKey:
			{
				if c == ':' {
					rd._state = MessageReaderStateValue
					rd._value.Reset()
				} else if c == '\n' {
					if rd._length == 0 {
						rd._state = MessageReaderStateKey
						rd._message.Content = nil
						return &rd._message, i + 1
					} else {
						rd._state = MessageReaderStateContent
						rd._content.Reset()
					}
				} else {
					rd._key.WriteByte(c)
				}
			}
		case MessageReaderStateValue:
			{
				if c == '\n' {
					var key, _ = rd._key.ReadString(0)
					var value, _ = rd._value.ReadString(0)
					switch key {
					case "METHOD":
						rd._message.Method = value
					case "FROM":
						rd._message.From = value
					case "TO":
						rd._message.To = value
					case "TYPE":
						rd._message.Type = value
					case "LENGTH":
						rd._length, _ = strconv.Atoi(value)
					}
					rd._key.Reset()
					rd._value.Reset()
				} else {
					rd._value.WriteByte(c)
				}
			}
		case MessageReaderStateContent:
			{
				rd._content.WriteByte(c)
				if rd._length == rd._content.Len() {
					rd._state = MessageReaderStateKey
					rd._message.Content = rd._content.Bytes()
					rd._length = 0
					rd._content.Reset()
					return &rd._message, i + 1
				}
			}
		}
		i++
	}

	return nil, i
}

func (rd *MessageReader) Read(reader IReader) (*Message, error) {

	if rd._unread != nil {
		var n = len(rd._unread)
		var v, i = rd.readBytes(rd._unread, n)
		if i == n {
			rd._unread = nil
		} else {
			rd._unread = rd._unread[i:]
		}
		if v != nil {
			return v, nil
		}
	}

	var n, err = reader.Read(rd._data)

	if n > 0 {
		var v, i = rd.readBytes(rd._data, n)
		if i < n {
			rd._unread = rd._data[i:]
		}
		if v != nil {
			return v, err
		}
	}

	return nil, err
}

type MessageWriter struct {
	_data *bytes.Buffer
}

func NewMessageWriter() *MessageWriter {
	var v = MessageWriter{}
	v._data = bytes.NewBuffer(nil)
	return &v
}

func (wd *MessageWriter) Done(writer IWriter) (bool, error) {

	if wd._data.Len() != 0 {
		var n, err = writer.Write(wd._data.Bytes())
		wd._data.Truncate(wd._data.Len() - n)
		return wd._data.Len() == 0, err
	}

	return true, nil
}

func (wd *MessageWriter) Write(writer IWriter, message *Message) (bool, error) {

	wd._data.WriteString("METHOD:")
	wd._data.WriteString(message.Method)
	wd._data.WriteByte('\n')

	wd._data.WriteString("FROM:")
	wd._data.WriteString(message.From)
	wd._data.WriteByte('\n')

	wd._data.WriteString("TO:")
	wd._data.WriteString(message.To)
	wd._data.WriteByte('\n')

	wd._data.WriteString("TYPE:")
	wd._data.WriteString(message.Type)
	wd._data.WriteByte('\n')

	wd._data.WriteString("LENGTH:")

	if message.Content == nil || len(message.Content) == 0 {
		wd._data.WriteString("0\n\n")
	} else {
		wd._data.WriteString(strconv.Itoa(len(message.Content)))
		wd._data.WriteString("\n\n")
		wd._data.Write(message.Content)
	}

	return wd.Done(writer)
}
