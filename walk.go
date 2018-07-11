package go9p

import (
	"fmt"
)

type TWalk struct {
	FCall
	Fid    uint32
	Newfid uint32
	Nwname uint16
	Wname  []string
}

func (walk *TWalk) String() string {
	ret := fmt.Sprintf("twalk: [%s, fid: %d, newfid: %d, nwname: %d, wname: <",
		&walk.FCall, walk.Fid, walk.Newfid, walk.Nwname)
	for _, s := range walk.Wname {
		ret += s + ", "
	}
	ret += ">]"
	return ret
}

func (walk *TWalk) Parse(buff []byte) ([]byte, error) {
	buff, err := fcParse(&walk.FCall, buff)
	if err != nil {
		return nil, err
	}

	walk.Fid, buff = fromLittleE32(buff)
	walk.Newfid, buff = fromLittleE32(buff)
	walk.Nwname, buff = fromLittleE16(buff)
	walk.Wname = make([]string, walk.Nwname)
	var i uint16
	for ; i < walk.Nwname; i++ {
		walk.Wname[i], buff = fromString(buff)
	}
	return buff, nil
}

func (walk *TWalk) Compose() []byte {
	// size[4] Twalk  tag[2] fid[4] newfid[4] nwname[2] nwname*(wname[s])
	length := 4 + 1 + 2 + 4 + 4 + 2
	for _, name := range walk.Wname {
		length += 2 + len(name)
	}
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = walk.Ctype
	buffer = buffer[1:]
	buffer = toLittleE16(walk.Tag, buffer)
	buffer = toLittleE32(walk.Fid, buffer)
	buffer = toLittleE32(walk.Newfid, buffer)
	buffer = toLittleE16(walk.Nwname, buffer)
	for _, name := range walk.Wname {
		buffer = toString(name, buffer)
	}

	return buff
}

func (walk *TWalk) Reply(fs *filesystem, conn *connection, s *Server) IFCall {
	file := fs.fileForPath(conn.pathForFid(walk.Fid))
	if file == nil {
		return &RWalk{FCall{Rwalk, walk.Tag}, 0, nil}
		//return &RError{FCall{walk.Ctype, walk.Tag}, "No such file."}
	}

	// Special case for ..
	if walk.Nwname > 0 && walk.Wname[0] == ".." {

		if file.Parent != nil {
			conn.setFidPath(walk.Newfid, file.Parent.Path)
			qids := make([]Qid, 1)
			qids[0] = file.Parent.Stat.Qid
			return &RWalk{FCall{Rwalk, walk.Tag}, 1, qids}
		} else {
			return &RWalk{FCall{Rwalk, walk.Tag}, 0, nil}
		}
	}

	qids := make([]Qid, 0)
	path := file.Path
	for i := 0; i < int(walk.Nwname); i++ {
		if path == "/" {
			path += walk.Wname[i]
		} else {
			path += "/" + walk.Wname[i]
		}
		currfile := fs.fileForPath(path)
		if currfile == nil {
			return &RError{FCall{Rerror, walk.Tag}, "No such path."}
		}

		qids = append(qids, currfile.Stat.Qid)
	}
	conn.setFidPath(walk.Newfid, path)
	return &RWalk{FCall{Rwalk, walk.Tag}, uint16(len(qids)), qids}
}

type RWalk struct {
	FCall
	Nwqid uint16
	Wqid  []Qid
}

func (walk *RWalk) String() string {
	ret := fmt.Sprintf("rwalk: [%s, nwqid: %d, wqid: <",
		&walk.FCall, walk.Nwqid)
	for _, qid := range walk.Wqid {
		ret += fmt.Sprintf("<%s>, ", &qid)
	}
	ret += ">]"
	return ret
}

func (walk *RWalk) Parse(buff []byte) ([]byte, error) {
	buff, err := fcParse(&walk.FCall, buff)
	if err != nil {
		return nil, err
	}

	walk.Nwqid, buff = fromLittleE16(buff)
	walk.Wqid = make([]Qid, walk.Nwqid)
	var i uint16
	for ; i < walk.Nwqid; i++ {
		buff, err = walk.Wqid[i].Parse(buff)
		if err != nil {
			return nil, err
		}
	}
	return buff, nil
}

func (walk *RWalk) Compose() []byte {
	// size[4] Rwalk tag[2] nwqid[2] nwqid*(wqid[13])
	length := 4 + 1 + 2 + 2 + (walk.Nwqid * 13)
	buff := make([]byte, length)
	buffer := buff

	buffer = toLittleE32(uint32(length), buffer)
	buffer[0] = walk.Ctype
	buffer = buffer[1:]
	buffer = toLittleE16(walk.Tag, buffer)
	buffer = toLittleE16(walk.Nwqid, buffer)
	for _, qid := range walk.Wqid {
		qidbuff := qid.Compose()
		copy(buffer, qidbuff)
		buffer = buffer[len(qidbuff):]
	}

	return buff
}
