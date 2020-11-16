package service

import (
	"bufio"
	"fmt"
	"github.com/vskit-tv/vlog/log"
	"go-app/app"
	"go-app/util"
	"net"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

type Request interface {
	GetConn() net.Conn
	GetId() string
	Do() *Response
}

type BadRequest struct {
	Id   string
	Conn net.Conn
}

func (req *BadRequest) Do() *Response {
	return &Response{
		Code:    StatusBadRequest,
		Headers: make(map[string]string),
	}
}

func (req *BadRequest) GetConn() net.Conn {
	return req.Conn
}

func (req *BadRequest) GetId() string {
	return req.Id
}

type RequestTimeout struct {
	Id   string
	Conn net.Conn
}

func (req *RequestTimeout) Do() *Response {
	return &Response{
		Code:    StatusRequestTimeout,
		Headers: make(map[string]string),
	}
}

func (req *RequestTimeout) GetId() string {
	return req.Id
}

func (req *RequestTimeout) GetConn() net.Conn {
	return req.Conn
}

type RequestOk struct {
	Id            string
	Conn          net.Conn
	Body          []byte
	Method        string
	Path          string
	Headers       map[string]string
	ContentLength int
}

func (req *RequestOk) GetConn() net.Conn {
	return req.Conn
}

func (req *RequestOk) GetId() string {
	return req.Id
}

func (req *RequestOk) Do() *Response {
	if strings.Contains(req.Path, "/vshow/file/upload") {
		filename := req.getFilename()
		filepath := path.Join(app.GetDataPath(), req.getSessionID())
		isFinished, rangeInfo, err := req.saveFile(filepath)
		if err != nil {
			log.Errorf("save file error %+v", err)
			return &Response{
				Code:    StatusInternalServerError,
				Headers: map[string]string{},
			}
		}

		if !isFinished {
			return &Response{
				Code: StatusCreated,
				Headers: map[string]string{
					"Range": rangeInfo,
				},
				Body: rangeInfo,
			}
		}

		return ForwardRequest(
			Method(MethodPost),
			Path(req.Path),
			Headers(map[string]string{
				CommonHeaderTag: req.Headers[CommonHeaderTag],
			}),
			FormData(map[string]string{
				".name": filename,
				".path": filepath,
			}),
		)
	}

	return ForwardRequest(
		Method(req.Method),
		Path(req.Path),
		Headers(req.Headers),
		Body(req.Body),
	)
}

// 视频上传处理,返回是否完成信息
func (req *RequestOk) saveFile(filepath string) (bool, string, error) {
	contentRange := req.getContentRange()
	start, end, length, err := getContentRange(contentRange)
	if err != nil {
		return false, "", err
	}
	records, err := getUploadRecord(filepath)
	if err != nil {
		return false, "", err
	}
	records = append(records, util.Interval{start, end})
	records = util.Merge(records)

	err = saveFile(filepath, req.Body, start, end)
	if err != nil {
		return false, "", err
	}

	err = saveUploadRecord(filepath, records, length)
	if err != nil {
		return false, "", err
	}

	isFinished, rangeInfo := isUploadFinished(records, length)
	return isFinished, rangeInfo, nil
}

func (req *RequestOk) getSessionID() string {
	return req.Headers["session-id"]
}

func (req *RequestOk) getFilename() string {
	re := regexp.MustCompile(`.+filename="(.+)"`)
	m := re.FindStringSubmatch(req.Headers["content-disposition"])
	return m[1]
}

func (req *RequestOk) getContentRange() string {
	return req.Headers["x-content-range"]
}

func getUploadRecord(filename string) ([]util.Interval, error) {
	fstate, err := os.OpenFile(filename+".state", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	defer fstate.Close()
	records := []util.Interval{}
	scanner := bufio.NewScanner(fstate)
	for scanner.Scan() {
		start, end, _, err := getContentRange(scanner.Text())
		if err != nil {
			return nil, err
		}
		records = append(records, util.Interval{start, end})
	}
	return records, nil
}

func saveUploadRecord(filename string, records []util.Interval, length int) error {
	fstate, err := os.OpenFile(filename+".state", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer fstate.Close()
	l := len(records) - 1
	for i, record := range records {
		content := fmt.Sprintf("%d-%d/%d", record.Start, record.End, length)
		if i < l {
			content += "\n"
		}
		_, err = fstate.WriteString(content)
		if err != nil {
			return err
		}
	}

	return nil
}

func saveFile(filename string, data []byte, start int, end int) error {
	fdata, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer fdata.Close()
	fdata.Seek(int64(start), 0)
	_, err = fdata.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func getContentRange(rangeMessage string) (start int, end int, length int, err error) {
	reg := regexp.MustCompile(`\D*(\d+)\-(\d+)\/(\d+)`)
	matches := reg.FindStringSubmatch(rangeMessage)
	if start, err = strconv.Atoi(matches[1]); err != nil {
		return
	}
	if end, err = strconv.Atoi(matches[2]); err != nil {
		return
	}
	if length, err = strconv.Atoi(matches[3]); err != nil {
		return
	}
	return
}

func isUploadFinished(records []util.Interval, length int) (bool, string) {
	var ranges []string
	for _, record := range records {
		ranges = append(ranges, fmt.Sprintf("%d-%d/%d", record.Start, record.End, length))
	}
	if len(records) != 1 {
		return false, strings.Join(ranges, ",")
	}
	if records[0].Start != 0 || records[0].End != length-1 {
		return false, strings.Join(ranges, ",")
	}
	return true, ""
}
