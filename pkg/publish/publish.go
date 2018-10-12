package publish

import (
	"time"
	"strconv"
	"os/exec"
	"fmt"
	"io"
	"bufio"
)

const (
	// Errors
	NO_CREDENTIAL	= "No credentials set"

	// Bitrate units
	KILOBYTE 		= "k"

	// FFMPEG options
	INPUT 			= "-i"
	OUTPUT		 	= "-f"
	NATIVE_FRAME 	= "-re"
	THREADS 		= "-threads"
	STREAM_LOOP 	= "-stream_loop"
	VIDEO_CODEC 	= "-codec:v"
	AUDIO_CODEC 	= "-codec:a"
	PROFILE 		= "-profile:v"
	VIDEO_BITRATE 	= "-b:v"
	AUIDO_BITRATE 	= "-b:a"
	MAX_BITRATE 	= "-maxrate"
	BUFSIZE 		= "-bufsize"

	// Codecs
	H264 			= "libx264"
	AAC 			= "aac"
	FLV 			= "flv"

	// Profiles
	BASELINE 		= "baseline"
	MAIN 			= "main"
)

type (
	option struct {
		nativFrame  bool   // -re 'Read input at native frame rate'
		threads     uint   // -threads
		looping     int    // -stream_loop
		videoCodec  string // -codec:v
		audioCodec  string // -codec:a
		profile     string // -profile
		vBitrate    uint   // -b:v
		vMaxBitrate uint   // -maxrate
		aBitrate    uint   // -b:a
		bufSize     uint   // -bufsize
	}

	Config struct {
		binary  string
		*option
	}

	client struct {
		uri         string
		port        uint
		application string
		stream      string
		id          string
		pw          string
		src         string
		duration    *time.Duration
		cmd         string
		exec        *exec.Cmd
		*Config
	}

	Builder interface {
		Build() Publisher
		SetDefaultConf() *client
		SetBinary(bin string) *client
		SetSource(uri string) *client
		SetUri(uri string) *client
		SetApplication(appName string) *client
		SetStream(streamName string) *client
		SetCredentials(id, pw string) *client
		Command() string
	}
)

func NewBuilder() Builder {
	return &client{
		port:1935,
		cmd:"",
		Config: &Config{},
	}
}

func DefaultConfig() *Config {
	return &Config{
		binary: "ffmpeg",
		option: defaultOption(),
	}
}

func defaultOption() *option {
	return &option{
		nativFrame:true,
		looping:0,
		videoCodec:H264,
		audioCodec:AAC,
		profile:BASELINE,
		vBitrate:1500,
		vMaxBitrate:2000,
		aBitrate:128,
		bufSize:1500,
	}
}

func (c *client) Build() Publisher {
	return c
}

func (c *client)SetDefaultConf() *client  {
	c.Config = DefaultConfig()
	return c
}

func (c *client) SetBinary(bin string) *client {
	c.Config.binary = bin
	return c
}

func (c *client)SetSource(uri string) *client {
	c.src = uri
	return c
}

func (c *client) SetUri(uri string) *client {
	c.uri = uri
	return c
}

func (c *client)SetApplication(appName string) *client {
	c.application = appName
	return c
}

func (c *client) SetStream(streamName string) *client {
	c.stream = streamName
	return c
}

func (c *client) SetCredentials(id, pw string) *client {
	c.id = id
	c.pw = pw
	return c
}

func (c *client) SetOption(opt string, val interface{}) *client {
	switch opt {
	case NATIVE_FRAME:
		c.option.nativFrame = true
	case THREADS:
		c.option.threads = val.(uint)
	case STREAM_LOOP:
		c.option.looping = val.(int)
	case VIDEO_CODEC:
		c.option.videoCodec = val.(string)
	case AUDIO_CODEC:
		c.option.audioCodec = val.(string)
	case VIDEO_BITRATE:
		c.option.vBitrate = val.(uint)
	case AUIDO_BITRATE:
		c.option.aBitrate = val.(uint)
	case PROFILE:
		c.option.profile = val.(string)
	case BUFSIZE:
		c.option.bufSize = val.(uint)
	case MAX_BITRATE:
		c.option.vMaxBitrate = val.(uint)
	}

	return c
}

func (c *client) Command() string {
	c.set(c.binary)
	if c.option.nativFrame {
		c.set(NATIVE_FRAME)
	}

	if c.option.threads == 0 {
		c.set(THREADS, "1")
	} else {
		c.set(THREADS, formatUint(c.option.threads))
	}

	if c.option.looping != 0 {
		c.set(STREAM_LOOP, strconv.Itoa(c.option.looping))
	}

	c.set(INPUT, c.src).
		set(VIDEO_CODEC, c.option.videoCodec).
		set(VIDEO_BITRATE, formatUint(c.option.vBitrate) + KILOBYTE).
		set(PROFILE, c.option.profile).
		set(MAX_BITRATE, formatUint(c.option.vMaxBitrate) + KILOBYTE).
		set(BUFSIZE, formatUint(c.option.bufSize) + KILOBYTE).
		set(AUDIO_CODEC, c.option.audioCodec).
		set(AUIDO_BITRATE, formatUint(c.option.aBitrate) + KILOBYTE).
		set(OUTPUT, FLV, c.fullUri())

	return c.cmd
}

func (c *client) fullUri() string {
	return fmt.Sprintf("rtmp://%s:%s@%s:%d/%s/%s",
		c.id, c.pw, c.uri, c.port, c.application, c.stream)
}

func (c *client) set(opt string, val ...string) *client {
	switch len(val) {
	case 0:
		c.cmd += fmt.Sprintf("%s ", opt)
	case 1:
		c.cmd += fmt.Sprintf("%s %s ", opt, val[0])
	default:
		c.cmd += opt
		for _, v := range val {
			c.cmd += " " + v
		}
		c.cmd += " "
	}

	return c
}

type Publisher interface {
	Initialize() Publisher
	Publish() error
	UnPublish() error
	PublishWithDuration(d time.Duration) error
}


var stdoutIn, stderrIn io.ReadCloser
func (c *client)Initialize() Publisher {
	c.exec = exec.Command("bash", "-c", c.Command())
	fmt.Println(c.cmd)
	stdoutIn, _ = c.exec.StdoutPipe()
	stderrIn, _ = c.exec.StderrPipe()

	return c
}

func (c *client) Publish() error {
	return c.PublishWithDuration(0)
}

func (c *client) UnPublish() error {
	err := c.exec.Process.Kill()

	return err
}

func (c *client) PublishWithDuration(d time.Duration) error {
	err := c.exec.Start()

	go func() {
		s := bufio.NewScanner(stdoutIn)
		for s.Scan() {
			fmt.Println(s.Text())
		}
	}()

	go func() {
		s := bufio.NewScanner(stderrIn)
		for s.Scan() {
			fmt.Println(s.Text())
		}
	}()

	if d != 0 {
		timer := time.AfterFunc(d, func() {
			err = c.UnPublish()
		})
		err = c.exec.Wait()
		timer.Stop()
	} else {
		err = c.exec.Wait()
	}

	if err.Error() == "signal: killed" {
		return nil
	}

	return err
}

func formatUint(i uint) string {
	return strconv.FormatUint(uint64(i), 10)
}