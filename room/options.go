package room

import (
	"github.com/liangdas/mqant/log"
	"reflect"
	"time"
)

type UpdateHandle func(ds time.Duration)

type NoFoundHandle func(msg *QueueMsg) (reflect.Value, error)

type ErrorHandle func(msg *QueueMsg, err error)

type RecoverHandle func(msg *QueueMsg, err error)

/**
当房间关闭是通知持有方注销
*/
type LifeCallback func(table BaseTable) error

/*
获取可以路由到该房间的地址路径
*/
type Route func(TableId string) string

func newOptions(opts ...Option) Options {
	opt := Options{
		TimeOut:          60,
		Capaciity:        256,
		SendMsgCapaciity: 256,
		RunInterval:100*time.Millisecond,
	}

	for _, o := range opts {
		o(&opt)
	}
	return opt
}

type Option func(*Options)

type Options struct {
	Update           UpdateHandle
	NoFound          NoFoundHandle
	ErrorHandle      ErrorHandle
	RecoverHandle    RecoverHandle
	DestroyCallbacks LifeCallback
	TableId          string
	Router           Route
	Trace            log.TraceSpan
	TimeOut          int64  //判断客户端超时时间单位秒
	Capaciity        uint32 //消息队列容量,真实容量为 Capaciity*2
	SendMsgCapaciity uint32 //每帧发送消息容量
	RunInterval		 time.Duration //运行间隔
}

func Update(fn UpdateHandle) Option {
	return func(o *Options) {
		o.Update = fn
	}
}

func NoFound(fn NoFoundHandle) Option {
	return func(o *Options) {
		o.NoFound = fn
	}
}

func SetErrorHandle(fn ErrorHandle) Option {
	return func(o *Options) {
		o.ErrorHandle = fn
	}
}

func SetRecoverHandle(fn RecoverHandle) Option {
	return func(o *Options) {
		o.RecoverHandle = fn
	}
}

func TableId(v string) Option {
	return func(o *Options) {
		o.TableId = v
	}
}

func Router(v Route) Option {
	return func(o *Options) {
		o.Router = v
	}
}

func Trace(v log.TraceSpan) Option {
	return func(o *Options) {
		o.Trace = v
	}
}

func DestroyCallbacks(v LifeCallback) Option {
	return func(o *Options) {
		o.DestroyCallbacks = v
	}
}

func TimeOut(v int64) Option {
	return func(o *Options) {
		o.TimeOut = v
	}
}

func Capaciity(v uint32) Option {
	return func(o *Options) {
		o.Capaciity = v
	}
}

func SendMsgCapaciity(v uint32) Option {
	return func(o *Options) {
		o.SendMsgCapaciity = v
	}
}


func RunInterval(v time.Duration) Option {
	return func(o *Options) {
		o.RunInterval = v
	}
}