package message

import(
	"io"
	"encoding/binary"
)

func send(writer io.Writer, data interface{}) error {
	return binary.Write(writer, binary.BigEndian, data)
}

func SendPingResp(writer io.Writer) error {
	return send(writer, GetRespType(Types["ping"]))
}

func SendLoginResp(writer io.Writer, code string) error {
	err := send(writer, GetRespType(Types["login"]))
	if err != nil {
		return err
	}
	return send(writer, LoginResponseCode[code])
}
