package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// fheader define for parse cmd into flag
// support such as:
//  --fheader one --fheader two
type fheader []*header

func (this *fheader) String() string {
	ret := []string{}
	for _, h := range *this {
		ret = append(ret, h.key+":"+h.value)
	}
	return strings.Join(ret, "\t")
}

func (this *fheader) Set(value string) error {
	if strings.Contains(value, ":") {
		arr := strings.SplitN(value, ":", 2)
		*this = append(*this, &header{arr[0], arr[1]})
	} else {
		*this = append(*this, &header{value, ""})
	}
	return nil
}

type header struct {
	key   string
	value string
}

// ConfigWithArgs Read settings from cmd
func (this *Stress) ConfigWithArgs() {
	flag.StringVar(&this.Url, "url", "http://localhost/",
		"request url\n",
	)
	flag.StringVar(&this.Method, "method", "GET",
		"request http method\n",
	)
	flag.StringVar(&this.Body, "body", "",
		"request body",
	)
	flag.IntVar(&this.Timeout, "timeout", 10000,
		"request timeout limit millisecond\n",
	)
	flag.Var(&this.Header, "header",
		"request header\n"+
			"multiple args will be appended to header\n"+
			"Example: \n\t-header 'Content-Type: application/json' -header 'Authorization: Basic xxx'",
	)
	flag.IntVar(&this.Status, "status", 200,
		"response http status\n"+
			"assign failed if different with the response http status\n",
	)
	flag.Float64Var(&this.MaxFailedRatio, "max-failed-ratio", 0,
		"the max response failed ratio allow 0-1 float",
	)
	flag.StringVar(&this.Response, "response", "",
		"expect response body\n"+
			"assign failed if different with the response body\n"+
			"always success if empty",
	)
	flag.IntVar(&this.Thread, "thread", 1,
		"thread number, must >= 1\n"+
			"dynamic mode start from this threads number\n",
	)
	flag.IntVar(&this.Interval, "interval", 0,
		"request interval milliseconds, must >= 0\n"+
			"worker will sleep interval times after the last request finished",
	)

	flag.StringVar(&this.Mode, "mode", MODE_DYNAMIC,
		"working mode:\n"+
			"\tdynamic or static\n"+
			"- dynamic: threads number will increase with <step> after <stay> time\n"+
			"- static: always working with initial threads number\n",
	)
	flag.IntVar(&this.Step, "step", 1,
		"for dynamic mode, threads increase step\n",
	)
	flag.IntVar(&this.Stay, "stay", 1,
		"for dynamic mode, stay seconds in each step\n",
	)

	flag.IntVar(&this.LogInterval, "log-interval", 1,
		"log output inverval second\n",
	)
	flag.StringVar(&this.LogPath, "log-path", "./",
		"log output path\n",
	)

	flag.StringVar(&this.configFilePath, "config", "",
		"* not support now\n"+
			"configure file path, override args settings on conflict",
	)

	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"Usage of %s: %s [options]\n"+
				"\n"+
				"\tExample: %s -url 'http://localhost/hello'\n"+
				"\n"+
				"Option list:\n",
			os.Args[0],
			os.Args[0],
			os.Args[0],
		)
		flag.PrintDefaults()
	}

	flag.Parse()
	if this.configFilePath == "" {
		this.configCheck()
	} else {
		this.ConfigWithFile(this.configFilePath)
	}
	this.configParse()
}

// ConfigWithFile read settings from a json file
func (this *Stress) ConfigWithFile(filepath string) {
	// TODO read json file and bind
	this.configCheck()
}

// parse some settings
func (this *Stress) configCheck() {
	if this.Url == "" {
		flag.Usage()
		os.Exit(2)
	}
}

func (this *Stress) configParse() {
	var err error
	if this.logInfoHandle, err = os.OpenFile(this.LogPath+"info.log", os.O_RDWR|os.O_CREATE, 0666); err != nil {
		panic(err)
	}
	if this.logErrorHandle, err = os.OpenFile(this.LogPath+"error.log", os.O_RDWR|os.O_CREATE, 0666); err != nil {
		panic(err)
	}
}
