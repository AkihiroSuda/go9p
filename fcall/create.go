package fcall

import "fmt"

type TCreate struct {
	FCall
	Fid uint32
	Name string
	Perm uint32
	Mode uint8
}

func (create *TCreate) String() string {
	return fmt.Sprintf("tcreate: [%s, fid: %d, name: %s, perm: %d, mode: %d]",
		&create.FCall, create.Fid, create.Name, create.Perm, create.Mode)
}

func (create *TCreate) Parse(buff []byte) ([]byte, error) {
	buff, err := fcParse(&create.FCall, buff)
	if err != nil {
		return nil, err
	}

	create.Fid, buff = FromLittleE32(buff)
	create.Name, buff = FromString(buff)
	create.Perm, buff = FromLittleE32(buff)
	create.Mode = buff[0]
	buff = buff[1:]
	return buff, nil
}

func (create *TCreate) Compose() []byte {
	// size[4] Tcreate tag[2] fid[4] name[s] perm[4] mode[1]
	length := 4 + 1 + 2 + 4 + (2 + len(create.Name)) + 4 + 1
	buff := make([]byte, length)
	buffer := buff

	buffer = ToLittleE32(uint32(length), buffer)
	buffer[0] = create.Ctype; buffer = buffer[1:]
	buffer = ToLittleE16(create.Tag, buffer)
	buffer = ToLittleE32(create.Fid, buffer)
	buffer = ToString(create.Name, buffer)
	buffer = ToLittleE32(create.Perm, buffer)
	buffer[0] = create.Mode; buffer = buffer[1:]
	return buff
}

type RCreate struct {
	FCall
	Qid Qid
	Iounit uint32
}

func (create *RCreate) String() string {
	return fmt.Sprintf("rcreate: [%s, qid: [%s], iounit: %d]",
		&create.FCall, &create.Qid, create.Iounit)
}

func (create *RCreate) Parse(buff []byte) ([]byte, error) {
	buff, err := fcParse(&create.FCall, buff)
	if err != nil {
		return nil, err
	}

	buff, err = create.Qid.Parse(buff)
	if err != nil {
		return nil, err
	}

	create.Iounit, buff = FromLittleE32(buff)
	return buff, nil
}

func (create *RCreate) Compose() []byte {
	// size[4] Rcreate tag[2] qid[13] iounit[4]
	length := 4 + 1 + 2 + 13 + 4
	buff := make([]byte, length)
	buffer := buff

	buffer = ToLittleE32(uint32(length), buffer)
	buffer[0] = create.Ctype; buffer = buffer[1:]
	buffer = ToLittleE16(create.Tag, buffer)
	qidbuff := create.Qid.Compose()
	copy(buffer, qidbuff)
	buffer = buffer[len(qidbuff):]
	buffer = ToLittleE32(create.Iounit, buffer)
	return buff
}
